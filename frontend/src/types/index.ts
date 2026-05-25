/** TypeScript type definitions */

export interface Page {
  id: number
  url: string
  canonical_url?: string | null
  canonical_key?: string | null
  title: string
  domain: string
  day_count: number
  visit_count: number
  is_bookmarked: boolean
  bookmark_title?: string | null
  bookmark_folder?: string | null
  bookmark_added_at?: string | null
  is_github_starred: boolean
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
  user_intent?: string | null
  key_points?: string[]
  value?: string | null
  next_action?: string | null
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

export type DomainSort = 'recent' | 'visits'

export interface DomainSummary {
  domain: string
  page_count: number
  visit_count: number
  last_visit_time: string
}

export interface DomainListResponse {
  domains: DomainSummary[]
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
