import { useState, useEffect } from 'react'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { GetVIPStatus } from '../wailsjs/go/main/App'
import Sidebar from './components/Sidebar'
import Dashboard from './components/Dashboard'
import DiskAnalyzerPage from './components/DiskAnalyzerPage'
import SettingsPage from './components/SettingsPage'
import VIPPage from './components/VIPPage'
import VIPModal from './components/VIPModal'
import type { Page, VIPStatus } from './types'

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('dashboard')
  const [vipStatus, setVipStatus] = useState<VIPStatus | null>(null)
  const [showVIPModal, setShowVIPModal] = useState(false)

  useEffect(() => {
    // Load VIP status on startup
    loadVIPStatus()

    // Listen for navigation events from system tray
    const unsubscribeNavigate = EventsOn('navigate', (page: string) => {
      if (page === 'disk-analyzer') {
        handleAnalyzerNavigation()
      } else {
        setCurrentPage(page as Page)
      }
    })

    return () => {
      unsubscribeNavigate()
    }
  }, [])

  // 启动时检查试用期是否过期，如果过期则显示弹窗
  useEffect(() => {
    if (vipStatus && !vipStatus.isVip && vipStatus.isTrialExpired) {
      setShowVIPModal(true)
    }
  }, [vipStatus])

  const loadVIPStatus = async () => {
    try {
      const status = await GetVIPStatus()
      setVipStatus(status)
    } catch (error) {
      console.error('Failed to load VIP status:', error)
    }
  }

  // 检查是否需要显示 VIP 弹窗
  const handleAnalyzerNavigation = () => {
    if (vipStatus && !vipStatus.isVip && vipStatus.isTrialExpired) {
      setShowVIPModal(true)
    } else {
      setCurrentPage('analyzer')
    }
  }

  const handleNavigate = (page: Page) => {
    if (page === 'analyzer') {
      handleAnalyzerNavigation()
    } else {
      setCurrentPage(page)
    }
  }

  const handleUpgrade = () => {
    setShowVIPModal(false)
    setCurrentPage('vip')
  }

  const renderPage = () => {
    switch (currentPage) {
      case 'dashboard':
        return <Dashboard onNavigate={handleNavigate} />
      case 'analyzer':
        return <DiskAnalyzerPage />
      case 'vip':
        return <VIPPage />
      case 'settings':
        return <SettingsPage />
      default:
        return <Dashboard onNavigate={handleNavigate} />
    }
  }

  return (
    <div className="flex h-screen bg-bg-primary">
      <Sidebar currentPage={currentPage} onNavigate={handleNavigate} />
      <main className="flex-1 overflow-auto">
        {renderPage()}
      </main>
      
      {/* VIP Modal */}
      <VIPModal
        isOpen={showVIPModal}
        onClose={() => setShowVIPModal(false)}
        onUpgrade={handleUpgrade}
        trialDaysLeft={vipStatus?.trialDaysLeft}
      />
    </div>
  )
}

export default App
