import { formatRelativeTime } from '../utils/time'
import AgentDropdown from './AgentDropdown'
import type { AgentInfo } from '../types'

interface HeaderProps {
  onSync: () => void
  onOpenBlacklist: () => void
  onSearch: (query: string) => void
  onDateFromChange: (date: string) => void
  onDateToChange: (date: string) => void
  onClearDateRange: () => void
  searchQuery: string
  dateFrom: string
  dateTo: string
  syncing: boolean
  lastSyncTime: string | null
  agents: AgentInfo[]
  selectedAgent: string | null
  onAgentSelect: (agentName: string) => void
}

export default function Header({
  onSync,
  onOpenBlacklist,
  onSearch,
  onDateFromChange,
  onDateToChange,
  onClearDateRange,
  searchQuery,
  dateFrom,
  dateTo,
  syncing,
  lastSyncTime,
  agents,
  selectedAgent,
  onAgentSelect
}: HeaderProps) {
  return (
    <header className="bg-white shadow-sm border-b border-gray-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-3xl font-bold text-gray-900">
            Six Sense
          </h1>
          <div className="flex items-center gap-3">
            {lastSyncTime && (
              <span className="text-xs text-gray-500 whitespace-nowrap">
                最近同步：{formatRelativeTime(lastSyncTime)}
              </span>
            )}
            <AgentDropdown
              agents={agents}
              selectedAgent={selectedAgent}
              onSelect={onAgentSelect}
            />
            <button
              onClick={onSync}
              disabled={syncing}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {syncing ? (
                <>
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  同步中...
                </>
              ) : (
                <>
                  🔄 同步
                </>
              )}
            </button>
            <button
              type="button"
              onClick={onOpenBlacklist}
              className="inline-flex items-center rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            >
              黑名单
            </button>
          </div>
        </div>

        <div className="relative">
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => onSearch(e.target.value)}
            placeholder="搜索标题、域名、URL..."
            className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
          />
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg className="h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
            </svg>
          </div>
        </div>

        <div className="mt-3 flex flex-wrap items-center gap-3">
          <label className="flex items-center gap-2 text-sm text-gray-600">
            <span className="whitespace-nowrap">日期</span>
            <input
              type="date"
              value={dateFrom}
              max={dateTo || undefined}
              onChange={(e) => onDateFromChange(e.target.value)}
              className="h-9 rounded-md border border-gray-300 px-3 text-sm text-gray-700 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </label>
          <span className="text-sm text-gray-400">至</span>
          <input
            type="date"
            value={dateTo}
            min={dateFrom || undefined}
            onChange={(e) => onDateToChange(e.target.value)}
            className="h-9 rounded-md border border-gray-300 px-3 text-sm text-gray-700 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          />
          {(dateFrom || dateTo) && (
            <button
              type="button"
              onClick={onClearDateRange}
              className="h-9 px-3 text-sm font-medium text-gray-600 hover:text-gray-900"
            >
              清空日期
            </button>
          )}
        </div>
      </div>
    </header>
  )
}
