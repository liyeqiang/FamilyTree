package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"familytree/config"
	"familytree/handlers"
	"familytree/pkg/di"
	"familytree/pkg/workerpool"
	"familytree/repository"
	"familytree/services"

	"github.com/gorilla/mux"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("加载配置文件失败，使用默认配置: %v", err)
		cfg = config.DefaultConfig()
	}

	// 设置日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("🚀 启动家族树应用，端口: %s", cfg.Port)

	// 创建应用实例
	app, err := createApp(cfg)
	if err != nil {
		log.Fatalf("创建应用失败: %v", err)
	}
	defer app.cleanup()

	// 启动服务器
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      app.router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// 优雅关闭
	go func() {
		log.Printf("服务器启动在端口 %s", cfg.Port)
		log.Printf("📖 请访问 http://localhost:%s 查看管理界面", cfg.Port)
		log.Printf("📖 请访问 http://localhost:%s/docs 查看API文档", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
	}

	log.Println("服务器已关闭")
}

// App 应用实例
type App struct {
	router  *mux.Router
	db      *sql.DB
	cleanup func()
}

// createApp 创建应用实例
func createApp(cfg *config.Config) (*App, error) {
	var cleanupFuncs []func()

	cleanup := func() {
		for _, fn := range cleanupFuncs {
			fn()
		}
	}

	// 创建依赖注入容器
	container := di.NewContainer()

	// 创建工作池
	var workerPool *workerpool.Pool
	if cfg.WorkerPool.Enabled && cfg.WorkerPool.WorkerCount > 0 {
		workerPool = workerpool.NewPool(cfg.WorkerPool.WorkerCount)
		cleanupFuncs = append(cleanupFuncs, func() {
			workerPool.Stop()
		})
	}

	// 直接创建SQLite应用，不再需要模式检查
	return createSQLiteApp(cfg, container, cleanup)
}

// createSQLiteApp 创建SQLite模式应用
func createSQLiteApp(cfg *config.Config, container *di.Container, cleanup func()) (*App, error) {
	log.Println("创建SQLite数据库版应用...")

	// 创建SQLite存储库
	repo, err := repository.NewSQLiteRepository(cfg.GetDatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("创建SQLite存储库失败: %v", err)
	}

	// 初始化数据库
	if closer, ok := interface{}(repo).(interface{ Close() error }); ok {
		cleanupFuncs := []func(){cleanup}
		cleanupFuncs = append(cleanupFuncs, func() {
			closer.Close()
		})

		newCleanup := func() {
			for _, fn := range cleanupFuncs {
				fn()
			}
		}
		cleanup = newCleanup
	}

	// 创建服务
	individualService := services.NewIndividualService(repo, repo)
	familyService := services.NewFamilyService(repo, repo)

	// 创建处理器
	individualHandler := handlers.NewIndividualHandler(individualService)
	familyHandler := handlers.NewFamilyHandler(familyService)

	// 设置路由
	router := setupRouter(individualHandler, familyHandler, cfg)

	return &App{
		router:  router,
		cleanup: cleanup,
	}, nil
}

