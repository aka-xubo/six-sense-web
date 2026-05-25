import type { Page, AgentInfo, BlacklistType } from '../types'
import { formatRelativeTime } from '../utils/time'
import { formatVisitCount, getVisitCountBadgeClass } from '../utils/visitCount'
import { useState } from 'react'
import StreamingText from './StreamingText'
import { useAnalyze } from '../hooks/useAnalyze'
import AgentIcon from './AgentIcon'

interface PageCardProps {
  page: Page
  agents: AgentInfo[]
  selectedAgent: string | null
  onAnalyzeComplete?: (pageId: number) => void
  onBlacklist?: (pageId: number, type: BlacklistType, pattern?: string) => Promise<void>
}

export default function PageCard({ page, agents, selectedAgent, onAnalyzeComplete, onBlacklist }: PageCardProps) {
  const [showBlacklistMenu, setShowBlacklistMenu] = useState(false)
  const [blacklisting, setBlacklisting] = useState(false)
  const { analyzing, streamText, error, analyze, reset } = useAnalyze()
  const displayUrl = page.canonical_url || page.url

  const startAnalyze = (force = false) => {
    if (!selectedAgent) {
      alert('请先在页面顶部选择一个 AI Agent')
      return
    }

    const selectedAgentInfo = agents.find(a => a.name === selectedAgent)
    if (!selectedAgentInfo?.available) {
      alert('当前选中的 Agent 不可用，请选择其他 Agent')
      return
    }

    analyze(page.id, selectedAgent, () => {
      onAnalyzeComplete?.(page.id)
    }, force)
  }

  const handleAnalyzeClick = () => {
    startAnalyze(false)
  }

  const handleReanalyzeClick = () => {
    startAnalyze(true)
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
    <div className="relative px-6 py-4 hover:bg-gray-50 transition-colors">
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
                href={displayUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm font-medium text-blue-600 hover:text-blue-800 hover:underline block truncate"
              >
                {page.title}
              </a>
              <div className="mt-1 flex flex-wrap items-center gap-1.5">
                <span className="max-w-[18rem] truncate text-xs text-gray-500">{page.domain}</span>
                {page.is_bookmarked && (
                  <span
                    className="inline-flex h-5 items-center rounded border border-yellow-200 bg-yellow-50 px-1.5 text-xs font-medium text-yellow-700"
                    title={page.bookmark_folder ? `收藏于 ${page.bookmark_folder}` : '已收藏'}
                  >
                    收藏
                  </span>
                )}
                {page.is_github_starred && (
                  <span className="inline-flex h-5 items-center rounded border border-purple-200 bg-purple-50 px-1.5 text-xs font-medium text-purple-700">
                    已 Star
                  </span>
                )}
                <span className={getVisitCountBadgeClass(page.day_count)}>
                  当天 {formatVisitCount(page.day_count)} 次
                </span>
                <span className={getVisitCountBadgeClass(page.visit_count)}>
                  累计 {formatVisitCount(page.visit_count)} 次
                </span>
                <span className="inline-flex h-5 items-center rounded border border-gray-200 bg-white px-1.5 text-xs text-gray-500">
                  最近 {formatRelativeTime(page.last_visit_time)}
                </span>
              </div>
              <div className="mt-0.5 text-xs text-gray-400 break-all">
                {displayUrl}
              </div>
            </div>
          </div>

          {/* Insights 或分析按钮 */}
          {page.has_insights && page.insights && !analyzing ? (
            <div className="mt-2 rounded-md border border-blue-100 bg-blue-50 p-3">
              <div className="mb-2 flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="text-sm font-medium leading-5 text-gray-800">
                    {page.insights.summary}
                  </div>
                  {page.insights.user_intent && (
                    <div className="mt-1 text-xs leading-5 text-gray-600">
                      意图：{page.insights.user_intent}
                    </div>
                  )}
                </div>
                <span className="inline-flex items-center pt-0.5" aria-label={`由 ${page.insights.agent_name} 分析`}>
                  <AgentIcon agentName={page.insights.agent_name} className="h-4 w-4" showFallbackText />
                </span>
              </div>

              {page.insights.key_points && page.insights.key_points.length > 0 && (
                <ul className="mb-2 space-y-1 text-xs leading-5 text-gray-700">
                  {page.insights.key_points.map((point, idx) => (
                    <li key={idx} className="flex gap-2">
                      <span className="mt-2 h-1 w-1 flex-shrink-0 rounded-full bg-blue-400" />
                      <span>{point}</span>
                    </li>
                  ))}
                </ul>
              )}

              {(page.insights.value || page.insights.next_action) && (
                <div className="mb-2 grid gap-1 text-xs leading-5 text-gray-600 sm:grid-cols-2">
                  {page.insights.value && (
                    <div>
                      <span className="font-medium text-gray-700">价值：</span>
                      {page.insights.value}
                    </div>
                  )}
                  {page.insights.next_action && (
                    <div>
                      <span className="font-medium text-gray-700">下一步：</span>
                      {page.insights.next_action}
                    </div>
                  )}
                </div>
              )}

              <div className="flex flex-wrap items-center justify-between gap-2 text-xs">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="rounded bg-blue-100 px-2 py-0.5 text-blue-700">
                    {page.insights.type}
                  </span>
                  {page.insights.keywords.map((keyword, idx) => (
                    <span key={idx} className="rounded bg-gray-100 px-2 py-0.5 text-gray-600">
                      {keyword}
                    </span>
                  ))}
                </div>
                <span className="text-gray-400">
                  {new Date(page.insights.analyzed_at).toLocaleDateString()}
                </span>
              </div>
              <button
                type="button"
                onClick={handleReanalyzeClick}
                className="mt-2 text-xs font-medium text-blue-600 hover:text-blue-700 focus:outline-none"
              >
                重新分析
              </button>
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
                  <div className="absolute left-0 z-50 mt-1 w-44 overflow-hidden rounded-md border border-gray-200 bg-white shadow-lg">
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

    </div>
  )
}
