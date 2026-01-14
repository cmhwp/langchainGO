import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { X, Save, Loader2, Eye, EyeOff } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Provider {
  name: string
  base_url: string
  models: string[]
}

interface SettingsProps {
  open: boolean
  onClose: () => void
}

export function Settings({ open, onClose }: SettingsProps) {
  const [providers, setProviders] = useState<Provider[]>([])
  const [selectedProvider, setSelectedProvider] = useState<string>('')
  const [baseUrl, setBaseUrl] = useState('')
  const [model, setModel] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [showApiKey, setShowApiKey] = useState(false)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  useEffect(() => {
    if (open) {
      loadProviders()
      loadSettings()
    }
  }, [open])

  const loadProviders = async () => {
    try {
      const res = await fetch('/api/providers')
      const data = await res.json()
      setProviders(data.providers || [])
    } catch (err) {
      console.error('Failed to load providers:', err)
    }
  }

  const loadSettings = async () => {
    setLoading(true)
    try {
      const res = await fetch('/api/settings')
      const data = await res.json()
      setSelectedProvider(data.provider || '')
      setBaseUrl(data.base_url || '')
      setModel(data.model || '')
      // API Key 是掩码的，不要覆盖用户输入
    } catch (err) {
      console.error('Failed to load settings:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleProviderChange = (providerName: string) => {
    setSelectedProvider(providerName)
    const provider = providers.find((p) => p.name === providerName)
    if (provider) {
      setBaseUrl(provider.base_url)
      if (provider.models.length > 0) {
        setModel(provider.models[0])
      }
    }
  }

  const handleSave = async () => {
    if (!baseUrl || !model || !apiKey) {
      setMessage({ type: 'error', text: '请填写所有必填字段' })
      return
    }

    setSaving(true)
    setMessage(null)

    try {
      const res = await fetch('/api/settings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          provider: selectedProvider,
          base_url: baseUrl,
          model: model,
          api_key: apiKey,
        }),
      })

      const data = await res.json()

      if (!res.ok) {
        throw new Error(data.error || '保存失败')
      }

      setMessage({ type: 'success', text: '设置已保存' })
      setApiKey('') // 清空 API Key 输入
      setTimeout(() => {
        onClose()
      }, 1000)
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : '保存失败' })
    } finally {
      setSaving(false)
    }
  }

  if (!open) return null

  const currentProvider = providers.find((p) => p.name === selectedProvider)

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <Card className="w-full max-w-lg mx-4">
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>AI 设置</CardTitle>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="w-4 h-4" />
          </Button>
        </CardHeader>
        <CardContent className="space-y-4">
          {loading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="w-6 h-6 animate-spin" />
            </div>
          ) : (
            <>
              {/* 提供商选择 */}
              <div className="space-y-2">
                <label className="text-sm font-medium">提供商</label>
                <div className="grid grid-cols-2 gap-2">
                  {providers.map((provider) => (
                    <button
                      key={provider.name}
                      onClick={() => handleProviderChange(provider.name)}
                      className={cn(
                        'p-2 text-sm rounded-lg border transition-colors text-left',
                        selectedProvider === provider.name
                          ? 'border-gray-900 bg-gray-100'
                          : 'border-gray-200 hover:border-gray-300'
                      )}
                    >
                      {provider.name}
                    </button>
                  ))}
                </div>
              </div>

              {/* Base URL */}
              <div className="space-y-2">
                <label className="text-sm font-medium">Base URL</label>
                <Input
                  value={baseUrl}
                  onChange={(e) => setBaseUrl(e.target.value)}
                  placeholder="https://api.openai.com/v1"
                />
              </div>

              {/* 模型选择 */}
              <div className="space-y-2">
                <label className="text-sm font-medium">模型</label>
                {currentProvider && currentProvider.models.length > 0 ? (
                  <div className="space-y-2">
                    <select
                      value={model}
                      onChange={(e) => setModel(e.target.value)}
                      className="w-full h-9 rounded-md border border-gray-200 bg-transparent px-3 py-1 text-sm"
                    >
                      {currentProvider.models.map((m) => (
                        <option key={m} value={m}>
                          {m}
                        </option>
                      ))}
                    </select>
                    <Input
                      value={model}
                      onChange={(e) => setModel(e.target.value)}
                      placeholder="或输入自定义模型名称"
                    />
                  </div>
                ) : (
                  <Input
                    value={model}
                    onChange={(e) => setModel(e.target.value)}
                    placeholder="输入模型名称"
                  />
                )}
              </div>

              {/* API Key */}
              <div className="space-y-2">
                <label className="text-sm font-medium">API Key</label>
                <div className="relative">
                  <Input
                    type={showApiKey ? 'text' : 'password'}
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    placeholder="输入新的 API Key"
                    className="pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowApiKey(!showApiKey)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                  >
                    {showApiKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
                <p className="text-xs text-gray-500">每次保存需要重新输入 API Key</p>
              </div>

              {/* 消息提示 */}
              {message && (
                <div
                  className={cn(
                    'p-3 rounded-lg text-sm',
                    message.type === 'success' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                  )}
                >
                  {message.text}
                </div>
              )}

              {/* 保存按钮 */}
              <Button onClick={handleSave} disabled={saving} className="w-full">
                {saving ? (
                  <Loader2 className="w-4 h-4 animate-spin mr-2" />
                ) : (
                  <Save className="w-4 h-4 mr-2" />
                )}
                保存设置
              </Button>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
