import { useCallback, useState, useEffect, useRef } from 'react'
import { api } from '../services/api'
import type { PageDateGroup, PageGroupListResponse } from '../types'

export function usePages(searchQuery: string = '', dateFrom: string = '', dateTo: string = '', domain: string = '') {
  const [groups, setGroups] = useState<PageDateGroup[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [nextCursor, setNextCursor] = useState<string | null>(null)
  const [hasMore, setHasMore] = useState(true)
  const activeQueryKeyRef = useRef('')
  const inFlightRequestsRef = useRef<Set<string>>(new Set())
  const lastInitialFetchKeyRef = useRef<string | null>(null)

  const groupLimit = dateFrom || dateTo ? 7 : 1
  const queryKey = JSON.stringify({ searchQuery, dateFrom, dateTo, domain, groupLimit })

  const fetchPages = useCallback(async (cursor: string | null = null, append: boolean = false) => {
    const requestKey = JSON.stringify({ queryKey, cursor, append })
    if (inFlightRequestsRef.current.has(requestKey)) return
    inFlightRequestsRef.current.add(requestKey)

    setLoading(true)
    setError(null)
    try {
      const response: PageGroupListResponse = await api.getPageGroups({
        q: searchQuery || undefined,
        domain: domain || undefined,
        cursor,
        dateFrom: dateFrom || undefined,
        dateTo: dateTo || undefined,
        limit: groupLimit
      })

      if (activeQueryKeyRef.current !== queryKey) return

      if (append) {
        setGroups(prev => [...prev, ...response.groups])
      } else {
        setGroups(response.groups)
      }

      setTotal(response.total)
      setHasMore(response.has_more)
      setNextCursor(response.next_cursor)

    } catch (err) {
      if (activeQueryKeyRef.current !== queryKey) return
      setError(err instanceof Error ? err.message : 'Failed to fetch pages')
    } finally {
      inFlightRequestsRef.current.delete(requestKey)
      if (activeQueryKeyRef.current === queryKey) {
        setLoading(false)
      }
    }
  }, [searchQuery, dateFrom, dateTo, domain, groupLimit, queryKey])

  const loadMore = useCallback(() => {
    if (!loading && hasMore && nextCursor) {
      fetchPages(nextCursor, true)
    }
  }, [loading, hasMore, nextCursor, fetchPages])

  const refreshPage = useCallback(async (pageId: number) => {
    const requestQueryKey = queryKey
    try {
      const updatedPage = await api.getPage(pageId)
      if (activeQueryKeyRef.current !== requestQueryKey) return

      setGroups(prev => prev.map(group => ({
        ...group,
        pages: group.pages.map(page => page.id === pageId ? updatedPage : page)
      })))
    } catch (err) {
      if (activeQueryKeyRef.current !== requestQueryKey) return
      setError(err instanceof Error ? err.message : 'Failed to refresh page')
    }
  }, [queryKey])

  useEffect(() => {
    if (lastInitialFetchKeyRef.current === queryKey) return
    lastInitialFetchKeyRef.current = queryKey
    activeQueryKeyRef.current = queryKey
    setGroups([])
    setTotal(0)
    setHasMore(true)
    setNextCursor(null)
    fetchPages(null, false)
  }, [queryKey, fetchPages])

  const pages = groups.flatMap(group => group.pages)

  return {
    pages,
    groups,
    total,
    loading,
    error,
    hasMore,
    loadMore,
    refreshPage,
    refetch: () => fetchPages(null, false)
  }
}
