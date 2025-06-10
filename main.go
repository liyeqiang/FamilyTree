package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"familytree/config"
	"familytree/handlers"
	"familytree/models"
	"familytree/repository"
	"familytree/services"

	"github.com/gorilla/mux"
)

// DemoRepository 内存存储库用于演示模式
type DemoRepository struct {
	individuals []models.Individual
	nextID      int
}

func NewDemoRepository() *DemoRepository {
	repo := &DemoRepository{
		individuals: make([]models.Individual, 0),
		nextID:      1,
	}
	
	// 添加示例数据
	now := time.Now()
	birthDate1950 := time.Date(1950, 1, 15, 0, 0, 0, 0, time.UTC)
	birthDate1955 := time.Date(1955, 3, 20, 0, 0, 0, 0, time.UTC)
	birthDate1975 := time.Date(1975, 6, 10, 0, 0, 0, 0, time.UTC)
	birthDate1978 := time.Date(1978, 9, 15, 0, 0, 0, 0, time.UTC)
	birthDate2005 := time.Date(2005, 12, 25, 0, 0, 0, 0, time.UTC)
	
	individuals := []models.Individual{
		{IndividualID: 1, FullName: "张伟", Gender: models.GenderMale, BirthDate: &birthDate1950, Occupation: "工程师", Notes: "家族族长", CreatedAt: now, UpdatedAt: now},
		{IndividualID: 2, FullName: "李丽", Gender: models.GenderFemale, BirthDate: &birthDate1955, Occupation: "教师", Notes: "张伟的妻子", CreatedAt: now, UpdatedAt: now},
		{IndividualID: 3, FullName: "张明", Gender: models.GenderMale, BirthDate: &birthDate1975, Occupation: "医生", Notes: "张伟和李丽的儿子", FatherID: &[]int{1}[0], MotherID: &[]int{2}[0], CreatedAt: now, UpdatedAt: now},
		{IndividualID: 4, FullName: "王美", Gender: models.GenderFemale, BirthDate: &birthDate1978, Occupation: "护士", Notes: "张明的妻子", CreatedAt: now, UpdatedAt: now},
		{IndividualID: 5, FullName: "张小宝", Gender: models.GenderMale, BirthDate: &birthDate2005, Occupation: "", Notes: "张明和王美的儿子", FatherID: &[]int{3}[0], MotherID: &[]int{4}[0], CreatedAt: now, UpdatedAt: now},
	}
	
	repo.individuals = individuals
	repo.nextID = 6
	
	return repo
}

// DemoRepository 实现 IndividualRepository 接口
func (r *DemoRepository) CreateIndividual(ctx context.Context, individual *models.Individual) (*models.Individual, error) {
	individual.IndividualID = r.nextID
	individual.CreatedAt = time.Now()
	individual.UpdatedAt = time.Now()
	r.nextID++
	
	r.individuals = append(r.individuals, *individual)
	return individual, nil
}

func (r *DemoRepository) GetIndividualByID(ctx context.Context, id int) (*models.Individual, error) {
	for _, individual := range r.individuals {
		if individual.IndividualID == id {
			return &individual, nil
		}
	}
	return nil, fmt.Errorf("个人不存在")
}

func (r *DemoRepository) UpdateIndividual(ctx context.Context, id int, individual *models.Individual) (*models.Individual, error) {
	for i, existing := range r.individuals {
		if existing.IndividualID == id {
			individual.IndividualID = id
			individual.CreatedAt = existing.CreatedAt
			individual.UpdatedAt = time.Now()
			r.individuals[i] = *individual
			return individual, nil
		}
	}
	return nil, fmt.Errorf("个人不存在")
}