// setupRouter 设置路由
func setupRouter(individualHandler *handlers.IndividualHandler, familyHandler *handlers.FamilyHandler, cfg *config.Config) *mux.Router {
	router := mux.NewRouter()

	// 添加中间件
	if cfg.Server.EnableCORS {
		router.Use(corsMiddleware)
	}

	if cfg.IsDevelopment() {
		router.Use(loggingMiddleware)
	}

	// API路由
	api := router.PathPrefix("/api/v1").Subrouter()

	// 个人信息路由
	individuals := api.PathPrefix("/individuals").Subrouter()
	individuals.HandleFunc("", individualHandler.CreateIndividual).Methods("POST")
	individuals.HandleFunc("", individualHandler.SearchIndividuals).Methods("GET")
	individuals.HandleFunc("/{id:[0-9]+}", individualHandler.GetIndividual).Methods("GET")
	individuals.HandleFunc("/{id:[0-9]+}", individualHandler.UpdateIndividual).Methods("PUT")
	individuals.HandleFunc("/{id:[0-9]+}", individualHandler.DeleteIndividual).Methods("DELETE")

	// 关系路由
	individuals.HandleFunc("/{id:[0-9]+}/children", individualHandler.GetChildren).Methods("GET")
	individuals.HandleFunc("/{id:[0-9]+}/parents", individualHandler.GetParents).Methods("GET")
	individuals.HandleFunc("/{id:[0-9]+}/siblings", individualHandler.GetSiblings).Methods("GET")
	individuals.HandleFunc("/{id:[0-9]+}/spouses", individualHandler.GetSpouses).Methods("GET")
	individuals.HandleFunc("/{id:[0-9]+}/ancestors", individualHandler.GetAncestors).Methods("GET")
	individuals.HandleFunc("/{id:[0-9]+}/descendants", individualHandler.GetDescendants).Methods("GET")
	individuals.HandleFunc("/{id:[0-9]+}/family-tree", individualHandler.GetFamilyTree).Methods("GET")

	// 添加父母路由
	individuals.HandleFunc("/{id:[0-9]+}/parents", individualHandler.AddParent).Methods("POST")

	// 配偶关系路由
	individuals.HandleFunc("/{id:[0-9]+}/add-spouse", familyHandler.AddSpouse).Methods("POST")

	// 家庭关系路由
	families := api.PathPrefix("/families").Subrouter()
	families.HandleFunc("", familyHandler.CreateFamily).Methods("POST")
	families.HandleFunc("/{id:[0-9]+}", familyHandler.GetFamily).Methods("GET")
	families.HandleFunc("/{id:[0-9]+}", familyHandler.UpdateFamily).Methods("PUT")
	families.HandleFunc("/{id:[0-9]+}", familyHandler.DeleteFamily).Methods("DELETE")
	families.HandleFunc("/{id:[0-9]+}/children", familyHandler.AddChild).Methods("POST")
	families.HandleFunc("/{id:[0-9]+}/children/{childId:[0-9]+}", familyHandler.RemoveChild).Methods("DELETE")
	families.HandleFunc("/husband/{id:[0-9]+}", familyHandler.GetFamiliesByHusband).Methods("GET")

	// 健康检查
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		response := map[string]string{
			"status":   "ok",
			"message":  "家谱系统SQLite版运行中",
			"database": cfg.Database.Path,
		}
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// API文档页面
	router.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		pageTitle := "家谱系统 - SQLite版"
		modeInfo := fmt.Sprintf(`
			<div class="info">
				<strong>模式:</strong> SQLite数据库版<br>
				<strong>数据库:</strong> %s<br>
				<strong>状态:</strong> 运行中<br>
				<strong>特性:</strong> 数据持久化存储
			</div>`, cfg.Database.Path)
		modeDescription := `
			<li>所有数据持久化存储在SQLite数据库中</li>
			<li>支持完整的CRUD操作和事务</li>
			<li>数据在重启后保持</li>`

		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>%s - API文档</title>
			<meta charset="utf-8">
			<style>
				body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 40px; }
				.container { max-width: 800px; margin: 0 auto; }
				.endpoint { background: #f5f5f5; padding: 10px; margin: 5px 0; border-radius: 5px; }
				.endpoint a { text-decoration: none; color: #0066cc; }
				.endpoint a:hover { text-decoration: underline; }
				.info { background: #e8f4fd; padding: 15px; border-radius: 8px; border-left: 4px solid #0066cc; margin: 20px 0; }
				.mode-switch { background: #fff3cd; padding: 15px; border-radius: 8px; border-left: 4px solid #ffa500; margin: 20px 0; }
				.ui-link { background: #28a745; color: white; padding: 15px 30px; text-decoration: none; border-radius: 8px; display: inline-block; margin: 20px 0; font-weight: bold; }
				.ui-link:hover { background: #218838; color: white; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>🌳 %s - API文档</h1>
				
				<a href="/ui" class="ui-link">🖥️ 打开管理界面</a>
				
				%s

				<div class="mode-switch">
					<strong>💡 启动方式:</strong><br>
					• 直接启动: <code>go run main.go</code>
				</div>

				<h2>🔗 API 端点</h2>
				<div class="endpoint"><strong>GET</strong> <a href="/api/v1/individuals">/api/v1/individuals</a> - 获取所有个人信息</div>
				<div class="endpoint"><strong>POST</strong> /api/v1/individuals - 创建个人信息</div>
				<div class="endpoint"><strong>GET</strong> <a href="/api/v1/individuals/1">/api/v1/individuals/1</a> - 获取ID为1的个人信息</div>
				<div class="endpoint"><strong>PUT</strong> /api/v1/individuals/{id} - 更新个人信息</div>
				<div class="endpoint"><strong>DELETE</strong> /api/v1/individuals/{id} - 删除个人信息</div>

				<h3>关系查询</h3>
				<div class="endpoint"><strong>GET</strong> <a href="/api/v1/individuals/1/children">/api/v1/individuals/1/children</a> - 获取子女</div>
				<div class="endpoint"><strong>GET</strong> <a href="/api/v1/individuals/3/parents">/api/v1/individuals/3/parents</a> - 获取父母</div>
				<div class="endpoint"><strong>GET</strong> <a href="/api/v1/individuals/3/siblings">/api/v1/individuals/3/siblings</a> - 获取兄弟姐妹</div>
				<div class="endpoint"><strong>GET</strong> <a href="/api/v1/individuals/1/ancestors">/api/v1/individuals/1/ancestors</a> - 获取祖先</div>
				<div class="endpoint"><strong>GET</strong> <a href="/api/v1/individuals/1/descendants">/api/v1/individuals/1/descendants</a> - 获取后代</div>
				<div class="endpoint"><strong>GET</strong> <a href="/api/v1/individuals/1/family-tree">/api/v1/individuals/1/family-tree</a> - 获取家族树</div>

				<h3>其他</h3>
				<div class="endpoint"><strong>GET</strong> <a href="/health">/health</a> - 健康检查</div>

				<h2>📊 示例数据</h2>
				<ul>
					<li><strong>张老爷子</strong> (ID: 1) - 农民，1920年出生</li>
					<li><strong>李老太太</strong> (ID: 2) - 家庭主妇，1925年出生</li>
					<li><strong>张伟</strong> (ID: 3) - 工程师，1950年出生，张老爷子和李老太太的儿子</li>
					<li><strong>王丽</strong> (ID: 4) - 教师，1955年出生</li>
					<li><strong>张明</strong> (ID: 5) - 医生，1975年出生，张伟和王丽的儿子</li>
					<li><strong>李美</strong> (ID: 6) - 护士，1978年出生</li>
				</ul>

				<h2>💡 特性说明</h2>
				<ul>
					%s
					<li>支持复杂的家族关系查询</li>
					<li>API返回JSON格式数据</li>
					<li>使用 <code>?limit=10&offset=0</code> 进行分页查询</li>
					<li>支持按姓名、职业、备注搜索</li>
				</ul>
			</div>
		</body>
		</html>
		`, pageTitle, pageTitle, modeInfo, modeDescription)
	}).Methods("GET")

	// 静态文件服务
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// 测试页面
	router.HandleFunc("/test_add_child.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "test_add_child.html")
	}).Methods("GET")

	// 测试配偶页面
	router.HandleFunc("/test_spouses.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "test_spouses.html")
	}).Methods("GET")

	// UI管理界面
	router.HandleFunc("/ui", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/index.html", http.StatusFound)
	}).Methods("GET")

	// 首页 - 重定向到UI界面
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui", http.StatusFound)
	}).Methods("GET")

	return router
}

// corsMiddleware CORS中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf(
			"%s %s %s %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}

// initializeDatabase 初始化数据库（创建表和示例数据）
func initializeDatabase(db *sql.DB) error {
	// 读取SQL初始化脚本
	sqlFile := filepath.Join("sql", "init.sql")
	sqlContent, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		return fmt.Errorf("读取SQL文件失败: %v", err)
	}

	// 清理SQL内容，移除注释
	lines := strings.Split(string(sqlContent), "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			cleanLines = append(cleanLines, line)
		}
	}
	cleanSQL := strings.Join(cleanLines, " ")

	// 使用更智能的分割方法
	statements := splitSQLStatements(cleanSQL)

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		fmt.Printf("执行SQL语句 %d: %s...\n", i+1, truncateString(stmt, 50))
		_, err := db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("执行SQL语句失败 '%s': %v", truncateString(stmt, 100), err)
		}
	}

	fmt.Println("✅ 数据库初始化完成")
	return nil
}

// splitSQLStatements 智能分割SQL语句
func splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	var stringChar byte
	beginEndLevel := 0

	// 将SQL转换为upper case来检测关键字
	upperSQL := strings.ToUpper(sql)

	for i := 0; i < len(sql); i++ {
		char := sql[i]

		// 处理字符串
		if (char == '\'' || char == '"') && (i == 0 || sql[i-1] != '\\') {
			if !inString {
				inString = true
				stringChar = char
			} else if char == stringChar {
				inString = false
			}
		}

		// 检测BEGIN关键字
		if !inString && i <= len(upperSQL)-5 {
			if upperSQL[i:i+5] == "BEGIN" && (i == 0 || !isAlphaNumeric(upperSQL[i-1])) && (i+5 >= len(upperSQL) || !isAlphaNumeric(upperSQL[i+5])) {
				beginEndLevel++
			}
		}

		// 检测END关键字
		if !inString && i <= len(upperSQL)-3 {
			if upperSQL[i:i+3] == "END" && (i == 0 || !isAlphaNumeric(upperSQL[i-1])) && (i+3 >= len(upperSQL) || !isAlphaNumeric(upperSQL[i+3])) {
				beginEndLevel--
			}
		}

		// 如果遇到分号且不在字符串中且不在BEGIN...END块中
		if char == ';' && !inString && beginEndLevel == 0 {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
			continue
		}

		current.WriteByte(char)
	}

	// 添加最后一个语句
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

// isAlphaNumeric 检查字符是否为字母或数字
func isAlphaNumeric(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
}

// truncateString 截断字符串用于显示
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
