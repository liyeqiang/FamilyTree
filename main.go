package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
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
	"familytree/interfaces"
	"familytree/pkg/di"
	"familytree/pkg/middleware"
	"familytree/pkg/workerpool"
	"familytree/repository"
	"familytree/services"

	"github.com/gorilla/mux"
)

// DocsPageData 文档页面数据结构
type DocsPageData struct {
	PageTitle    string
	DatabasePath string
}

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
	router     *mux.Router
	db         *sql.DB
	cache      *repository.CacheRepository
	workerPool *workerpool.Pool
	container  *di.Container
	cleanup    func()
}

// createApp 创建应用实例
func createApp(cfg *config.Config) (*App, error) {
	var cleanupFuncs []func()

	cleanup := func() {
		log.Println("正在清理资源...")
		for _, fn := range cleanupFuncs {
			fn()
		}
		log.Println("资源清理完成")
	}

	// 创建依赖注入容器
	container := di.NewContainer()
	log.Println("✅ 依赖注入容器已创建")

	// 创建工作池
	var workerPool *workerpool.Pool
	if cfg.WorkerPool.Enabled && cfg.WorkerPool.WorkerCount > 0 {
		workerPool = workerpool.NewPool(cfg.WorkerPool.WorkerCount)
		cleanupFuncs = append(cleanupFuncs, func() {
			log.Println("正在停止工作池...")
			workerPool.Stop()
			log.Println("✅ 工作池已停止")
		})
		log.Printf("✅ 工作池已创建 (工作者数量: %d)", cfg.WorkerPool.WorkerCount)
	}

	// 直接创建SQLite应用，集成所有模块
	return createSQLiteApp(cfg, container, workerPool, cleanup)
}

// createSQLiteApp 创建SQLite模式应用
func createSQLiteApp(cfg *config.Config, container *di.Container, workerPool *workerpool.Pool, cleanup func()) (*App, error) {
	log.Println("🔧 创建SQLite数据库版应用...")

	var cleanupFuncs []func()
	cleanupFuncs = append(cleanupFuncs, cleanup)

	// 创建SQLite存储库
	repo, err := repository.NewSQLiteRepository(cfg.GetDatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("创建SQLite存储库失败: %v", err)
	}
	log.Println("✅ SQLite存储库已创建")

	// 注册存储库到容器
	container.Register(repo)

	// 初始化数据库连接池清理
	if closer, ok := interface{}(repo).(interface{ Close() error }); ok {
		cleanupFuncs = append(cleanupFuncs, func() {
			log.Println("正在关闭数据库连接...")
			closer.Close()
			log.Println("✅ 数据库连接已关闭")
		})
	}

	// 创建Redis缓存（如果启用）
	var cacheRepo *repository.CacheRepository
	if cfg.RedisEnabled && cfg.CacheEnabled {
		log.Println("🔄 初始化Redis缓存...")
		cache, err := repository.NewCacheRepository(&cfg.Redis, 10*time.Minute)
		if err != nil {
			log.Printf("⚠️  Redis缓存初始化失败: %v，继续使用无缓存模式", err)
		} else {
			cacheRepo = cache
			cleanupFuncs = append(cleanupFuncs, func() {
				log.Println("正在关闭Redis连接...")
				cacheRepo.Close()
				log.Println("✅ Redis连接已关闭")
			})
			container.Register(cacheRepo)
			log.Println("✅ Redis缓存已启用")
		}
	}

	// 创建服务层
	baseIndividualService := services.NewIndividualService(repo, repo)
	baseFamilyService := services.NewFamilyService(repo, repo)

	// 如果有缓存，使用缓存装饰器
	var individualService interfaces.IndividualService
	if cacheRepo != nil {
		individualService = services.NewCachedIndividualService(baseIndividualService, cacheRepo)
		log.Println("✅ 个人信息服务（带缓存）已创建")
	} else {
		individualService = baseIndividualService
		log.Println("✅ 个人信息服务已创建")
	}

	// 注册服务到容器
	container.Register(individualService)
	container.Register(baseFamilyService)

	// 创建处理器
	individualHandler := handlers.NewIndividualHandler(individualService)
	familyHandler := handlers.NewFamilyHandler(baseFamilyService)
	log.Println("✅ HTTP处理器已创建")

	// 注册处理器到容器
	container.Register(individualHandler)
	container.Register(familyHandler)

	// 设置路由（集成高级中间件）
	router := setupAdvancedRouter(individualHandler, familyHandler, cfg)
	log.Println("✅ 高级路由和中间件已配置")

	// 构建最终的清理函数
	finalCleanup := func() {
		for _, fn := range cleanupFuncs {
			fn()
		}
	}

	return &App{
		router:     router,
		cache:      cacheRepo,
		workerPool: workerPool,
		container:  container,
		cleanup:    finalCleanup,
	}, nil
}

