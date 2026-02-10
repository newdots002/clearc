import { useState, useEffect } from 'react'
import { EventsOn } from '../wailsjs/runtime/runtime'
import Sidebar from './components/Sidebar'
import Dashboard from './components/Dashboard'
import DiskAnalyzerPage from './components/DiskAnalyzerPage'
import SettingsPage from './components/SettingsPage'
import type { Page } from './types'

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('dashboard')

  useEffect(() => {
    // Listen for navigation events from system tray
    const unsubscribeNavigate = EventsOn('navigate', (page: Page) => {
      setCurrentPage(page)
    })

    return () => {
      unsubscribeNavigate()
    }
  }, [])

  const renderPage = () => {
    switch (currentPage) {
      case 'dashboard':
        return <Dashboard onNavigate={setCurrentPage} />
      case 'analyzer':
        return <DiskAnalyzerPage />
      case 'settings':
        return <SettingsPage />
      default:
        return <Dashboard onNavigate={setCurrentPage} />
    }
  }

  return (
    <div className="flex h-screen bg-bg-primary">
      <Sidebar currentPage={currentPage} onNavigate={setCurrentPage} />
      <main className="flex-1 overflow-auto">
        {renderPage()}
      </main>
    </div>
  )
}

export default App
