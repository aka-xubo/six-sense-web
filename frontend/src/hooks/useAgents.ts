import { useState, useEffect, useCallback } from 'react'
import { api } from '../services/api'
import type { AgentInfo } from '../types'

export function useAgents() {
  const [agents, setAgents] = useState<AgentInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchAgents = useCallback(async (cancelled?: () => boolean) => {
    setLoading(true)
    setError(null)
    try {
      const response = await api.getAgents()
      if (!cancelled?.()) {
        setAgents(response.agents)
      }
    } catch (err) {
      if (!cancelled?.()) {
        setError(err instanceof Error ? err.message : 'Failed to fetch agents')
      }
    } finally {
      if (!cancelled?.()) {
        setLoading(false)
      }
    }
  }, [])

  const refresh = useCallback(() => fetchAgents(), [fetchAgents])

  useEffect(() => {
    let cancelled = false

    fetchAgents(() => cancelled)

    return () => {
      cancelled = true
    }
  }, [fetchAgents])

  return { agents, loading, error, refresh }
}
