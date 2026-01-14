package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"langchaingo-demo/models"
	"langchaingo-demo/services"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	aiService *services.AIService
}

func NewChatHandler(aiService *services.AIService) *ChatHandler {
	return &ChatHandler{
		aiService: aiService,
	}
}

// ChatStream SSE 流式聊天
// POST /api/chat/stream
func (h *ChatHandler) ChatStream(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 获取 writer 用于 flush
	w := c.Writer

	callbacks := services.StreamCallbacks{
		OnStart: func(conversationID uint) error {
			data, _ := json.Marshal(gin.H{
				"type":            "start",
				"conversation_id": conversationID,
			})
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.Flush()
			return nil
		},
		OnContent: func(chunk string) error {
			data, _ := json.Marshal(gin.H{
				"type":    "content",
				"content": chunk,
			})
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.Flush()
			return nil
		},
	}

	_, conversationID, err := h.aiService.ChatStream(
		c.Request.Context(),
		req.ConversationID,
		req.Message,
		callbacks,
	)

	if err != nil {
		data, _ := json.Marshal(gin.H{
			"type":  "error",
			"error": err.Error(),
		})
		fmt.Fprintf(w, "data: %s\n\n", data)
		w.Flush()
		return
	}

	// 发送完成信号
	data, _ := json.Marshal(gin.H{
		"type":            "done",
		"conversation_id": conversationID,
	})
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.Flush()
}

// GetConversations 获取所有会话列表
// GET /api/conversations
func (h *ChatHandler) GetConversations(c *gin.Context) {
	conversations, err := h.aiService.ListConversations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

// GetConversationHistory 获取会话历史
// GET /api/conversations/:id/messages
func (h *ChatHandler) GetConversationHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	messages, err := h.aiService.GetConversationHistory(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// HealthCheck 健康检查
// GET /health
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetSettings 获取当前设置
// GET /api/settings
func (h *ChatHandler) GetSettings(c *gin.Context) {
	settings := h.aiService.GetConfig()
	c.JSON(http.StatusOK, settings)
}

// UpdateSettings 更新设置
// POST /api/settings
func (h *ChatHandler) UpdateSettings(c *gin.Context) {
	var req models.SettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.aiService.UpdateConfig(req.Provider, req.Model, req.BaseURL, req.APIKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设置已更新"})
}

// GetProviders 获取预设提供商列表
// GET /api/providers
func (h *ChatHandler) GetProviders(c *gin.Context) {
	providers := h.aiService.GetProviderPresets()
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}
