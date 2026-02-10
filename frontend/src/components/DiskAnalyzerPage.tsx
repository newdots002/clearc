import { useState, useEffect, useCallback, useRef } from 'react'
import { 
  HardDrive, Folder, File, ChevronRight, ChevronDown, 
  Trash2, Loader2, Shield, RefreshCw, AlertTriangle,
  CheckSquare, Square, Filter, Eye, EyeOff, Zap, X
} from 'lucide-react'
import { AnalyzeDrive, ExpandNode, DeletePaths, GetAnalyzeProgress, GetAnalyzeStatus, GetWhitelistDirs, SetAnalyzerMinSize, GetAnalyzerMinSize, AnalyzeQuickScan, GetSystemDrive } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

interface FileNode {
  name: string
  path: string
  size: number
  isDir: boolean
  children?: FileNode[]
  isProtected: boolean
  fileCount: number
  dirType: string
  recommendation: string
  description: string
}

const dirTypeLabels: Record<string, { label: string; color: string }> = {
  system: { label: '系统', color: 'text-red-500' },
  application: { label: '应用', color: 'text-blue-500' },
  cache: { label: '缓存', color: 'text-green-500' },
  temp: { label: '临时', color: 'text-green-500' },
  dev_cache: { label: '开发缓存', color: 'text-emerald-500' },
  build_output: { label: '构建输出', color: 'text-emerald-500' },
  user_data: { label: '用户数据', color: 'text-purple-500' },
  downloads: { label: '下载', color: 'text-orange-500' },
  logs: { label: '日志', color: 'text-yellow-500' },
  backup: { label: '备份', color: 'text-cyan-500' },
  unknown: { label: '未知', color: 'text-gray-500' },
}

const recommendationLabels: Record<string, { label: string; color: string; bgColor: string }> = {
  safe: { label: '可删除', color: 'text-green-400', bgColor: 'bg-green-500/20' },
  caution: { label: '谨慎', color: 'text-yellow-400', bgColor: 'bg-yellow-500/20' },
  keep: { label: '保留', color: 'text-blue-400', bgColor: 'bg-blue-500/20' },
  never: { label: '勿删', color: 'text-red-400', bgColor: 'bg-red-500/20' },
}

