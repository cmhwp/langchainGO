package main

import (
	"log"

	"langchaingo-demo/config"
	"langchaingo-demo/database"
	"langchaingo-demo/routes"
	"langchaingo-demo/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	if err := database.Init(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	// 初始化 AI 服务
	aiService, err := services.NewAIService(&cfg.AI)
	if err != nil {
		log.Fatalf("Failed to initialize AI service: %v", err)
	}
	log.Println("AI service initialized successfully")

	// 创建 Gin 引擎
	r := gin.Default()

	// 设置路由
	routes.SetupRoutes(r, aiService)

	// 启动服务器
	log.Printf("Server starting on port %s...", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