func (r *DemoRepository) DeleteIndividual(ctx context.Context, id int) error {
	for i, individual := range r.individuals {
		if individual.IndividualID == id {
			r.individuals = append(r.individuals[:i], r.individuals[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("个人不存在")
}

func (r *DemoRepository) SearchIndividuals(ctx context.Context, query string, limit, offset int) ([]models.Individual, int, error) {
	var results []models.Individual
	for _, individual := range r.individuals {
		if query == "" || contains(individual.FullName, query) || contains(individual.Notes, query) {
			results = append(results, individual)
		}
	}
	
	total := len(results)
	
	// 分页
	start := offset
	if start > len(results) {
		start = len(results)
	}
	
	end := start + limit
	if end > len(results) {
		end = len(results)
	}
	
	return results[start:end], total, nil
}

func (r *DemoRepository) GetIndividualsByParentID(ctx context.Context, parentID int) ([]models.Individual, error) {
	var children []models.Individual
	for _, individual := range r.individuals {
		if (individual.FatherID != nil && *individual.FatherID == parentID) ||
		   (individual.MotherID != nil && *individual.MotherID == parentID) {
			children = append(children, individual)
		}
	}
	return children, nil
}

func (r *DemoRepository) GetIndividualsByIDs(ctx context.Context, ids []int) ([]models.Individual, error) {
	var results []models.Individual
	for _, id := range ids {
		for _, individual := range r.individuals {
			if individual.IndividualID == id {
				results = append(results, individual)
				break
			}
		}
	}
	return results, nil
}

func main() {
	// 检查命令行参数
	mode := "demo"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	switch mode {
	case "sqlite", "db":
		runSQLiteMode()
	case "demo", "memory":
		runDemoMode()
	default:
		fmt.Println("用法: go run main.go [demo|sqlite]")
		fmt.Println("  demo   - 内存演示模式（默认）")
		fmt.Println("  sqlite - SQLite数据库模式")
		os.Exit(1)
	}
}

// runDemoMode 运行演示模式（内存存储）
func runDemoMode() {
	fmt.Println("🚀 启动家谱系统（内存演示版）...")

	// 创建演示存储库和服务
	repo := NewDemoRepository()
	individualService := services.NewIndividualService(repo)

	// 创建处理器
	individualHandler := handlers.NewIndividualHandler(individualService)

	// 创建并配置路由器
	router := setupRouter(individualHandler, "demo", "")

	// 启动服务器
	startServer(router)
}

// runSQLiteMode 运行SQLite数据库模式
func runSQLiteMode() {
	fmt.Println("🚀 启动家谱系统（SQLite版）...")

	// 加载配置
	cfg := config.LoadConfig()
	
	// 连接数据库
	db, err := cfg.Connect()
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 创建存储库
	sqliteRepo := repository.NewSQLiteRepository(db)

	// 创建服务
	individualService := services.NewIndividualService(sqliteRepo)

	// 创建处理器
	individualHandler := handlers.NewIndividualHandler(individualService)

	// 创建并配置路由器
	router := setupRouter(individualHandler, "sqlite", cfg.DBPath)

	// 启动服务器
	startServer(router)
}

// setupRouter 设置路由器
func setupRouter(individualHandler *handlers.IndividualHandler, mode, dbPath string) *mux.Router {
	router := mux.NewRouter()

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

	// 健康检查
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		response := map[string]string{
			"status":  "ok",
			"mode":    mode,
		}
		if mode == "sqlite" {
			response["message"] = "家谱系统SQLite版运行中"
			response["database"] = dbPath
		} else {
			response["message"] = "家谱系统演示版运行中"
		}
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// 首页
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		
		var pageTitle, modeInfo, modeDescription string
		if mode == "sqlite" {
			pageTitle = "家谱系统 - SQLite版"
			modeInfo = fmt.Sprintf(`
				<div class="info">
					<strong>模式:</strong> SQLite数据库版<br>
					<strong>数据库:</strong> %s<br>
					<strong>状态:</strong> 运行中<br>
					<strong>特性:</strong> 数据持久化存储
				</div>`, dbPath)
			modeDescription = `
				<li>所有数据持久化存储在SQLite数据库中</li>
				<li>支持完整的CRUD操作和事务</li>
				<li>数据在重启后保持</li>`
		} else {
			pageTitle = "家谱系统 - 演示版"
			modeInfo = `
				<div class="info">
					<strong>模式:</strong> 内存演示版<br>
					<strong>状态:</strong> 运行中<br>
					<strong>特性:</strong> 无需数据库，即开即用
				</div>`
			modeDescription = `
				<li>数据存储在内存中，重启后重置</li>
				<li>无需数据库配置，开箱即用</li>
				<li>适合演示和测试</li>`
		}

		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>%s</title>
			<meta charset="utf-8">
			<style>
				body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 40px; }
				.container { max-width: 800px; margin: 0 auto; }
				.endpoint { background: #f5f5f5; padding: 10px; margin: 5px 0; border-radius: 5px; }
				.endpoint a { text-decoration: none; color: #0066cc; }
				.endpoint a:hover { text-decoration: underline; }
				.info { background: #e8f4fd; padding: 15px; border-radius: 8px; border-left: 4px solid #0066cc; margin: 20px 0; }
				.mode-switch { background: #fff3cd; padding: 15px; border-radius: 8px; border-left: 4px solid #ffa500; margin: 20px 0; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>🌳 %s</h1>
				
				%s

				<div class="mode-switch">
					<strong>💡 模式切换:</strong><br>
					• 演示模式: <code>go run main.go demo</code><br>
					• SQLite模式: <code>go run main.go sqlite</code>
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
					<li><strong>张伟</strong> (ID: 1) - 工程师，1950年出生</li>
					<li><strong>李丽</strong> (ID: 2) - 教师，1955年出生</li>
					<li><strong>张明</strong> (ID: 3) - 医生，1975年出生，张伟和李丽的儿子</li>
					<li><strong>王美</strong> (ID: 4) - 护士，1978年出生</li>
					<li><strong>张小宝</strong> (ID: 5) - 2005年出生，张明和王美的儿子</li>
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

	// 中间件
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware)

	return router
}

// startServer 启动HTTP服务器
func startServer(router *mux.Router) {
	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		fmt.Println("✅ 服务器启动在 http://localhost:8080")
		fmt.Println("📖 请访问 http://localhost:8080 查看API文档")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("正在关闭服务器...")

	// 给服务器5秒时间来完成正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("服务器强制关闭: %v", err)
	}

	fmt.Println("服务器已关闭")
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

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (len(substr) == 0 || func() bool {
			   for i := 0; i <= len(s)-len(substr); i++ {
				   if s[i:i+len(substr)] == substr {
					   return true
				   }
			   }
			   return false
		   }())
} 