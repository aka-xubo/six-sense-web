/** API client for Six Sense Web backend */
import type { BlacklistListResponse, BlacklistResponse, BlacklistType, PageGroupListResponse, PageListResponse, SyncResponse, AgentsResponse, Page } from '../types'

const API_BASE = '/api'

export const api = {
  // Pages
  async getPages(params?: {
    q?: string
    limit?: number
    offset?: number
    sort?: string
  }): Promise<PageListResponse> {
    const searchParams = new URLSearchParams()
    if (params?.q) searchParams.set('q', params.q)
    if (params?.limit) searchParams.set('limit', params.limit.toString())
    if (params?.offset) searchParams.set('offset', params.offset.toString())
    if (params?.sort) searchParams.set('sort', params.sort)

    const response = await fetch(`${API_BASE}/pages?${searchParams}`)
    if (!response.ok) throw new Error('Failed to fetch pages')
    return response.json()
  },

  async getPage(id: number): Promise<Page> {
    const response = await fetch(`${API_BASE}/pages/${id}`)
    if (!response.ok) throw new Error('Failed to fetch page')
    return response.json()
  },

  async getPageGroups(params?: {
    q?: string
    cursor?: string | null
    dateFrom?: string
    dateTo?: string
    limit?: number
  }): Promise<PageGroupListResponse> {
    const searchParams = new URLSearchParams()
    if (params?.q) searchParams.set('q', params.q)
    if (params?.cursor) searchParams.set('cursor', params.cursor)
    if (params?.dateFrom) searchParams.set('date_from', params.dateFrom)
    if (params?.dateTo) searchParams.set('date_to', params.dateTo)
    if (params?.limit) searchParams.set('limit', params.limit.toString())

    const response = await fetch(`${API_BASE}/page-groups?${searchParams}`)
    if (!response.ok) throw new Error('Failed to fetch page groups')
    return response.json()
  },

  // Sync
  async sync(months: number = 2): Promise<SyncResponse> {
    const response = await fetch(`${API_BASE}/sync`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ months })
    })
    if (!response.ok) throw new Error('Failed to sync')
    return response.json()
  },

  async blacklistPage(pageId: number, type: BlacklistType = 'url', pattern?: string): Promise<BlacklistResponse> {
    const searchParams = new URLSearchParams()
    searchParams.set('type', type)
    if (pattern) searchParams.set('pattern', pattern)

    const response = await fetch(`${API_BASE}/blacklist/pages/${pageId}?${searchParams}`, {
      method: 'POST'
    })
    if (!response.ok) throw new Error('Failed to add page to blacklist')
    return response.json()
  },

  async getBlacklist(): Promise<BlacklistListResponse> {
    const response = await fetch(`${API_BASE}/blacklist`)
    if (!response.ok) throw new Error('Failed to fetch blacklist')
    return response.json()
  },

  async deleteBlacklistEntry(entryId: number): Promise<BlacklistResponse> {
    const response = await fetch(`${API_BASE}/blacklist/${entryId}`, {
      method: 'DELETE'
    })
    if (!response.ok) throw new Error('Failed to delete blacklist entry')
    return response.json()
  },

  // Agents
  async getAgents(): Promise<AgentsResponse> {
    const response = await fetch(`${API_BASE}/agents`)
    if (!response.ok) throw new Error('Failed to fetch agents')
    return response.json()
  },

  // Analyze (SSE)
  analyzeStream(pageId: number, agentName: string): EventSource {
    return new EventSource(
      `${API_BASE}/analyze?page_id=${pageId}&agent_name=${agentName}`
    )
  }
}
