import { LayoutDashboard, Settings, HardDrive, Crown } from 'lucide-react'
import type { Page } from '../types'

interface SidebarProps {
  currentPage: Page
  onNavigate: (page: Page) => void
}

const navItems: { id: Page; label: string; icon: typeof LayoutDashboard; highlight?: boolean }[] = [
  { id: 'dashboard', label: '仪表盘', icon: LayoutDashboard },
  { id: 'analyzer', label: '磁盘分析', icon: HardDrive },
  { id: 'vip', label: 'VIP 会员', icon: Crown, highlight: true },
  { id: 'settings', label: '设置', icon: Settings },
]

// ClearC Logo Component - A stylized "C" with a broom/clean effect
function ClearCLogo() {
  return (
    <svg
      width="36"
      height="36"
      viewBox="0 0 36 36"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="flex-shrink-0"
    >
      {/* Background */}
      <rect width="36" height="36" rx="8" fill="#3B82F6" />
      {/* Letter C */}
      <path
        d="M23 11C21.5 9.5 19.5 8.5 17 8.5C12 8.5 8 12.5 8 18C8 23.5 12 27.5 17 27.5C19.5 27.5 21.5 26.5 23 25"
        stroke="white"
        strokeWidth="3.5"
        strokeLinecap="round"
        fill="none"
      />
      {/* Sparkle effects */}
      <circle cx="26" cy="10" r="1.5" fill="white" opacity="0.9" />
      <circle cx="29" cy="14" r="1" fill="white" opacity="0.7" />
      <circle cx="27" cy="17" r="0.8" fill="white" opacity="0.5" />
    </svg>
  )
}

export default function Sidebar({ currentPage, onNavigate }: SidebarProps) {
  return (
    <aside className="w-60 h-full bg-bg-secondary border-r border-border flex flex-col">
      {/* Logo */}
      <div className="p-5 flex items-center gap-3">
        <ClearCLogo />
        <span className="font-primary text-xl font-semibold text-text-primary">
          ClearC
        </span>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-3 py-2">
        <ul className="space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = currentPage === item.id
            return (
              <li key={item.id}>
                <button
                  onClick={() => onNavigate(item.id)}
                  className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors ${
                    isActive
                      ? 'bg-bg-hover text-text-primary'
                      : item.highlight
                        ? 'text-yellow-500 hover:bg-yellow-500/10'
                        : 'text-text-secondary hover:bg-bg-hover hover:text-text-primary'
                  }`}
                >
                  <Icon
                    size={20}
                    className={isActive ? 'text-accent-blue' : item.highlight ? 'text-yellow-500' : ''}
                  />
                  <span
                    className={`font-secondary text-sm ${
                      isActive ? 'font-medium' : ''
                    }`}
                  >
                    {item.label}
                  </span>
                </button>
              </li>
            )
          })}
        </ul>
      </nav>

      {/* Version info */}
      <div className="p-4 border-t border-border">
        <p className="text-xs text-text-muted text-center">
          ClearC v1.0.0
        </p>
      </div>
    </aside>
  )
}
