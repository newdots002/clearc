import { useState, useEffect } from 'react'
import { Power, RefreshCw, Minimize2, Github, BookOpen, Loader2, Save, Check } from 'lucide-react'
import { GetConfig, SaveConfig } from '../../wailsjs/go/main/App'
import { config } from '../../wailsjs/go/models'

interface ToggleProps {
  enabled: boolean
  onChange: (enabled: boolean) => void
}

function Toggle({ enabled, onChange }: ToggleProps) {
  return (
    <button
      onClick={() => onChange(!enabled)}
      className={`relative w-11 h-6 rounded-full transition-colors ${
        enabled ? 'bg-accent-blue' : 'bg-border'
      }`}
    >
      <div
        className={`absolute top-1 w-4 h-4 bg-white rounded-full transition-transform ${
          enabled ? 'left-6' : 'left-1'
        }`}
      />
    </button>
  )
}

export default function SettingsPage() {
  const [appConfig, setAppConfig] = useState<config.Config | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [saveSuccess, setSaveSuccess] = useState(false)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setIsLoading(true)
    try {
      const cfg = await GetConfig()
      setAppConfig(cfg)
    } catch (error) {
      console.error('Failed to load config:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const updateConfig = (key: keyof config.Config, value: any) => {
    if (!appConfig) return
    setAppConfig({ ...appConfig, [key]: value })
    setSaveSuccess(false)
  }

  const handleSave = async () => {
    if (!appConfig) return
    
    setIsSaving(true)
    try {
      await SaveConfig(appConfig)
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    } catch (error) {
      console.error('Failed to save config:', error)
    } finally {
      setIsSaving(false)
    }
  }

  const generalSettings = [
    {
      icon: Power,
      label: '开机自启动',
      description: '系统启动时自动运行 ClearC',
      key: 'startOnBoot' as const,
    },
    {
      icon: Minimize2,
      label: '最小化到托盘',
      description: '关闭窗口时最小化到系统托盘',
      key: 'minimizeToTray' as const,
    },
    {
      icon: RefreshCw,
      label: '定期扫描分析',
      description: '后台定期分析重点目录，加速下次查看',
      key: 'autoAnalyze' as const,
    },
  ]

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Loader2 className="w-8 h-8 animate-spin text-accent-blue" />
      </div>
    )
  }

  if (!appConfig) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-text-secondary">加载配置失败</p>
      </div>
    )
  }

  return (
    <div className="p-8 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-primary text-3xl font-semibold text-text-primary">
            设置
          </h1>
          <p className="text-text-secondary text-sm mt-1">
            配置应用程序行为
          </p>
        </div>
        <button
          onClick={handleSave}
          disabled={isSaving}
          className={`flex items-center gap-2 px-5 py-2.5 rounded-lg transition-colors ${
            saveSuccess 
              ? 'bg-accent-green text-white' 
              : 'bg-accent-blue text-white hover:bg-blue-600'
          } disabled:opacity-50`}
        >
          {isSaving ? (
            <Loader2 size={16} className="animate-spin" />
          ) : saveSuccess ? (
            <Check size={16} />
          ) : (
            <Save size={16} />
          )}
          <span className="font-secondary text-sm font-medium">
            {isSaving ? '保存中...' : saveSuccess ? '已保存' : '保存设置'}
          </span>
        </button>
      </div>

      {/* Main Content */}
      <div className="flex gap-6 items-start">
        {/* Left Column - General Settings */}
        <div className="flex-1">
          <div className="bg-bg-card border border-border rounded-xl p-6 space-y-4">
            <h2 className="font-primary text-lg font-semibold text-text-primary">
              常规设置
            </h2>
            <div className="space-y-4">
              {generalSettings.map((setting) => {
                const Icon = setting.icon
                return (
                  <div
                    key={setting.key}
                    className="flex items-center justify-between py-2"
                  >
                    <div className="flex items-center gap-3">
                      <div className="w-9 h-9 bg-bg-hover rounded-lg flex items-center justify-center">
                        <Icon size={18} className="text-text-secondary" />
                      </div>
                      <div>
                        <p className="font-secondary text-sm font-medium text-text-primary">
                          {setting.label}
                        </p>
                        <p className="text-text-muted text-xs">
                          {setting.description}
                        </p>
                      </div>
                    </div>
                    <Toggle
                      enabled={appConfig[setting.key] as boolean}
                      onChange={(value) => updateConfig(setting.key, value)}
                    />
                  </div>
                )
              })}
            </div>
          </div>
        </div>

        {/* Right Column - About Card */}
        <div className="w-80">
          <div className="bg-bg-card border border-border rounded-xl p-6 space-y-4">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 bg-accent-blue rounded-xl flex items-center justify-center">
                <span className="text-white font-bold text-lg">C</span>
              </div>
              <div>
                <h3 className="font-primary text-lg font-semibold text-text-primary">
                  ClearC
                </h3>
                <p className="text-text-muted text-sm">版本 1.0.0</p>
              </div>
            </div>
            <p className="text-text-secondary text-sm leading-relaxed">
              跨平台系统清理工具，专为开发者设计。
              快速清理开发垃圾，释放磁盘空间。
            </p>
            <div className="flex gap-3">
              <a
                href="https://github.com/newdots002/clearc"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 px-4 py-2 bg-bg-hover rounded-lg hover:bg-border transition-colors"
              >
                <Github size={16} className="text-text-secondary" />
                <span className="text-text-primary text-sm">GitHub</span>
              </a>
              <a
                href="https://clearc.top"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 px-4 py-2 bg-bg-hover rounded-lg hover:bg-border transition-colors"
              >
                <BookOpen size={16} className="text-text-secondary" />
                <span className="text-text-primary text-sm">官网</span>
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
