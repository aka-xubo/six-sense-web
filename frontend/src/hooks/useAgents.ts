import { useState, useEffect, useRef } from 'react'
import { api } from '../services/api'
import type { AgentInfo } from '../types'

export function useAgents() {
  const [agents, setAgents] = useState<AgentInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const fetchedRef = useRef(false)

  useEffect(() => {
    if (fetchedRef.current) return
    fetchedRef.current = true
    let cancelled = false

    const fetchAgents = async () => {
      try {
        const response = await api.getAgents()
        if (!cancelled) {
          setAgents(response.agents)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to fetch agents')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    fetchAgents()

    return () => {
      cancelled = true
    }
  }, [])

  return { agents, loading, error }
}
