import { useState, useEffect } from 'react'
import { Power, BellRing, Minimize2, Github, BookOpen, HardDrive, Loader2, Save, Check } from 'lucide-react'
import { GetConfig, SaveConfig, GetDiskUsage } from '../../wailsjs/go/main/App'
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

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

export default function SettingsPage() {
  const [appConfig, setAppConfig] = useState<config.Config | null>(null)
  const [diskUsage, setDiskUsage] = useState<{ total: number; used: number; free: number } | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [saveSuccess, setSaveSuccess] = useState(false)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setIsLoading(true)
    try {
      const [cfg, usage] = await Promise.all([
        GetConfig(),
        GetDiskUsage()
      ])
      setAppConfig(cfg)
      setDiskUsage(usage)
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
      icon: BellRing,
      label: '定期扫描提醒',
      description: '定期提醒清理垃圾文件',
      key: 'scanReminder' as const,
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

  const usedPercent = diskUsage ? (diskUsage.used / diskUsage.total) * 100 : 0

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
      <div className="flex gap-6">
        {/* Left Column */}
        <div className="flex-1 space-y-6">
          {/* General Settings */}
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

        {/* Right Column */}
        <div className="w-80 space-y-6">
          {/* About Card */}
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
                href="https://github.com"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 px-4 py-2 bg-bg-hover rounded-lg hover:bg-border transition-colors"
              >
                <Github size={16} className="text-text-secondary" />
                <span className="text-text-primary text-sm">GitHub</span>
              </a>
              <a
                href="#"
                className="flex items-center gap-2 px-4 py-2 bg-bg-hover rounded-lg hover:bg-border transition-colors"
              >
                <BookOpen size={16} className="text-text-secondary" />
                <span className="text-text-primary text-sm">文档</span>
              </a>
            </div>
          </div>

          {/* Storage Stats */}
          {diskUsage && (
            <div className="bg-bg-card border border-border rounded-xl p-6 space-y-4">
              <div className="flex items-center gap-2">
                <HardDrive size={18} className="text-accent-blue" />
                <h3 className="font-primary text-base font-semibold text-text-primary">
                  存储统计
                </h3>
              </div>
              <div className="space-y-3">
                <div className="flex justify-between text-sm">
                  <span className="text-text-secondary">已使用</span>
                  <span className="text-text-primary font-medium">
                    {formatBytes(diskUsage.used)}
                  </span>
                </div>
                <div className="w-full h-2 bg-border rounded-full overflow-hidden">
                  <div
                    className="h-full rounded-full"
                    style={{
                      width: `${usedPercent}%`,
                      background: usedPercent > 90 
                        ? '#EF4444' 
                        : usedPercent > 70 
                          ? '#F59E0B' 
                          : 'linear-gradient(90deg, #3B82F6 0%, #22C55E 100%)',
                    }}
                  />
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-text-secondary">可用空间</span>
                  <span className="text-accent-green font-medium">
                    {formatBytes(diskUsage.free)}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-text-secondary">总容量</span>
                  <span className="text-text-primary font-medium">
                    {formatBytes(diskUsage.total)}
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
