import { useCallback, useState, useEffect, useRef } from 'react'
import { api } from '../services/api'
import type { PageDateGroup, PageGroupListResponse } from '../types'

export function usePages(searchQuery: string = '', dateFrom: string = '', dateTo: string = '') {
  const [groups, setGroups] = useState<PageDateGroup[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [nextCursor, setNextCursor] = useState<string | null>(null)
  const [hasMore, setHasMore] = useState(true)
  const fetchingRef = useRef(false)
  const lastInitialFetchKeyRef = useRef<string | null>(null)

  const groupLimit = dateFrom || dateTo ? 7 : 1

  const fetchPages = useCallback(async (cursor: string | null = null, append: boolean = false) => {
    if (fetchingRef.current) return
    fetchingRef.current = true

    setLoading(true)
    setError(null)
    try {
      const response: PageGroupListResponse = await api.getPageGroups({
        q: searchQuery || undefined,
        cursor,
        dateFrom: dateFrom || undefined,
        dateTo: dateTo || undefined,
        limit: groupLimit
      })

      if (append) {
        setGroups(prev => [...prev, ...response.groups])
      } else {
        setGroups(response.groups)
      }

      setTotal(response.total)
      setHasMore(response.has_more)
      setNextCursor(response.next_cursor)

    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch pages')
    } finally {
      fetchingRef.current = false
      setLoading(false)
    }
  }, [searchQuery, dateFrom, dateTo, groupLimit])

  const loadMore = useCallback(() => {
    if (!loading && hasMore && nextCursor) {
      fetchPages(nextCursor, true)
    }
  }, [loading, hasMore, nextCursor, fetchPages])

  useEffect(() => {
    const fetchKey = JSON.stringify({ searchQuery, dateFrom, dateTo })
    if (lastInitialFetchKeyRef.current === fetchKey) return
    lastInitialFetchKeyRef.current = fetchKey
    setNextCursor(null)
    fetchPages(null, false)
  }, [searchQuery, dateFrom, dateTo, fetchPages])

  const pages = groups.flatMap(group => group.pages)

  return {
    pages,
    groups,
    total,
    loading,
    error,
    hasMore,
    loadMore,
    refetch: () => fetchPages(null, false)
  }
}
