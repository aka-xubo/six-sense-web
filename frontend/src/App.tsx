import { useEffect, useRef, useState } from 'react'
import Header from './components/Header'
import EmptyState from './components/EmptyState'
import PageList from './components/PageList'
import BlacklistManager from './components/BlacklistManager'
import { usePages } from './hooks/usePages'
import { useAgents } from './hooks/useAgents'
import { api } from './services/api'
import { formatShortDate } from './utils/time'
import type { BlacklistEntry, BlacklistType } from './types'

function App() {
  const [searchQuery, setSearchQuery] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [syncing, setSyncing] = useState(false)
  const [showBlacklistManager, setShowBlacklistManager] = useState(false)
  const [blacklistEntries, setBlacklistEntries] = useState<BlacklistEntry[]>([])
  const [blacklistLoading, setBlacklistLoading] = useState(false)
  const [deletingBlacklistId, setDeletingBlacklistId] = useState<number | null>(null)
  const [lastSyncTime, setLastSyncTime] = useState<string | null>(() => {
    return localStorage.getItem('six-sense:last-sync-time')
  })
  const lastAutoLoadScrollYRef = useRef(0)
  const lastAutoLoadAtRef = useRef(0)
  const autoLoadArmedRef = useRef(true)
  const [lastGroupCollapsed, setLastGroupCollapsed] = useState(false)
  const { pages, groups, total, loading, hasMore, loadMore, refetch } = usePages(searchQuery, dateFrom, dateTo)
  const { agents } = useAgents()
  const [selectedAgent, setSelectedAgent] = useState<string | null>(null)

  // Initialize selected agent from localStorage or first available
  useEffect(() => {
    if (agents.length === 0) return

    const saved = localStorage.getItem('six-sense:selected-agent')
    if (saved && agents.some(a => a.name === saved && a.available)) {
      setSelectedAgent(saved)
    } else {
      const firstAvailable = agents.find(a => a.available)
      if (firstAvailable) {
        setSelectedAgent(firstAvailable.name)
        localStorage.setItem('six-sense:selected-agent', firstAvailable.name)
      }
    }
  }, [agents])

  useEffect(() => {
    const handleScroll = () => {
      if (!hasMore || loading || lastGroupCollapsed) return

      const scrollY = window.scrollY
      const distanceToBottom = document.documentElement.scrollHeight - (scrollY + window.innerHeight)
      const now = Date.now()

      if (distanceToBottom > 600) {
        autoLoadArmedRef.current = true
      }

      if (
        autoLoadArmedRef.current &&
        scrollY > 0 &&
        distanceToBottom < 240 &&
        now - lastAutoLoadAtRef.current > 1200 &&
        Math.abs(scrollY - lastAutoLoadScrollYRef.current) > 80
      ) {
        autoLoadArmedRef.current = false
        lastAutoLoadAtRef.current = now
        lastAutoLoadScrollYRef.current = scrollY
        loadMore()
      }
    }

    window.addEventListener('scroll', handleScroll, { passive: true })

    return () => {
      window.removeEventListener('scroll', handleScroll)
    }
  }, [hasMore, loading, lastGroupCollapsed, loadMore])

  const handleSync = async () => {
    setSyncing(true)
    try {
      const response = await api.sync(2)
      setLastSyncTime(response.sync_time)
      localStorage.setItem('six-sense:last-sync-time', response.sync_time)
      await refetch()
    } catch (error) {
      console.error('Sync failed:', error)
      alert('同步失败，请检查后端服务是否正常运行')
    } finally {
      setSyncing(false)
    }
  }

  const openBlacklistManager = async () => {
    setShowBlacklistManager(true)
    setBlacklistLoading(true)
    try {
      const response = await api.getBlacklist()
      setBlacklistEntries(response.entries)
    } catch (error) {
      console.error('Fetch blacklist failed:', error)
      alert('获取黑名单失败，请检查后端服务是否正常运行')
    } finally {
      setBlacklistLoading(false)
    }
  }

  const handleDeleteBlacklistEntry = async (entry: BlacklistEntry) => {
    const confirmed = window.confirm(`确认移除此黑名单规则？\n\n${entry.pattern}`)
    if (!confirmed) return

    setDeletingBlacklistId(entry.id)
    try {
      await api.deleteBlacklistEntry(entry.id)
      const response = await api.getBlacklist()
      setBlacklistEntries(response.entries)
      await refetch()
    } catch (error) {
      console.error('Delete blacklist failed:', error)
      alert('移除黑名单失败，请检查后端服务是否正常运行')
    } finally {
      setDeletingBlacklistId(null)
    }
  }

  const handleSearch = (query: string) => {
    setSearchQuery(query)
  }

  const handleDateFromChange = (date: string) => {
    setDateFrom(date)
    if (dateTo && date && date > dateTo) {
      setDateTo(date)
    }
  }

  const handleDateToChange = (date: string) => {
    setDateTo(date)
    if (dateFrom && date && date < dateFrom) {
      setDateFrom(date)
    }
  }

  const handleClearDateRange = () => {
    setDateFrom('')
    setDateTo('')
  }

  const handleAnalyzeComplete = () => {
    // Refresh page list after analysis completes
    refetch()
  }

  const handleBlacklistPage = async (pageId: number, type: BlacklistType, pattern?: string) => {
    try {
      const response = await api.blacklistPage(pageId, type, pattern)
      await refetch()
      alert(`已加入黑名单，并隐藏 ${response.hidden_pages} 条历史记录`)
    } catch (error) {
      console.error('Add blacklist failed:', error)
      alert('加入黑名单失败，请检查后端服务是否正常运行')
    }
  }

  const handleAgentSelect = (agentName: string) => {
    setSelectedAgent(agentName)
    localStorage.setItem('six-sense:selected-agent', agentName)
  }

  const loadedTimeRange = (() => {
    if (pages.length === 0) return null

    const timestamps = pages.map(page => new Date(page.last_visit_time).getTime())
    const from = new Date(Math.min(...timestamps)).toISOString()
    const to = new Date(Math.max(...timestamps)).toISOString()
    const fromText = formatShortDate(from)
    const toText = formatShortDate(to)

    return fromText === toText ? fromText : `${fromText} 至 ${toText}`
  })()

  return (
    <div className="min-h-screen bg-gray-50">
      <Header
        onSync={handleSync}
        onOpenBlacklist={openBlacklistManager}
        onSearch={handleSearch}
        onDateFromChange={handleDateFromChange}
        onDateToChange={handleDateToChange}
        onClearDateRange={handleClearDateRange}
        searchQuery={searchQuery}
        dateFrom={dateFrom}
        dateTo={dateTo}
        syncing={syncing}
        lastSyncTime={lastSyncTime}
        agents={agents}
        selectedAgent={selectedAgent}
        onAgentSelect={handleAgentSelect}
      />

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {!loading && pages.length === 0 && !searchQuery ? (
          <EmptyState onSync={handleSync} syncing={syncing} />
        ) : (
          <>
            {total > 0 && (
              <div className="mb-4 text-sm text-gray-600">
                显示 {pages.length} / {total} 个页面
                {loadedTimeRange && ` · 时间：${loadedTimeRange}`}
                {searchQuery && ` · 搜索 "${searchQuery}"`}
                {(dateFrom || dateTo) && ` · 日期条件：${dateFrom || '最早'} 至 ${dateTo || '最新'}`}
              </div>
            )}
            <PageList
              key="page-list"
              groups={groups}
              loading={loading && pages.length === 0}
              agents={agents}
              selectedAgent={selectedAgent}
              onAnalyzeComplete={handleAnalyzeComplete}
              onBlacklist={handleBlacklistPage}
              onLastGroupCollapsedChange={setLastGroupCollapsed}
            />
            {loading && pages.length > 0 && (
              <div className="mt-8 flex justify-center">
                <svg className="animate-spin h-6 w-6 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              </div>
            )}
          </>
        )}
      </main>

      {showBlacklistManager && (
        <BlacklistManager
          entries={blacklistEntries}
          loading={blacklistLoading}
          deletingId={deletingBlacklistId}
          onDelete={handleDeleteBlacklistEntry}
          onClose={() => setShowBlacklistManager(false)}
        />
      )}
    </div>
  )
}

export default App
