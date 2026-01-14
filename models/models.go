package models

import (
	"time"

	"gorm.io/gorm"
)

// Conversation 对话会话
type Conversation struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Title     string         `gorm:"size:255" json:"title"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Messages  []Message      `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

// Message 聊天消息
type Message struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	ConversationID uint           `gorm:"index" json:"conversation_id"`
	Role           string         `gorm:"size:20" json:"role"`  // user, assistant, system
	Content        string         `gorm:"type:text" json:"content"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	ConversationID uint   `json:"conversation_id"`
	Message        string `json:"message" binding:"required"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ConversationID uint   `json:"conversation_id"`
	Reply          string `json:"reply"`
}

// SettingsRequest 设置请求
type SettingsRequest struct {
	Provider string `json:"provider"`
	Model    string `json:"model" binding:"required"`
	BaseURL  string `json:"base_url" binding:"required"`
	APIKey   string `json:"api_key" binding:"required"`
}

// SettingsResponse 设置响应
type SettingsResponse struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"` // 返回时掩码处理
}

// ProviderPreset 预设提供商
type ProviderPreset struct {
	Name    string   `json:"name"`
	BaseURL string   `json:"base_url"`
	Models  []string `json:"models"`
}
