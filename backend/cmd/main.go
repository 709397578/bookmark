package main

import (
	"log"
	"os"
	"pintree-backend/config"
	"pintree-backend/internal"
	"pintree-backend/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	
	// 设置GIN模式
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	log.Println("========================================")
	log.Println("  Pintree Backend Server Starting...")
	log.Println("========================================")
	log.Printf("Port: %s", cfg.Port)
	log.Printf("Database Driver: %s", cfg.DBDriver)
	if cfg.DBDriver == "sqlite" {
		log.Printf("Database Path: %s", cfg.DBPath)
	} else {
		log.Printf("Database URL: [REDACTED]")
	}
	log.Printf("JWT Secret: configured (length: %d)", len(cfg.JWTSecret))
	log.Printf("Frontend URL: %s", cfg.FrontendURL)
	log.Println("========================================")
	
	// 检查数据库配置
	if cfg.DBDriver == "sqlite" && cfg.DBPath == "" {
		log.Fatal("ERROR: DB_PATH is not configured! Please check your .env file")
	}
	if cfg.DBDriver == "postgres" && cfg.DatabaseURL == "" {
		log.Fatal("ERROR: DATABASE_URL is not configured! Please check your .env file")
	}
	
	// 初始化数据库
	internal.InitDB(cfg)
	
	// 设置路由
	router := routes.SetupRoutes(cfg)
	
	// 启动服务器
	addr := ":" + cfg.Port
	log.Printf("Server is running on http://localhost%s", addr)
	log.Printf("Health check: http://localhost%s/health", addr)
	
	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