function formatBytes(bytes: number): string {
  if (bytes < 0) return '计算中...'
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatNumber(num: number): string {
  return num.toLocaleString('zh-CN')
}

// Custom Confirm Dialog Component
interface ConfirmDialogProps {
  isOpen: boolean
  title: string
  message: string
  detail?: string
  confirmText?: string
  cancelText?: string
  onConfirm: () => void
  onCancel: () => void
  variant?: 'danger' | 'warning' | 'info'
}

function ConfirmDialog({
  isOpen,
  title,
  message,
  detail,
  confirmText = '确定',
  cancelText = '取消',
  onConfirm,
  onCancel,
  variant = 'danger'
}: ConfirmDialogProps) {
  if (!isOpen) return null

  const variantStyles = {
    danger: {
      icon: <Trash2 className="w-6 h-6 text-red-400" />,
      iconBg: 'bg-red-500/20',
      confirmBtn: 'bg-red-500 hover:bg-red-600 text-white'
    },
    warning: {
      icon: <AlertTriangle className="w-6 h-6 text-yellow-400" />,
      iconBg: 'bg-yellow-500/20',
      confirmBtn: 'bg-yellow-500 hover:bg-yellow-600 text-white'
    },
    info: {
      icon: <AlertTriangle className="w-6 h-6 text-blue-400" />,
      iconBg: 'bg-blue-500/20',
      confirmBtn: 'bg-blue-500 hover:bg-blue-600 text-white'
    }
  }

  const styles = variantStyles[variant]

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onCancel}
      />
      
      {/* Dialog */}
      <div className="relative bg-bg-secondary border border-border-primary rounded-xl shadow-2xl w-full max-w-md mx-4 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-border-primary">
          <div className="flex items-center gap-3">
            <div className={`p-2 rounded-lg ${styles.iconBg}`}>
              {styles.icon}
            </div>
            <h3 className="font-semibold text-text-primary text-lg">{title}</h3>
          </div>
          <button
            onClick={onCancel}
            className="p-1.5 rounded-lg hover:bg-bg-tertiary text-text-secondary hover:text-text-primary transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>
        
        {/* Content */}
        <div className="px-5 py-4">
          <p className="text-text-primary">{message}</p>
          {detail && (
            <p className="mt-2 text-sm text-text-secondary">{detail}</p>
          )}
        </div>
        
        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-5 py-4 bg-bg-tertiary/50 border-t border-border-primary">
          <button
            onClick={onCancel}
            className="px-4 py-2 rounded-lg bg-bg-tertiary hover:bg-border-primary text-text-secondary hover:text-text-primary transition-colors font-medium"
          >
            {cancelText}
          </button>
          <button
            onClick={onConfirm}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${styles.confirmBtn}`}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  )
}

interface TreeNodeProps {
  node: FileNode
  level: number
  selectedPaths: Set<string>
  expandedPaths: Set<string>
  onToggleSelect: (path: string, node: FileNode) => void
  onToggleExpand: (path: string) => void
  onLoadChildren: (path: string) => Promise<void>
  maxSize: number
  showSystemDirs: boolean
}

function TreeNode({ 
  node, 
  level, 
  selectedPaths, 
  expandedPaths, 
  onToggleSelect, 
  onToggleExpand,
  onLoadChildren,
  maxSize,
  showSystemDirs
}: TreeNodeProps) {
  const [isLoading, setIsLoading] = useState(false)
  const isExpanded = expandedPaths.has(node.path)
  const isSelected = selectedPaths.has(node.path)
  const canExpand = node.isDir

  // Filter children to hide system/never-delete dirs if showSystemDirs is false
  const visibleChildren = node.children?.filter(child => 
    showSystemDirs || (!child.isProtected && child.recommendation !== 'never')
  ) || []
  const hasChildren = node.isDir && visibleChildren.length > 0

  const sizePercent = maxSize > 0 ? (node.size / maxSize) * 100 : 0

  // Hide this node if it's a system dir and showSystemDirs is false (except root)
  if (!showSystemDirs && level > 0 && (node.isProtected || node.recommendation === 'never')) {
    return null
  }

  const handleExpand = async () => {
    if (!node.isDir) return
    
    if (!isExpanded && (!node.children || node.children.length === 0)) {
      setIsLoading(true)
      await onLoadChildren(node.path)
      setIsLoading(false)
    }
    onToggleExpand(node.path)
  }

  const getBarColor = () => {
    if (node.isProtected) return 'bg-gray-500'
    if (node.recommendation === 'safe') return 'bg-green-500'
    if (node.recommendation === 'caution') return 'bg-yellow-500'
    if (node.recommendation === 'keep') return 'bg-blue-500'
    if (node.recommendation === 'never') return 'bg-red-500'
    if (sizePercent > 20) return 'bg-accent-red'
    if (sizePercent > 10) return 'bg-accent-yellow'
    return 'bg-accent-blue'
  }

  const typeInfo = node.dirType ? dirTypeLabels[node.dirType] : dirTypeLabels.unknown
  const recInfo = node.recommendation ? recommendationLabels[node.recommendation] : recommendationLabels.caution

  return (
    <div>
      <div 
        className={`flex items-center gap-1.5 py-1 px-2 hover:bg-bg-hover rounded cursor-pointer group ${
          isSelected ? 'bg-accent-blue/10' : ''
        }`}
        style={{ paddingLeft: `${level * 16 + 4}px` }}
      >
        {/* Expand/Collapse button */}
        <button 
          onClick={handleExpand}
          className="w-5 h-5 flex items-center justify-center flex-shrink-0"
          disabled={!canExpand}
        >
          {isLoading ? (
            <Loader2 size={12} className="animate-spin text-text-muted" />
          ) : canExpand ? (
            isExpanded ? (
              <ChevronDown size={12} className="text-text-secondary" />
            ) : (
              <ChevronRight size={12} className="text-text-secondary" />
            )
          ) : (
            <span className="w-3" />
          )}
        </button>

        {/* Checkbox */}
        <button
          onClick={() => onToggleSelect(node.path, node)}
          className="flex-shrink-0"
          disabled={node.isProtected || node.recommendation === 'never'}
        >
          {node.isProtected || node.recommendation === 'never' ? (
            <Shield size={14} className="text-gray-500" />
          ) : isSelected ? (
            <CheckSquare size={14} className="text-accent-blue" />
          ) : (
            <Square size={14} className="text-text-muted group-hover:text-text-secondary" />
          )}
        </button>

        {/* Icon */}
        <div className="flex-shrink-0">
          {node.isDir ? (
            <Folder size={14} className={node.isProtected ? 'text-gray-500' : 'text-accent-yellow'} />
          ) : (
            <File size={14} className="text-text-muted" />
          )}
        </div>

        {/* Name */}
        <span className={`min-w-0 flex-1 text-xs truncate ${
          node.isProtected ? 'text-gray-500' : 'text-text-primary'
        }`} title={node.description}>
          {node.name}
        </span>

        {/* Directory Type Tag */}
        {node.isDir && (
          <span className={`text-[10px] px-1 py-0.5 rounded ${typeInfo.color} bg-bg-secondary flex-shrink-0 w-14 text-center`}>
            {typeInfo.label}
          </span>
        )}

        {/* Recommendation Tag */}
        {node.isDir && (
          <span className={`text-[10px] px-1 py-0.5 rounded ${recInfo.color} ${recInfo.bgColor} flex-shrink-0 w-12 text-center`}>
            {recInfo.label}
          </span>
        )}

        {/* File count */}
        {node.isDir && (
          <span className="text-[10px] text-text-muted flex-shrink-0 w-14 text-right">
            {node.fileCount > 0 ? formatNumber(node.fileCount) : '-'}
          </span>
        )}

        {/* Size bar */}
        <div className="w-16 h-1.5 bg-border rounded-full overflow-hidden flex-shrink-0">
          <div 
            className={`h-full rounded-full ${getBarColor()}`}
            style={{ width: `${Math.max(sizePercent, 1)}%` }}
          />
        </div>

        {/* Size */}
        <span className={`text-xs font-medium w-16 text-right flex-shrink-0 ${
          node.isProtected ? 'text-gray-500' : 'text-text-primary'
        }`}>
          {formatBytes(node.size)}
        </span>
      </div>

      {/* Children */}
      {isExpanded && hasChildren && (
        <div>
          {visibleChildren.map((child) => (
            <TreeNode
              key={child.path}
              node={child}
              level={level + 1}
              selectedPaths={selectedPaths}
              expandedPaths={expandedPaths}
              onToggleSelect={onToggleSelect}
              onToggleExpand={onToggleExpand}
              onLoadChildren={onLoadChildren}
              maxSize={maxSize}
              showSystemDirs={showSystemDirs}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export default function DiskAnalyzerPage() {
  const [rootNode, setRootNode] = useState<FileNode | null>(null)
  const [isAnalyzing, setIsAnalyzing] = useState(false)
  const [isCalculatingSizes, setIsCalculatingSizes] = useState(false)
  const [progress, setProgress] = useState(0)
  const [status, setStatus] = useState('')
  const [selectedPaths, setSelectedPaths] = useState<Set<string>>(new Set())
  const [expandedPaths, setExpandedPaths] = useState<Set<string>>(new Set())
  const [isDeleting, setIsDeleting] = useState(false)
  const [deleteResult, setDeleteResult] = useState<{ size: number; files: number; errors: string[] } | null>(null)
  const [whitelistDirs, setWhitelistDirs] = useState<string[]>([])
  const [minSizeMB, setMinSizeMB] = useState(100) // Default 100MB minimum
  const [showSystemDirs, setShowSystemDirs] = useState(false) // Hide system dirs by default
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false) // Delete confirmation dialog
  const rootNodeRef = useRef<FileNode | null>(null)

  // Keep ref in sync with state
  useEffect(() => {
    rootNodeRef.current = rootNode
  }, [rootNode])

  // Listen for size updates from backend
  useEffect(() => {
    const handleSizeUpdate = (data: { path: string; size: number; fileCount: number }) => {
      setRootNode(prev => {
        if (!prev) return prev
        return updateNodeSize(prev, data.path, data.size, data.fileCount)
      })
    }

    const handleSizeComplete = () => {
      setIsCalculatingSizes(false)
      setStatus('分析完成')
      // Sort and filter after all sizes are calculated
      setRootNode(prev => {
        if (!prev) return prev
        return sortAndFilterNode(prev, minSizeMB * 1024 * 1024)
      })
    }

    const handleExpandComplete = (data: { path: string }) => {
      // Sort children of expanded node after sizes are calculated
      setRootNode(prev => {
        if (!prev) return prev
        return sortNodeChildren(prev, data.path, minSizeMB * 1024 * 1024)
      })
    }

    EventsOn('sizeUpdate', handleSizeUpdate)
    EventsOn('sizeCalculationComplete', handleSizeComplete)
    EventsOn('expandComplete', handleExpandComplete)

    return () => {
      EventsOff('sizeUpdate')
      EventsOff('sizeCalculationComplete')
      EventsOff('expandComplete')
    }
  }, [minSizeMB])

  useEffect(() => {
    loadWhitelist()
    loadMinSize()
  }, [])

  const loadWhitelist = async () => {
    try {
      const dirs = await GetWhitelistDirs()
      setWhitelistDirs(dirs)
    } catch (error) {
      console.error('Failed to load whitelist:', error)
    }
  }

  const loadMinSize = async () => {
    try {
      const size = await GetAnalyzerMinSize()
      if (size > 0) {
        setMinSizeMB(size)
      }
    } catch (error) {
      console.error('Failed to load min size:', error)
    }
  }

  const handleMinSizeChange = async (value: number) => {
    setMinSizeMB(value)
    try {
      await SetAnalyzerMinSize(value)
    } catch (error) {
      console.error('Failed to set min size:', error)
    }
  }

  const startAnalyze = async (quickScan = false) => {
    setIsAnalyzing(true)
    setIsCalculatingSizes(true)
    setProgress(0)
    setStatus(quickScan ? '快速扫描重点目录...' : '正在读取目录...')
    setRootNode(null)
    setSelectedPaths(new Set())
    setExpandedPaths(new Set())
    setDeleteResult(null)

    // Poll progress for size calculation
    const pollInterval = setInterval(async () => {
      try {
        const [prog, stat] = await Promise.all([
          GetAnalyzeProgress(),
          GetAnalyzeStatus()
        ])
        setProgress(prog)
        setStatus(stat)
      } catch (e) {
        // ignore
      }
    }, 500)

    try {
      // Pass skipSystem = !showSystemDirs (skip system dirs when they are hidden)
      const skipSystem = !showSystemDirs
      // Get system drive path (C:\ on Windows, / on Linux/macOS)
      const systemDrive = await GetSystemDrive()
      const result = quickScan 
        ? await AnalyzeQuickScan()
        : await AnalyzeDrive(systemDrive, skipSystem)
      setRootNode(result)
      setIsAnalyzing(false)
      // Auto-expand root
      if (result) {
        setExpandedPaths(new Set([result.path]))
      }
    } catch (error) {
      console.error('Analyze failed:', error)
      setStatus('分析失败: ' + String(error))
      setIsAnalyzing(false)
      setIsCalculatingSizes(false)
      clearInterval(pollInterval)
      return
    }

    // Keep polling until size calculation completes
    // The interval will be cleared when sizeCalculationComplete event is received
    setTimeout(() => {
      clearInterval(pollInterval)
    }, 300000) // Max 5 minutes timeout
  }

  const loadChildren = useCallback(async (path: string) => {
    try {
      const result = await ExpandNode(path)
      if (result && rootNode) {
        // Update the tree with new children
        setRootNode(prevRoot => {
          if (!prevRoot) return prevRoot
          return updateNodeChildren(prevRoot, path, result.children || [])
        })
      }
    } catch (error) {
      console.error('Failed to expand node:', error)
    }
  }, [rootNode])

  const updateNodeChildren = (node: FileNode, targetPath: string, children: FileNode[]): FileNode => {
    if (node.path === targetPath) {
      return { ...node, children }
    }
    if (node.children) {
      return {
        ...node,
        children: node.children.map(child => updateNodeChildren(child, targetPath, children))
      }
    }
    return node
  }

  const toggleSelect = (path: string, node: FileNode) => {
    if (node.isProtected || node.recommendation === 'never') return
    
    setSelectedPaths(prev => {
      const next = new Set(prev)
      if (next.has(path)) {
        next.delete(path)
      } else {
        next.add(path)
      }
      return next
    })
  }

  const toggleExpand = (path: string) => {
    setExpandedPaths(prev => {
      const next = new Set(prev)
      if (next.has(path)) {
        next.delete(path)
      } else {
        next.add(path)
      }
      return next
    })
  }

  const handleDeleteClick = () => {
    if (selectedPaths.size === 0) return
    setShowDeleteConfirm(true)
  }

  const handleDeleteConfirm = async () => {
    setShowDeleteConfirm(false)
    setIsDeleting(true)
    setDeleteResult(null)

    try {
      const paths = Array.from(selectedPaths)
      const result = await DeletePaths(paths)
      setDeleteResult({
        size: result.cleanedSize,
        files: result.cleanedFiles,
        errors: result.errors || []
      })
      
      // Remove deleted nodes from tree and update parent sizes
      const deletedPaths = new Set(paths)
      // Filter out paths that had errors (they weren't actually deleted)
      if (result.errors && result.errors.length > 0) {
        result.errors.forEach(err => {
          // Extract path from error message if possible
          const match = err.match(/: (.+)$/)
          if (match) {
            deletedPaths.delete(match[1])
          }
        })
      }
      
      setRootNode(prevRoot => {
        if (!prevRoot) return prevRoot
        return removeNodesAndUpdateSizes(prevRoot, deletedPaths)
      })
      
      // Also remove from expanded paths
      setExpandedPaths(prev => {
        const next = new Set(prev)
        deletedPaths.forEach(p => next.delete(p))
        return next
      })
      
      setSelectedPaths(new Set())
    } catch (error) {
      console.error('Delete failed:', error)
      setDeleteResult({
        size: 0,
        files: 0,
        errors: [String(error)]
      })
    } finally {
      setIsDeleting(false)
    }
  }

  const handleDeleteCancel = () => {
    setShowDeleteConfirm(false)
  }

  const selectedSize = rootNode ? calculateSelectedSize(rootNode, selectedPaths) : 0

  return (
    <div className="p-4 space-y-3 h-full flex flex-col">
      {/* Compact Header Bar */}
      <div className="flex items-center justify-between flex-shrink-0">
        <div className="flex items-center gap-4">
          <h1 className="font-primary text-xl font-semibold text-text-primary">
            磁盘分析
          </h1>
          {/* Inline Progress */}
          {(isAnalyzing || isCalculatingSizes) && (
            <div className="flex items-center gap-2 text-xs text-text-muted">
              <Loader2 size={12} className="animate-spin text-accent-blue" />
              <span className="max-w-48 truncate">{status}</span>
              <span className="text-accent-blue">{progress}%</span>
            </div>
          )}
        </div>
        <div className="flex items-center gap-2">
          {/* Show/Hide System Dirs Toggle */}
          <button
            onClick={() => setShowSystemDirs(!showSystemDirs)}
            className={`flex items-center gap-1 px-2 py-1.5 rounded-lg text-xs transition-colors ${
              showSystemDirs 
                ? 'bg-gray-500/20 text-gray-400' 
                : 'bg-bg-secondary text-text-muted hover:text-text-secondary'
            }`}
            title={showSystemDirs ? '隐藏系统目录' : '显示系统目录'}
          >
            {showSystemDirs ? <Eye size={14} /> : <EyeOff size={14} />}
            <span>系统</span>
          </button>
          {/* Min Size Filter - Compact */}
          <div className="flex items-center gap-1">
            <Filter size={12} className="text-text-muted" />
            <select
              value={minSizeMB}
              onChange={(e) => handleMinSizeChange(Number(e.target.value))}
              className="bg-bg-secondary border border-border rounded px-1.5 py-1 text-xs text-text-primary focus:outline-none focus:border-accent-blue [&>option]:bg-bg-secondary [&>option]:text-text-primary"
              style={{ colorScheme: 'dark' }}
            >
              <option value={0}>全部</option>
              <option value={50}>50MB</option>
              <option value={100}>100MB</option>
              <option value={500}>500MB</option>
              <option value={1024}>1GB</option>
            </select>
          </div>
          {/* Quick Scan Button */}
          <button
            onClick={() => startAnalyze(true)}
            disabled={isAnalyzing}
            className="flex items-center gap-1 px-2.5 py-1.5 bg-accent-green/20 text-accent-green rounded-lg hover:bg-accent-green/30 transition-colors disabled:opacity-50 text-xs"
            title="快速扫描 Users、缓存等重点目录"
          >
            <Zap size={14} />
            <span>快速</span>
          </button>
          {/* Full Scan Button */}
          <button
            onClick={() => startAnalyze(false)}
            disabled={isAnalyzing}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-accent-blue text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50 text-sm"
          >
            {isAnalyzing ? (
              <Loader2 size={14} className="animate-spin" />
            ) : (
              <RefreshCw size={14} />
            )}
            <span>{isAnalyzing ? '分析中' : '全盘扫描'}</span>
          </button>
        </div>
      </div>

      {/* Delete Result - Compact */}
      {deleteResult && (
        <div className={`px-3 py-2 rounded-lg border flex-shrink-0 text-sm ${
          deleteResult.errors.length > 0 
            ? 'bg-red-500/10 border-red-500/30' 
            : 'bg-green-500/10 border-green-500/30'
        }`}>
          <div className="flex items-center gap-2">
            {deleteResult.errors.length > 0 ? (
              <AlertTriangle className="text-accent-red" size={16} />
            ) : (
              <Trash2 className="text-accent-green" size={16} />
            )}
            <span className={deleteResult.errors.length > 0 ? 'text-accent-red' : 'text-accent-green'}>
              {deleteResult.errors.length > 0 ? '删除完成，但有错误 - ' : '删除成功 - '}
              已删除 {deleteResult.files} 个文件，释放 {formatBytes(deleteResult.size)}
            </span>
          </div>
        </div>
      )}

      {/* Tree View - Main Content Area */}
      <div className="flex-1 bg-bg-card border border-border rounded-lg overflow-hidden flex flex-col min-h-0">
        {/* Tree Header - Compact */}
        <div className="flex items-center gap-2 px-3 py-2 border-b border-border bg-bg-secondary text-xs text-text-muted">
          <span className="w-5" />
          <span className="w-4" />
          <span className="w-4" />
          <span className="flex-1">名称</span>
          <span className="w-14 text-center">类型</span>
          <span className="w-12 text-center">建议</span>
          <span className="w-14 text-right">文件数</span>
          <span className="w-16 text-center">占比</span>
          <span className="w-16 text-right">大小</span>
        </div>

        {/* Tree Content */}
        <div className="flex-1 overflow-auto p-1">
          {!rootNode && !isAnalyzing ? (
            <div className="flex items-center justify-center h-full text-text-secondary">
              <div className="text-center">
                <HardDrive size={40} className="mx-auto mb-3 text-text-muted" />
                <p className="text-sm">点击"全盘扫描"分析磁盘空间</p>
              </div>
            </div>
          ) : rootNode ? (
            <TreeNode
              node={rootNode}
              level={0}
              selectedPaths={selectedPaths}
              expandedPaths={expandedPaths}
              onToggleSelect={toggleSelect}
              onToggleExpand={toggleExpand}
              onLoadChildren={loadChildren}
              maxSize={rootNode.size}
              showSystemDirs={showSystemDirs}
            />
          ) : null}
        </div>

        {/* Bottom Bar - Inside Tree Card */}
        <div className="flex items-center justify-between px-3 py-2 border-t border-border bg-bg-secondary flex-shrink-0">
          <div className="flex items-center gap-4 text-xs">
            <span className="text-text-muted">
              已选 <span className="text-text-primary font-medium">{selectedPaths.size}</span> 项
            </span>
            <span className="text-accent-blue font-medium">
              {formatBytes(selectedSize)}
            </span>
            {whitelistDirs.length > 0 && (
              <span className="text-text-muted flex items-center gap-1">
                <Shield size={12} />
                {whitelistDirs.length} 个受保护目录
              </span>
            )}
          </div>
          <button
            onClick={handleDeleteClick}
            disabled={isDeleting || selectedPaths.size === 0}
            className="flex items-center gap-1.5 px-4 py-1.5 bg-accent-red text-white rounded-lg hover:bg-red-600 transition-colors disabled:opacity-50 text-sm"
          >
            {isDeleting ? (
              <Loader2 size={14} className="animate-spin" />
            ) : (
              <Trash2 size={14} />
            )}
            <span>{isDeleting ? '删除中...' : '删除选中'}</span>
          </button>
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        isOpen={showDeleteConfirm}
        title="确认删除"
        message={`确定要删除选中的 ${selectedPaths.size} 个项目吗？`}
        detail={`将释放约 ${formatBytes(selectedSize)} 空间。此操作不可恢复！`}
        confirmText="确认删除"
        cancelText="取消"
        onConfirm={handleDeleteConfirm}
        onCancel={handleDeleteCancel}
        variant="danger"
      />
    </div>
  )
}

// Helper function to calculate selected size
function calculateSelectedSize(node: FileNode, selectedPaths: Set<string>): number {
  let size = 0
  if (selectedPaths.has(node.path)) {
    size += Math.max(0, node.size)
  }
  if (node.children) {
    for (const child of node.children) {
      // Don't double count if parent is selected
      if (!selectedPaths.has(node.path)) {
        size += calculateSelectedSize(child, selectedPaths)
      }
    }
  }
  return size
}

// Helper function to update a node's size by path
function updateNodeSize(node: FileNode, targetPath: string, size: number, fileCount: number): FileNode {
  if (node.path === targetPath) {
    return { ...node, size, fileCount }
  }
  if (node.children) {
    const updatedChildren = node.children.map(child => 
      updateNodeSize(child, targetPath, size, fileCount)
    )
    // Recalculate parent size
    const totalSize = updatedChildren.reduce((sum, c) => sum + Math.max(0, c.size), 0)
    const totalCount = updatedChildren.reduce((sum, c) => sum + c.fileCount, 0)
    return { ...node, children: updatedChildren, size: totalSize, fileCount: totalCount }
  }
  return node
}

// Helper function to sort and filter nodes after size calculation
function sortAndFilterNode(node: FileNode, minSize: number): FileNode {
  if (!node.children) return node

  // Filter children by minSize and sort by size
  const filteredChildren = node.children
    .filter(child => child.size >= minSize || child.size < 0) // Keep calculating ones
    .map(child => sortAndFilterNode(child, minSize))
    .sort((a, b) => b.size - a.size)

  return { ...node, children: filteredChildren }
}

// Helper function to sort children of a specific node by path
function sortNodeChildren(node: FileNode, targetPath: string, minSize: number): FileNode {
  if (node.path === targetPath && node.children) {
    const sortedChildren = [...node.children]
      .filter(child => child.size >= minSize || child.size < 0)
      .sort((a, b) => b.size - a.size)
    return { ...node, children: sortedChildren }
  }
  if (node.children) {
    return {
      ...node,
      children: node.children.map(child => sortNodeChildren(child, targetPath, minSize))
    }
  }
  return node
}

// Helper function to remove deleted nodes and update parent sizes
function removeNodesAndUpdateSizes(node: FileNode, deletedPaths: Set<string>): FileNode {
  // If this node was deleted, return a marker (will be filtered out by parent)
  if (deletedPaths.has(node.path)) {
    return { ...node, size: -999 } // Marker for deletion
  }
  
  if (!node.children) return node
  
  // Process children, removing deleted ones
  const remainingChildren = node.children
    .map(child => removeNodesAndUpdateSizes(child, deletedPaths))
    .filter(child => child.size !== -999) // Remove deleted nodes
  
  // Recalculate size based on remaining children
  let newSize = 0
  let newFileCount = 0
  for (const child of remainingChildren) {
    if (child.size > 0) {
      newSize += child.size
      newFileCount += child.fileCount
    }
  }
  
  return {
    ...node,
    children: remainingChildren,
    size: newSize,
    fileCount: newFileCount
  }
}
