package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"familytree/config"
)

func main() {
	fmt.Println("🗃️ 初始化SQLite数据库...")

	// 加载配置
	cfg := config.LoadConfig()
	
	// 连接数据库
	db, err := cfg.Connect()
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 读取并执行SQL文件
	sqlFile, err := os.Open("sql/init.sql")
	if err != nil {
		log.Fatalf("打开SQL文件失败: %v", err)
	}
	defer sqlFile.Close()

	// 读取整个文件
	sqlContent, err := io.ReadAll(sqlFile)
	if err != nil {
		log.Fatalf("读取SQL文件失败: %v", err)
	}

	// 分割SQL语句
	statements := strings.Split(string(sqlContent), ";")
	
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		// 执行SQL语句
		_, err := db.Exec(stmt)
		if err != nil {
			// 忽略一些常见的错误（如表已存在）
			if !strings.Contains(err.Error(), "already exists") &&
			   !strings.Contains(err.Error(), "duplicate") {
				log.Printf("执行SQL语句失败: %v\n语句: %s", err, stmt)
			}
		}
	}

	fmt.Println("✅ 数据库初始化完成！")
	fmt.Printf("📁 数据库文件: %s\n", cfg.DBPath)
	
	// 验证数据
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM individuals").Scan(&count)
	if err != nil {
		log.Printf("查询个人数据失败: %v", err)
	} else {
		fmt.Printf("👥 个人记录数: %d\n", count)
	}
}

// 交互式初始化
func interactiveInit() {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Print("请输入数据库文件路径 (默认: familytree.db): ")
	dbPath, _ := reader.ReadString('\n')
	dbPath = strings.TrimSpace(dbPath)
	if dbPath == "" {
		dbPath = "familytree.db"
	}
	
	// 设置环境变量
	os.Setenv("DB_PATH", dbPath)
	
	fmt.Printf("将使用数据库文件: %s\n", dbPath)
} 