// setupAdvancedRouter 设置带高级中间件的路由
func setupAdvancedRouter(individualHandler *handlers.IndividualHandler, familyHandler *handlers.FamilyHandler, cfg *config.Config) *mux.Router {
	router := mux.NewRouter()

	// 添加中间件（使用Gorilla mux兼容的方式）

	// 1. 恢复中间件（最外层）
	if cfg.Middleware.EnableRecovery {
		router.Use(func(next http.Handler) http.Handler {
			return middleware.Recover(next)
		})
		log.Println("✅ 恢复中间件已启用")
	}

	// 2. 日志中间件
	if cfg.Middleware.EnableLogging {
		router.Use(func(next http.Handler) http.Handler {
			return middleware.Logger(next)
		})
		log.Println("✅ 日志中间件已启用")
	}

	// 3. CORS中间件
	if cfg.Middleware.EnableCORS {
		router.Use(func(next http.Handler) http.Handler {
			return middleware.CORS(next)
		})
		log.Println("✅ CORS中间件已启用")
	}

	// 4. 限流中间件
	if cfg.Middleware.EnableRateLimit {
		rateLimitMiddleware := middleware.RateLimit(
			cfg.Middleware.RateLimit.RequestsPerMinute,
			time.Minute,
		)
		router.Use(func(next http.Handler) http.Handler {
			return rateLimitMiddleware(next)
		})
		log.Printf("✅ 限流中间件已启用 (每分钟%d次请求)", cfg.Middleware.RateLimit.RequestsPerMinute)
	}

	// 5. 指标中间件
	var metricsCollector *middleware.Metrics
	if cfg.Middleware.EnableMetrics {
		metricsCollector = middleware.NewMetrics()
		router.Use(func(next http.Handler) http.Handler {
			return metricsCollector.MetricsMiddleware(next)
		})
		log.Println("✅ 指标中间件已启用")
	}

	// API路由（带超时中间件）
	api := router.PathPrefix("/api/v1").Subrouter()

	// 超时中间件（针对API路由）
	timeoutMiddleware := middleware.Timeout(30 * time.Second)
	api.Use(func(next http.Handler) http.Handler {
		return timeoutMiddleware(next)
	})

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

	// 健康检查（带缓存检查）
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"status":   "ok",
			"message":  "家谱系统SQLite版运行中",
			"database": cfg.Database.Path,
			"features": map[string]bool{
				"redis_cache":   cfg.RedisEnabled && cfg.CacheEnabled,
				"worker_pool":   cfg.WorkerPool.Enabled,
				"rate_limiting": cfg.Middleware.EnableRateLimit,
				"metrics":       cfg.Middleware.EnableMetrics,
			},
			"timestamp": time.Now(),
		}
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// 指标查看API（如果启用了指标）
	if cfg.Middleware.EnableMetrics && metricsCollector != nil {
		api.HandleFunc("/metrics", metricsCollector.MetricsHandler).Methods("GET")
		log.Println("✅ 指标查看API已启用 (/api/v1/metrics)")
	}

	// 缓存管理API（如果启用了缓存）
	if cfg.RedisEnabled && cfg.CacheEnabled {
		cache := router.PathPrefix("/api/v1/cache").Subrouter()
		cache.Use(func(next http.Handler) http.Handler {
			return timeoutMiddleware(next)
		})

		// 清除所有缓存
		cache.HandleFunc("/clear", func(w http.ResponseWriter, r *http.Request) {
			// TODO: 实现缓存清除逻辑
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "缓存清除功能待实现",
			})
		}).Methods("DELETE")
	}

	// API文档页面 - 使用模板文件
	router.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// 解析模板文件
		tmpl, err := template.ParseFiles("static/docs.html")
		if err != nil {
			http.Error(w, "加载模板文件失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 准备模板数据
		data := DocsPageData{
			PageTitle:    "家谱系统 - SQLite版（完整功能）",
			DatabasePath: cfg.Database.Path,
		}

		// 执行模板
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "渲染模板失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
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
