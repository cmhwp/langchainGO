package routes

import (
	"langchaingo-demo/handlers"
	"langchaingo-demo/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, aiService *services.AIService) {
	// 健康检查
	r.GET("/health", handlers.HealthCheck)

	// 创建处理器
	chatHandler := handlers.NewChatHandler(aiService)

	// API 路由组
	api := r.Group("/api")
	{
		// 聊天接口 (SSE 流式)
		api.POST("/chat/stream", chatHandler.ChatStream)

		// 会话管理
		api.GET("/conversations", chatHandler.GetConversations)
		api.GET("/conversations/:id/messages", chatHandler.GetConversationHistory)

		// 设置管理
		api.GET("/settings", chatHandler.GetSettings)
		api.POST("/settings", chatHandler.UpdateSettings)
		api.GET("/providers", chatHandler.GetProviders)
	}
}
