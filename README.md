# LangChainGo Chat

Go + Gin + SQLite 的流式聊天后端，配套一个 Vite + React 的 Web 聊天 UI。后端通过 `langchaingo` 走 OpenAI 兼容接口（OpenAI / Kimi / DeepSeek / 智谱 / 通义 / Ollama 等），并用 SSE 实时推送生成内容。

## 功能

- SSE 流式输出（`/api/chat/stream`），前端边生成边显示
- 会话/消息持久化（SQLite：`chat.db`）
- Web 端可配置 Provider / Base URL / Model / API Key
- Markdown 渲染（GFM）+ 代码块高亮

## 快速开始（本地开发）

### 1) 后端（Go）

1. 创建配置文件：
   - 复制 `.env.example` 为 `.env`
   - 填好 `AI_BASE_URL / AI_MODEL / AI_API_KEY`
2. 启动后端：
   - PowerShell：`go run .`
   - 启动后默认监听：`http://localhost:8080`

验证：`GET http://localhost:8080/health` 返回 `{"status":"ok"}`

### 2) 前端（React / Vite）

```bash
cd web
npm install
npm run dev
```

- 前端默认：`http://localhost:5173`
- `web/vite.config.ts` 已将 `/api` 代理到 `http://localhost:8080`

## 配置（.env）

见 `.env.example`。

- `SERVER_PORT`：后端端口（默认 `8080`）
- `DATABASE_DSN`：SQLite 文件路径（默认 `chat.db`）
- `AI_PROVIDER`：仅用于记录（当前实现走 OpenAI 兼容接口）
- `AI_MODEL`：模型名（如 `gpt-4o-mini` / `deepseek-chat` / `moonshot-v1-32k` / `llama3`）
- `AI_BASE_URL`：OpenAI 兼容 Base URL（如 `https://api.openai.com/v1` / `http://localhost:11434/v1`）
- `AI_API_KEY`：API Key（启动时会同步设置到 `OPENAI_API_KEY` 以供 `langchaingo` 使用）

## API

- `GET /health`：健康检查
- `POST /api/chat/stream`：SSE 流式聊天
  - 请求：`{ "conversation_id": 0, "message": "..." }`
  - SSE 事件 `data:` JSON：
    - `{"type":"start","conversation_id":123}`
    - `{"type":"content","content":"..."}`（多次）
    - `{"type":"done","conversation_id":123}`
    - `{"type":"error","error":"..."}`
- `GET /api/conversations`：会话列表
- `GET /api/conversations/:id/messages`：会话消息
- `GET /api/settings` / `POST /api/settings`：读取/更新 AI 设置
- `GET /api/providers`：预设 Provider 列表（前端“AI 设置”面板使用）

## 项目结构

- `main.go`：后端入口
- `routes/`：路由
- `handlers/`：HTTP 处理器（包含 SSE）
- `services/ai_service.go`：LLM 调用与流式回调
- `database/` + `models/`：Gorm + SQLite 表结构
- `web/`：前端（Vite + React + Tailwind）

## 生产构建

- 前端：`cd web && npm run build`（产物在 `web/dist`）
- 后端：`go build -o app .`

本仓库当前不自动由 Go 服务托管 `web/dist`；你可以用 Nginx/静态文件服务器托管前端，后端独立部署即可。

## 安全提示

- 不要提交 `.env`（已在 `.gitignore` 中忽略）；如果你曾把真实 Key 推到远端，请立即轮换。

