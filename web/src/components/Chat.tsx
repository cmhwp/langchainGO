import { useState, useRef, useEffect, type CSSProperties } from 'react'
import ReactMarkdown, { type Components } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Send, Loader2, MessageSquare, Settings as SettingsIcon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Settings } from './Settings'

interface Message {
  id: number
  role: 'user' | 'assistant'
  content: string
  created_at: string
}

interface Conversation {
  id: number
  title: string
  created_at: string
}

const markdownComponents: Components = {
  code({ className, children, ...props }) {
    const match = /language-(\w+)/.exec(className || '')
    const code = String(children).replace(/\n$/, '')

    if (match) {
      return (
        <SyntaxHighlighter
          style={oneDark as Record<string, CSSProperties>}
          language={match[1]}
          PreTag="div"
          className="rounded-md overflow-x-auto"
          customStyle={{ margin: 0, padding: '0.75rem' }}
        >
          {code}
        </SyntaxHighlighter>
      )
    }

    return (
      <code className={className} {...props}>
        {children}
      </code>
    )
  },
}

export function Chat() {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [conversationId, setConversationId] = useState<number | null>(null)
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [streamingContent, setStreamingContent] = useState('')
  const [settingsOpen, setSettingsOpen] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages, streamingContent])

  useEffect(() => {
    loadConversations()
  }, [])

  const loadConversations = async () => {
    try {
      const res = await fetch('/api/conversations')
      const data = await res.json()
      setConversations(data.conversations || [])
    } catch (err) {
      console.error('Failed to load conversations:', err)
    }
  }

  const loadConversation = async (id: number) => {
    try {
      const res = await fetch(`/api/conversations/${id}/messages`)
      const data = await res.json()
      setMessages(data.messages || [])
      setConversationId(id)
    } catch (err) {
      console.error('Failed to load conversation:', err)
    }
  }

  const startNewChat = () => {
    setMessages([])
    setConversationId(null)
    setStreamingContent('')
  }

  const sendMessage = async () => {
    if (!input.trim() || loading) return

    const userMessage = input.trim()
    setInput('')
    setLoading(true)
    setStreamingContent('')

    // 添加用户消息到界面
    const tempUserMsg: Message = {
      id: Date.now(),
      role: 'user',
      content: userMessage,
      created_at: new Date().toISOString(),
    }
    setMessages((prev) => [...prev, tempUserMsg])

    try {
      const res = await fetch('/api/chat/stream', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          conversation_id: conversationId || 0,
          message: userMessage,
        }),
      })

      if (!res.ok) {
        throw new Error('请求失败')
      }

      const reader = res.body?.getReader()
      const decoder = new TextDecoder()
      let fullContent = ''
      let newConversationId = conversationId
      let buffer = ''

      if (!reader) {
        throw new Error('无法读取响应流')
      }

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() ?? ''

        for (const line of lines) {
          const trimmedLine = line.trimEnd()
          if (trimmedLine.startsWith('data: ')) {
            try {
              const data = JSON.parse(trimmedLine.slice(6))

              if (data.type === 'start' && data.conversation_id) {
                newConversationId = data.conversation_id
                if (!conversationId) {
                  setConversationId(data.conversation_id)
                }
              } else if (data.type === 'content') {
                fullContent += data.content
                setStreamingContent(fullContent)
              } else if (data.type === 'error') {
                throw new Error(data.error)
              } else if (data.type === 'done') {
                // 流结束，添加完整消息
                const assistantMsg: Message = {
                  id: Date.now() + 1,
                  role: 'assistant',
                  content: fullContent,
                  created_at: new Date().toISOString(),
                }
                setMessages((prev) => [...prev, assistantMsg])
                setStreamingContent('')

                // 刷新会话列表
                if (newConversationId !== conversationId) {
                  loadConversations()
                }
              }
            } catch (e) {
              // 忽略解析错误
              if (e instanceof Error && e.message !== 'Unexpected end of JSON input') {
                console.error('Parse error:', e)
              }
            }
          }
        }
      }
    } catch (err) {
      console.error('Failed to send message:', err)
      const errorMsg: Message = {
        id: Date.now() + 1,
        role: 'assistant',
        content: `错误: ${err instanceof Error ? err.message : '发送失败'}`,
        created_at: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, errorMsg])
      setStreamingContent('')
    } finally {
      setLoading(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  return (
    <div className="flex h-screen bg-gray-50">
      {/* 侧边栏 */}
      <div className="w-64 bg-white border-r border-gray-200 flex flex-col">
        <div className="p-4 border-b border-gray-200">
          <Button onClick={startNewChat} className="w-full">
            <MessageSquare className="w-4 h-4 mr-2" />
            新对话
          </Button>
        </div>
        <div className="flex-1 overflow-y-auto p-2">
          {conversations.map((conv) => (
            <button
              key={conv.id}
              onClick={() => loadConversation(conv.id)}
              className={cn(
                'w-full text-left p-3 rounded-lg mb-1 text-sm truncate transition-colors',
                conversationId === conv.id
                  ? 'bg-gray-100 text-gray-900'
                  : 'hover:bg-gray-50 text-gray-600'
              )}
            >
              {conv.title || '新对话'}
            </button>
          ))}
        </div>
        {/* 设置按钮 */}
        <div className="p-4 border-t border-gray-200">
          <Button variant="outline" onClick={() => setSettingsOpen(true)} className="w-full">
            <SettingsIcon className="w-4 h-4 mr-2" />
            AI 设置
          </Button>
        </div>
      </div>

      {/* 主聊天区域 */}
      <div className="flex-1 flex flex-col">
        <Card className="flex-1 m-4 flex flex-col overflow-hidden">
          <CardHeader className="border-b">
            <CardTitle className="text-lg">LangChainGo Chat</CardTitle>
          </CardHeader>
          <CardContent className="flex-1 overflow-y-auto p-4 space-y-4">
            {messages.length === 0 && !streamingContent ? (
              <div className="flex items-center justify-center h-full text-gray-400">
                开始一个新对话吧
              </div>
            ) : (
              <>
                {messages.map((msg) => (
                  <div
                    key={msg.id}
                    className={cn(
                      'flex',
                      msg.role === 'user' ? 'justify-end' : 'justify-start'
                    )}
                  >
                    <div
                      className={cn(
                        'max-w-[70%] rounded-lg px-4 py-2',
                        msg.role === 'user'
                          ? 'bg-gray-900 text-white'
                          : 'bg-gray-100 text-gray-900'
                      )}
                    >
                      {msg.role === 'assistant' ? (
                        <div className="prose prose-sm max-w-none prose-code:text-pink-600 prose-code:bg-gray-200 prose-code:px-1 prose-code:rounded prose-pre:bg-gray-900 prose-pre:text-gray-100 prose-pre:px-3 prose-pre:py-2 prose-pre:rounded-md prose-pre:overflow-x-auto prose-pre:whitespace-pre-wrap prose-ul:my-2 prose-ol:my-2 prose-li:my-1 prose-hr:my-4">
                          <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
                            {msg.content}
                          </ReactMarkdown>
                        </div>
                      ) : (
                        <span className="whitespace-pre-wrap">{msg.content}</span>
                      )}
                    </div>
                  </div>
                ))}
                {/* 流式输出中的内容 */}
                {streamingContent && (
                  <div className="flex justify-start">
                    <div className="max-w-[70%] rounded-lg px-4 py-2 bg-gray-100 text-gray-900">
                      <div className="prose prose-sm max-w-none prose-code:text-pink-600 prose-code:bg-gray-200 prose-code:px-1 prose-code:rounded prose-pre:bg-gray-900 prose-pre:text-gray-100 prose-pre:px-3 prose-pre:py-2 prose-pre:rounded-md prose-pre:overflow-x-auto prose-pre:whitespace-pre-wrap prose-ul:my-2 prose-ol:my-2 prose-li:my-1 prose-hr:my-4">
                        <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
                          {streamingContent}
                        </ReactMarkdown>
                      </div>
                    </div>
                  </div>
                )}
              </>
            )}
            {loading && !streamingContent && (
              <div className="flex justify-start">
                <div className="bg-gray-100 rounded-lg px-4 py-2">
                  <Loader2 className="w-4 h-4 animate-spin" />
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </CardContent>
          <div className="p-4 border-t">
            <div className="flex gap-2">
              <Input
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="输入消息..."
                disabled={loading}
                className="flex-1"
              />
              <Button onClick={sendMessage} disabled={loading || !input.trim()}>
                {loading ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Send className="w-4 h-4" />
                )}
              </Button>
            </div>
          </div>
        </Card>
      </div>

      {/* 设置面板 */}
      <Settings open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </div>
  )
}
