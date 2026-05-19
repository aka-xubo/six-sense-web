import { useEffect, useState } from 'react'
import type { AgentInfo, BlacklistType, PageDateGroup } from '../types'
import PageCard from './PageCard'

interface PageListProps {
  groups: PageDateGroup[]
  loading: boolean
  agents: AgentInfo[]
  selectedAgent: string | null
  onAnalyzeComplete?: () => void
  onBlacklist?: (pageId: number, type: BlacklistType, pattern?: string) => Promise<void>
  onLastGroupCollapsedChange?: (collapsed: boolean) => void
}

export default function PageList({
  groups,
  loading,
  agents,
  selectedAgent,
  onAnalyzeComplete,
  onBlacklist,
  onLastGroupCollapsedChange
}: PageListProps) {
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(() => new Set())

  const toggleGroup = (dateKey: string) => {
    setCollapsedGroups(prev => {
      const next = new Set(prev)
      if (next.has(dateKey)) {
        next.delete(dateKey)
      } else {
        next.add(dateKey)
      }
      return next
    })
  }

  useEffect(() => {
    const lastGroup = groups[groups.length - 1]
    onLastGroupCollapsedChange?.(lastGroup ? collapsedGroups.has(lastGroup.date_key) : false)
  }, [groups, collapsedGroups, onLastGroupCollapsedChange])

  if (loading) {
    return (
      <div className="flex justify-center items-center py-12">
        <svg className="animate-spin h-8 w-8 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>
    )
  }

  if (groups.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">没有找到匹配的页面</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {groups.map((group) => {
        const collapsed = collapsedGroups.has(group.date_key)

        return (
          <section key={group.date_key} className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
            {/* 日期标题 */}
            <button
              type="button"
              onClick={() => toggleGroup(group.date_key)}
              className="flex w-full items-center justify-between gap-4 px-6 py-3 bg-gray-50 border-b border-gray-200 text-left hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
              aria-expanded={!collapsed}
            >
              <span className="flex min-w-0 items-center gap-2">
                <svg
                  className={`h-4 w-4 flex-shrink-0 text-gray-500 transition-transform ${collapsed ? '' : 'rotate-90'}`}
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path fillRule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clipRule="evenodd" />
                </svg>
                <span className="truncate text-sm font-medium text-gray-700">{group.title}</span>
              </span>
              <span className="flex-shrink-0 text-xs text-gray-500">{group.pages.length} 条</span>
            </button>

            {/* 页面列表 */}
            {!collapsed && (
              <div className="divide-y divide-gray-100">
                {group.pages.map((page) => (
                  <PageCard
                    key={page.id}
                    page={page}
                    agents={agents}
                    selectedAgent={selectedAgent}
                    onAnalyzeComplete={onAnalyzeComplete}
                    onBlacklist={onBlacklist}
                  />
                ))}
              </div>
            )}
          </section>
        )
      })}
    </div>
  )
}
