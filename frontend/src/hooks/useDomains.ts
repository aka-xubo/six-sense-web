import { useCallback, useEffect, useState } from 'react'
import { api } from '../services/api'
import type { DomainSort, DomainSummary } from '../types'

export function useDomains(sort: DomainSort) {
  const [domains, setDomains] = useState<DomainSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchDomains = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await api.getDomains(sort)
      setDomains(response.domains)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch domains')
    } finally {
      setLoading(false)
    }
  }, [sort])

  useEffect(() => {
    fetchDomains()
  }, [fetchDomains])

  return {
    domains,
    loading,
    error,
    refresh: fetchDomains
  }
}
