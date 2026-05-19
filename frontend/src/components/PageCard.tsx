import type { Page, AgentInfo, BlacklistType } from '../types'
import { formatRelativeTime } from '../utils/time'
import { useState } from 'react'
import AgentSelector from './AgentSelector'
import StreamingText from './StreamingText'
import { useAnalyze } from '../hooks/useAnalyze'

interface PageCardProps {
  page: Page
  agents: AgentInfo[]
  onAnalyzeComplete?: () => void
  onBlacklist?: (pageId: number, type: BlacklistType, pattern?: string) => Promise<void>
}

export default function PageCard({ page, agents, onAnalyzeComplete, onBlacklist }: PageCardProps) {
  const [showAgentSelector, setShowAgentSelector] = useState(false)
  const [showBlacklistMenu, setShowBlacklistMenu] = useState(false)
  const [blacklisting, setBlacklisting] = useState(false)
  const { analyzing, streamText, error, analyze, reset } = useAnalyze()

  const handleAnalyzeClick = () => {
    setShowAgentSelector(true)
  }

  const handleAgentSelect = (agentName: string) => {
    setShowAgentSelector(false)
    analyze(page.id, agentName, () => {
      onAnalyzeComplete?.()
    })
  }

  const getPathPattern = () => {
    try {
      const url = new URL(page.url)
      return `${url.hostname}${url.pathname || '/'}`
    } catch {
      return `${page.domain}/`
    }
  }

  const handleBlacklistClick = async (type: BlacklistType) => {
    if (!onBlacklist || blacklisting) return
    const pattern = type === 'url' ? page.url : type === 'domain' ? page.domain : getPathPattern()
    const labels: Record<BlacklistType, string> = {
      url: '完整地址',
      domain: '域名',
      path: '域名路径'
    }
    const confirmed = window.confirm(`确认将此${labels[type]}加入黑名单并隐藏匹配记录？\n\n${pattern}`)
    if (!confirmed) return

    setBlacklisting(true)
    setShowBlacklistMenu(false)
    try {
      await onBlacklist(page.id, type, pattern)
    } finally {
      setBlacklisting(false)
    }
  }

  return (
    <div className="px-6 py-4 hover:bg-gray-50 transition-colors">
      <div className="flex items-start space-x-3">
        {/* Favicon */}
        <img
          src={`https://${page.domain}/favicon.ico`}
          alt=""
          className="w-4 h-4 mt-0.5 flex-shrink-0"
          onError={(e) => {
            e.currentTarget.src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="gray"><circle cx="12" cy="12" r="10"/></svg>'
          }}
        />

        {/* 内容区域 */}
        <div className="flex-1 min-w-0">
          {/* 标题和时间 */}
          <div className="mb-1">
            <div className="min-w-0">
              <a
                href={page.url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm font-medium text-blue-600 hover:text-blue-800 hover:underline block truncate"
              >
                {page.title}
              </a>
              <div className="flex items-center gap-x-2 gap-y-0.5 mt-0.5 flex-wrap">
                <span className="text-xs text-gray-500 truncate">{page.domain}</span>
                <span className="text-xs text-gray-400">·</span>
                <span className="text-xs text-gray-500">{page.visit_count} 次访问</span>
                <span className="text-xs text-gray-400">·</span>
                <span className="text-xs text-gray-500">{formatRelativeTime(page.last_visit_time)}</span>
              </div>
              <div className="mt-0.5 text-xs text-gray-400 break-all">
                {page.url}
              </div>
            </div>
          </div>

          {/* Insights 或分析按钮 */}
          {page.has_insights && page.insights && !analyzing ? (
            <div className="mt-2 p-3 bg-blue-50 rounded-md border border-blue-100">
              <div className="text-sm text-gray-700 mb-2">
                📝 {page.insights.summary}
              </div>
              <div className="flex items-center justify-between text-xs">
                <div className="flex items-center space-x-2 flex-wrap">
                  <span className="px-2 py-0.5 bg-blue-100 text-blue-700 rounded">
                    {page.insights.type}
                  </span>
                  {page.insights.keywords.map((keyword, idx) => (
                    <span key={idx} className="px-2 py-0.5 bg-gray-100 text-gray-600 rounded">
                      {keyword}
                    </span>
                  ))}
                </div>
                <span className="text-gray-500">
                  {page.insights.agent_name}
                </span>
              </div>
            </div>
          ) : analyzing ? (
            <div className="mt-2">
              <StreamingText text={streamText} />
            </div>
          ) : error ? (
            <div className="mt-2 p-2 bg-red-50 rounded-md border border-red-200">
              <div className="text-xs text-red-700">
                ❌ {error}
              </div>
              <button
                onClick={() => reset()}
                className="mt-1 text-xs text-red-600 hover:text-red-700"
              >
                重试
              </button>
            </div>
          ) : (
            <div className="mt-2 flex flex-wrap items-center gap-2">
              <button
                onClick={handleAnalyzeClick}
                className="px-3 py-1 text-xs font-medium text-blue-600 bg-blue-50 rounded hover:bg-blue-100 focus:outline-none"
              >
                💡 分析
              </button>
              <div className="relative">
                <button
                  onClick={() => setShowBlacklistMenu(prev => !prev)}
                  disabled={blacklisting}
                  className="px-3 py-1 text-xs font-medium text-gray-600 bg-gray-100 rounded hover:bg-gray-200 focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {blacklisting ? '加入中...' : '加入黑名单'}
                </button>
                {showBlacklistMenu && (
                  <div className="absolute left-0 z-10 mt-1 w-44 overflow-hidden rounded-md border border-gray-200 bg-white shadow-lg">
                    <button
                      type="button"
                      onClick={() => handleBlacklistClick('url')}
                      className="block w-full px-3 py-2 text-left text-xs text-gray-700 hover:bg-gray-50"
                    >
                      屏蔽此 URL
                    </button>
                    <button
                      type="button"
                      onClick={() => handleBlacklistClick('domain')}
                      className="block w-full px-3 py-2 text-left text-xs text-gray-700 hover:bg-gray-50"
                    >
                      屏蔽此域名
                    </button>
                    <button
                      type="button"
                      onClick={() => handleBlacklistClick('path')}
                      className="block w-full px-3 py-2 text-left text-xs text-gray-700 hover:bg-gray-50"
                    >
                      屏蔽此域名路径
                    </button>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Agent Selector Modal */}
      {showAgentSelector && (
        <AgentSelector
          agents={agents}
          onSelect={handleAgentSelect}
          onClose={() => setShowAgentSelector(false)}
        />
      )}
    </div>
  )
}
