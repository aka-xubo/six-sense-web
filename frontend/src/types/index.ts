/** TypeScript type definitions */

export interface Page {
  id: number
  url: string
  title: string
  domain: string
  visit_count: number
  last_visit_time: string
  first_visit_time: string
  created_at: string
  updated_at: string
  has_insights: boolean
  insights?: Insights
}

export interface Insights {
  id: number
  page_id: number
  summary: string
  type: string
  keywords: string[]
  agent_name: string
  analyzed_at: string
  expires_at: string
}

export interface PageListResponse {
  pages: Page[]
  total: number
  has_more: boolean
}

export interface PageDateGroup {
  date_key: string
  title: string
  pages: Page[]
}

export interface PageGroupListResponse {
  groups: PageDateGroup[]
  total: number
  has_more: boolean
  next_cursor: string | null
}

export interface SyncResponse {
  status: string
  new_pages: number
  updated_pages: number
  total_pages: number
  sync_time: string
}

export interface BlacklistResponse {
  hidden_pages: number
}

export type BlacklistType = 'url' | 'domain' | 'path'

export interface BlacklistEntry {
  id: number
  type: BlacklistType
  pattern: string
  created_at: string
}

export interface BlacklistListResponse {
  entries: BlacklistEntry[]
}

export interface AgentInfo {
  name: string
  display_name: string
  version: string | null
  available: boolean
}

export interface AgentsResponse {
  agents: AgentInfo[]
}
