package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"langchaingo-demo/config"
	"langchaingo-demo/database"
	"langchaingo-demo/models"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type AIService struct {
	llm llms.Model
	cfg *config.AIConfig
	mu  sync.RWMutex
}

func NewAIService(cfg *config.AIConfig) (*AIService, error) {
	service := &AIService{
		cfg: cfg,
	}

	if err := service.initLLM(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *AIService) initLLM() error {
	opts := []openai.Option{
		openai.WithModel(s.cfg.Model),
	}

	if s.cfg.APIKey != "" {
		opts = append(opts, openai.WithToken(s.cfg.APIKey))
	}

	if s.cfg.BaseURL != "" {
		opts = append(opts, openai.WithBaseURL(s.cfg.BaseURL))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	s.llm = llm
	return nil
}

// UpdateConfig 更新配置
func (s *AIService) UpdateConfig(provider, model, baseURL, apiKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cfg.Provider = provider
	s.cfg.Model = model
	s.cfg.BaseURL = baseURL
	s.cfg.APIKey = apiKey

	return s.initLLM()
}

// GetConfig 获取当前配置
func (s *AIService) GetConfig() models.SettingsResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 掩码处理 API Key
	maskedKey := ""
	if len(s.cfg.APIKey) > 8 {
		maskedKey = s.cfg.APIKey[:4] + "****" + s.cfg.APIKey[len(s.cfg.APIKey)-4:]
	} else if len(s.cfg.APIKey) > 0 {
		maskedKey = "****"
	}

	return models.SettingsResponse{
		Provider: s.cfg.Provider,
		Model:    s.cfg.Model,
		BaseURL:  s.cfg.BaseURL,
		APIKey:   maskedKey,
	}
}

// GetProviderPresets 获取预设提供商列表
func (s *AIService) GetProviderPresets() []models.ProviderPreset {
	return []models.ProviderPreset{
		{
			Name:    "OpenAI",
			BaseURL: "https://api.openai.com/v1",
			Models:  []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo"},
		},
		{
			Name:    "Moonshot (Kimi)",
			BaseURL: "https://api.moonshot.cn/v1",
			Models:  []string{"moonshot-v1-128k", "moonshot-v1-32k", "moonshot-v1-8k"},
		},
		{
			Name:    "DeepSeek",
			BaseURL: "https://api.deepseek.com/v1",
			Models:  []string{"deepseek-chat", "deepseek-coder"},
		},
		{
			Name:    "智谱 (Zhipu)",
			BaseURL: "https://open.bigmodel.cn/api/paas/v4",
			Models:  []string{"glm-4-plus", "glm-4", "glm-4-flash"},
		},
		{
			Name:    "通义千问 (Qwen)",
			BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
			Models:  []string{"qwen-turbo", "qwen-plus", "qwen-max"},
		},
		{
			Name:    "百川 (Baichuan)",
			BaseURL: "https://api.baichuan-ai.com/v1",
			Models:  []string{"Baichuan4", "Baichuan3-Turbo", "Baichuan2-Turbo"},
		},
		{
			Name:    "Ollama (本地)",
			BaseURL: "http://localhost:11434/v1",
			Models:  []string{"llama3", "llama2", "mistral", "codellama", "qwen2"},
		},
		{
			Name:    "自定义",
			BaseURL: "",
			Models:  []string{},
		},
	}
}

// StreamCallbacks 流式回调
type StreamCallbacks struct {
	OnStart   func(conversationID uint) error
	OnContent func(chunk string) error
}

// ChatStream 流式聊天
func (s *AIService) ChatStream(ctx context.Context, conversationID uint, userMessage string, callbacks StreamCallbacks) (string, uint, error) {
	s.mu.RLock()
	llm := s.llm
	s.mu.RUnlock()

	db := database.GetDB()

	// 如果没有会话ID，创建新会话
	if conversationID == 0 {
		conversation := &models.Conversation{
			Title: truncateString(userMessage, 50),
		}
		if err := db.Create(conversation).Error; err != nil {
			return "", 0, err
		}
		conversationID = conversation.ID
	}

	// 通知开始，传递 conversationID
	if callbacks.OnStart != nil {
		if err := callbacks.OnStart(conversationID); err != nil {
			return "", conversationID, err
		}
	}

	// 保存用户消息
	userMsg := &models.Message{
		ConversationID: conversationID,
		Role:           "user",
		Content:        userMessage,
	}
	if err := db.Create(userMsg).Error; err != nil {
		return "", conversationID, err
	}

	// 获取历史消息
	var historyMessages []models.Message
	db.Where("conversation_id = ?", conversationID).Order("created_at asc").Find(&historyMessages)

	// 构建 langchain 消息
	var chatMessages []llms.MessageContent

	// 添加系统提示
	chatMessages = append(chatMessages, llms.MessageContent{
		Role:  llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{llms.TextContent{Text: "你是一个有帮助的AI助手。"}},
	})

	// 添加历史消息
	for _, msg := range historyMessages {
		var role llms.ChatMessageType
		switch msg.Role {
		case "user":
			role = llms.ChatMessageTypeHuman
		case "assistant":
			role = llms.ChatMessageTypeAI
		default:
			continue
		}
		chatMessages = append(chatMessages, llms.MessageContent{
			Role:  role,
			Parts: []llms.ContentPart{llms.TextContent{Text: msg.Content}},
		})
	}

	// 收集完整回复
	var fullResponse strings.Builder

	// 流式调用 LLM
	_, err := llm.GenerateContent(ctx, chatMessages,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			text := string(chunk)
			fullResponse.WriteString(text)
			if callbacks.OnContent != nil {
				return callbacks.OnContent(text)
			}
			return nil
		}),
	)
	if err != nil {
		return "", conversationID, fmt.Errorf("failed to generate response: %w", err)
	}

	// 保存助手回复
	assistantMsg := &models.Message{
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        fullResponse.String(),
	}
	if err := db.Create(assistantMsg).Error; err != nil {
		return "", conversationID, err
	}

	return fullResponse.String(), conversationID, nil
}

// GetConversationHistory 获取会话历史
func (s *AIService) GetConversationHistory(conversationID uint) ([]models.Message, error) {
	db := database.GetDB()
	var messages []models.Message
	err := db.Where("conversation_id = ?", conversationID).Order("created_at asc").Find(&messages).Error
	return messages, err
}

// ListConversations 列出所有会话
func (s *AIService) ListConversations() ([]models.Conversation, error) {
	db := database.GetDB()
	var conversations []models.Conversation
	err := db.Order("updated_at desc").Find(&conversations).Error
	return conversations, err
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
