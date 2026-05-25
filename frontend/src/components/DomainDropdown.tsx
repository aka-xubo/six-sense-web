import { useEffect, useRef, useState } from 'react'
import type { DomainSort, DomainSummary } from '../types'
import { formatRelativeTime } from '../utils/time'
import { formatVisitCount, getVisitCountBadgeClass } from '../utils/visitCount'

interface DomainDropdownProps {
  domains: DomainSummary[]
  selectedDomain: string
  sort: DomainSort
  loading?: boolean
  error?: string | null
  onSelect: (domain: string) => void
  onSortChange: (sort: DomainSort) => void
  onRefresh?: () => void
}

const sortLabels: Record<DomainSort, string> = {
  recent: '最近访问',
  visits: '访问次数'
}

function faviconUrl(domain: string) {
  return `https://${domain}/favicon.ico`
}

export default function DomainDropdown({
  domains,
  selectedDomain,
  sort,
  loading = false,
  error = null,
  onSelect,
  onSortChange,
  onRefresh
}: DomainDropdownProps) {
  const [open, setOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const selectedSummary = selectedDomain ? domains.find(item => item.domain === selectedDomain) : null

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleSelect = (domain: string) => {
    onSelect(domain)
    setOpen(false)
  }

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        type="button"
        onClick={() => setOpen(prev => !prev)}
        className="inline-flex h-9 max-w-full items-center gap-2 rounded-md border border-gray-300 bg-white px-3 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
        aria-expanded={open}
      >
        {selectedDomain ? (
          <img
            src={faviconUrl(selectedDomain)}
            alt=""
            className="h-4 w-4 flex-shrink-0"
            onError={(e) => {
              e.currentTarget.src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="gray"><circle cx="12" cy="12" r="10"/></svg>'
            }}
          />
        ) : (
          <svg className="h-4 w-4 flex-shrink-0 text-gray-500" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M4.25 5.5A2.25 2.25 0 016.5 3.25h7A2.25 2.25 0 0115.75 5.5v9A2.25 2.25 0 0113.5 16.75h-7a2.25 2.25 0 01-2.25-2.25v-9zm2.25-.75a.75.75 0 00-.75.75v1h8.5v-1a.75.75 0 00-.75-.75h-7zm7.75 3.25h-8.5v6.5c0 .414.336.75.75.75h7a.75.75 0 00.75-.75V8z" clipRule="evenodd" />
          </svg>
        )}
        <span className="max-w-[180px] truncate">{selectedDomain || '所有域名'}</span>
        {selectedSummary && (
          <span className="hidden text-xs font-normal text-gray-500 sm:inline">
            {selectedSummary.page_count} 页
          </span>
        )}
        <svg className={`h-4 w-4 flex-shrink-0 text-gray-500 transition-transform ${open ? 'rotate-180' : ''}`} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
          <path fillRule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clipRule="evenodd" />
        </svg>
      </button>

      {open && (
        <div className="absolute right-0 z-20 mt-2 w-[min(24rem,calc(100vw-2rem))] overflow-hidden rounded-md border border-gray-200 bg-white shadow-lg">
          <div className="border-b border-gray-100 p-2">
            <div className="grid grid-cols-2 rounded-md bg-gray-100 p-1 text-xs font-medium text-gray-600">
              {(['recent', 'visits'] as DomainSort[]).map(item => (
                <button
                  key={item}
                  type="button"
                  onClick={() => onSortChange(item)}
                  className={`rounded px-2 py-1.5 ${sort === item ? 'bg-white text-gray-900 shadow-sm' : 'hover:text-gray-900'}`}
                >
                  {sortLabels[item]}
                </button>
              ))}
            </div>
            <button
              type="button"
              onClick={() => handleSelect('')}
              className={`mt-2 flex w-full items-center justify-between gap-3 rounded-md px-2 py-2 text-left text-sm hover:bg-gray-50 ${selectedDomain === '' ? 'bg-blue-50 text-blue-700' : 'text-gray-700'}`}
            >
              <span className="font-medium">所有域名（不过滤）</span>
              <span className="text-xs text-gray-500">{domains.length} 个站点</span>
            </button>
          </div>

          <div className="max-h-80 overflow-y-auto py-1">
            {loading ? (
              <div className="px-3 py-6 text-center text-sm text-gray-500">加载域名中...</div>
            ) : error ? (
              <div className="px-3 py-4">
                <div className="text-sm text-red-600">域名加载失败</div>
                {onRefresh && (
                  <button type="button" onClick={onRefresh} className="mt-2 text-xs font-medium text-blue-600 hover:text-blue-700">
                    重试
                  </button>
                )}
              </div>
            ) : (
              domains.map(domain => (
                <button
                  key={domain.domain}
                  type="button"
                  onClick={() => handleSelect(domain.domain)}
                  className={`flex w-full items-center gap-3 px-3 py-2 text-left hover:bg-gray-50 ${selectedDomain === domain.domain ? 'bg-blue-50' : ''}`}
                >
                  <img
                    src={faviconUrl(domain.domain)}
                    alt=""
                    className="h-4 w-4 flex-shrink-0"
                    onError={(e) => {
                      e.currentTarget.src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="gray"><circle cx="12" cy="12" r="10"/></svg>'
                    }}
                  />
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-sm font-medium text-gray-800">{domain.domain}</span>
                    <span className="block truncate text-xs text-gray-400">最近 {formatRelativeTime(domain.last_visit_time)}</span>
                  </span>
                  <span className="flex min-w-[4.5rem] flex-shrink-0 flex-col items-end gap-1 text-xs">
                    <span className="text-gray-400">{domain.page_count} 页</span>
                    <span className={getVisitCountBadgeClass(domain.visit_count, 'domain')}>
                      {formatVisitCount(domain.visit_count)} 次
                    </span>
                  </span>
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  )
}
