import { useState, useEffect } from 'react'
import { HardDrive, CheckCircle, ArrowRight, Loader2, FolderSearch } from 'lucide-react'
import { GetDiskUsage } from '../../wailsjs/go/main/App'
import type { Page, DiskUsage } from '../types'

interface DashboardProps {
  onNavigate: (page: Page) => void
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

export default function Dashboard({ onNavigate }: DashboardProps) {
  const [diskUsage, setDiskUsage] = useState<DiskUsage | null>(null)
  const [isLoadingDisk, setIsLoadingDisk] = useState(true)

  useEffect(() => {
    loadDiskUsage()
  }, [])

  const loadDiskUsage = async () => {
    try {
      const usage = await GetDiskUsage()
      setDiskUsage(usage)
    } catch (error) {
      console.error('Failed to load disk usage:', error)
    } finally {
      setIsLoadingDisk(false)
    }
  }

  const usedPercent = diskUsage ? (diskUsage.used / diskUsage.total) * 100 : 0

  const stats = [
    {
      label: '总磁盘空间',
      value: diskUsage ? formatBytes(diskUsage.total) : '-',
      icon: HardDrive,
      color: '#3B82F6',
      bgColor: '#3B82F620',
      loading: isLoadingDisk,
    },
    {
      label: '可用空间',
      value: diskUsage ? formatBytes(diskUsage.free) : '-',
      icon: CheckCircle,
      color: '#22C55E',
      bgColor: '#22C55E20',
      loading: isLoadingDisk,
    },
    {
      label: '已使用',
      value: diskUsage ? formatBytes(diskUsage.used) : '-',
      icon: HardDrive,
      color: usedPercent > 80 ? '#EF4444' : '#F59E0B',
      bgColor: usedPercent > 80 ? '#EF444420' : '#F59E0B20',
      loading: isLoadingDisk,
    },
  ]

  return (
    <div className="p-8 space-y-6">
      {/* Header */}
      <div>
        <h1 className="font-primary text-3xl font-semibold text-text-primary">
          仪表盘
        </h1>
        <p className="text-text-secondary text-sm mt-1">
          查看磁盘使用情况，快速清理系统垃圾
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-3 gap-4">
        {stats.map((stat, index) => {
          const Icon = stat.icon
          return (
            <div
              key={index}
              className="bg-bg-card border border-border rounded-xl p-5 space-y-3"
            >
              <div
                className="w-10 h-10 rounded-lg flex items-center justify-center"
                style={{ backgroundColor: stat.bgColor }}
              >
                <Icon size={20} style={{ color: stat.color }} />
              </div>
              <div>
                {stat.loading ? (
                  <div className="flex items-center gap-2">
                    <Loader2 size={20} className="animate-spin" style={{ color: stat.color }} />
                  </div>
                ) : (
                  <p
                    className="font-primary text-2xl font-semibold"
                    style={{ color: stat.color }}
                  >
                    {stat.value}
                  </p>
                )}
                <p className="text-text-secondary text-sm">{stat.label}</p>
              </div>
            </div>
          )
        })}
      </div>

      {/* Disk Usage Bar */}
      {diskUsage && (
        <div className="bg-bg-card border border-border rounded-xl p-6 space-y-4">
          <h2 className="font-primary text-lg font-semibold text-text-primary">
            磁盘使用情况
          </h2>
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-text-secondary">
                已使用 {formatBytes(diskUsage.used)} / {formatBytes(diskUsage.total)}
              </span>
              <span className={`font-medium ${usedPercent > 80 ? 'text-accent-red' : 'text-accent-blue'}`}>
                {usedPercent.toFixed(1)}%
              </span>
            </div>
            <div className="w-full h-3 bg-border rounded-full overflow-hidden">
              <div
                className="h-full rounded-full transition-all"
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
          </div>
        </div>
      )}

      {/* Quick Actions */}
      <div className="space-y-4">
        <h2 className="font-primary text-lg font-semibold text-text-primary">
          快速操作
        </h2>
        {/* Disk Analyzer Card */}
        <button
          onClick={() => onNavigate('analyzer')}
          className="w-full bg-bg-card border border-border rounded-xl p-6 text-left hover:border-accent-blue transition-colors group"
        >
          <div className="flex items-start justify-between">
            <div className="space-y-3">
              <div className="w-12 h-12 bg-accent-blue/20 rounded-lg flex items-center justify-center">
                <FolderSearch size={24} className="text-accent-blue" />
              </div>
              <div>
                <h3 className="font-primary text-lg font-semibold text-text-primary">
                  磁盘分析
                </h3>
                <p className="text-text-secondary text-sm mt-1">
                  分析磁盘空间占用，智能识别可清理的目录（缓存、临时文件、开发依赖等）
                </p>
              </div>
            </div>
            <ArrowRight size={20} className="text-text-muted group-hover:text-accent-blue transition-colors" />
          </div>
        </button>
      </div>

      {/* Tips */}
      <div className="bg-accent-blue/10 border border-accent-blue/30 rounded-xl p-4">
        <p className="text-accent-blue text-sm">
          <strong>提示：</strong>使用"磁盘分析"功能可以查看每个目录的大小和类型，系统会智能识别哪些目录可以安全删除。
        </p>
      </div>
    </div>
  )
}
