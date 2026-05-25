# Global Agent Selector Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a global agent selector dropdown in the header that persists user's agent choice and eliminates per-analysis agent selection dialogs.

**Architecture:** Create a new AgentDropdown component in the Header, manage selected agent state in App with localStorage persistence, and modify PageCard to use the global agent directly instead of showing a selection dialog.

**Tech Stack:** React, TypeScript, TailwindCSS, localStorage

---

## File Structure

**New Files:**
- `frontend/src/components/AgentDropdown.tsx` - Dropdown component for agent selection

**Modified Files:**
- `frontend/src/components/Header.tsx` - Integrate AgentDropdown
- `frontend/src/App.tsx` - Add global agent state management
- `frontend/src/components/PageCard.tsx` - Remove AgentSelector dialog, use global agent

---

### Task 1: Create AgentDropdown Component

**Files:**
- Create: `frontend/src/components/AgentDropdown.tsx`

- [ ] **Step 1: Write the component structure test**

```typescript
// This is a visual component, we'll verify it manually after implementation
// No automated test needed for this task
```

- [ ] **Step 2: Create AgentDropdown component with basic structure**

Create `frontend/src/components/AgentDropdown.tsx`:

```typescript
import { useState, useRef, useEffect } from 'react'
import type { AgentInfo } from '../types'

interface AgentDropdownProps {
  agents: AgentInfo[]
  selectedAgent: string | null
  onSelect: (agentName: string) => void
}

export default function AgentDropdown({ agents, selectedAgent, onSelect }: AgentDropdownProps) {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => {
        document.removeEventListener('mousedown', handleClickOutside)
      }
    }
  }, [isOpen])

  const selectedAgentInfo = agents.find(a => a.name === selectedAgent)
  const hasAvailableAgents = agents.some(a => a.available)

  const getAgentIcon = (agentName: string) => {
    if (agentName === 'claude') return '🤖'
    if (agentName === 'codex') return '⚡'
    return '🔧'
  }

  const handleToggle = () => {
    setIsOpen(!isOpen)
  }

  const handleSelect = (agentName: string) => {
    onSelect(agentName)
    setIsOpen(false)
  }

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Dropdown Button */}
      <button
        onClick={handleToggle}
        aria-label="选择 AI Agent"
      me={`inline-flex items-center gap-2 px-3 py-2 border rounded-md text-sm font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 ${
          !hasAvailableAgents || !selectedAgentInfo
            ? 'border-yellow-400 bg-yellow-50 text-yellow-800'
            : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
        }`}
      >
        {selectedAgentInfo ? (
          <>
            <span className="text-base">{getAgentIcon(selectedAgentInfo.name)}</span>
            <span>{selectedAgentInfo.display_name}</span>
            <span
              className={`inline-block w-2 h-2 ed-full ${
                selectedAgentInfo.available ? 'bg-green-500' : 'bg-gray-400'
              }`}
            />
          </>
        ) : (
          <>
            <span className="text-base">⚠️</span>
            <span>未检测到 Agent</span>
          </>
        )}
        <svg
          className={`w-4 h-4 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 20 20"
          fill="currentColor"
        >
          <path
            fillRule="evenodd"
            d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
            clipRule="evenodd"
          />
        </svg>
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <div
          role="menu"
          className="absolute right-0 mt-2 w-64 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50"
        >
          {agents.map((agent) => (
            <button
              key={agent.name}
              role="menuitem"
              onClick={() => handleSelect(agent.name)}
              disabled={!agent.available}
              className={`w-full text-left px-4 py-3 flex items-start gap-3 transition-colors ${
                agent.name === selectedAgent
                  ? 'bg-blue-50'
                  : agent.available
                  ? 'hover:bg-gray-50'
                  : 'opacity-50 cursor-not-allowed'
              }`}
            >
              {/* Checkmark for selected agent */}
              <span className="w-4 h-4 flex-shrink-0 mt-0.5">
                {agent.name === selectedAgent && (
                  <svg
                    className="w-4 h-4 text-blue-600"
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                      clipRule="evenodd"
                    />
                  </svg>
                )}
              </span>

              {/* Agent info */}
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-base">{getAgentIcon(agent.name)}</span>
                  <span className="font-medium text-gray-900">{agent.display_name}</span>
                </div>
                <div className="flex items-center gap-2 mt-1 text-xs text-gray-500">
                  <span
                    className={`inline-block w-2 h-2 rounded-full ${
                      agent.available ? 'bg-green-500' : 'bg-gray-400'
                    }`}
                  />
                  <span>{agent.available ? '可用' : '不可用'}</span>
                  {agent.version && agent.available && (
                    <>
                      <span>·</span>
                      <span>v{agent.version}</span>
                    </>
                  )}
                </div>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 3: Verify component renders correctly**

Start dev server if not running:
```bash
cd frontend && npm run dev
```

Expected: Component file created, no TypeScript errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/AgentDropdown.tsx
git commit -m "feat: create AgentDropdown component"
```

---

### Task 2: Integrate AgentDropdown into Header

**Files:**
- Modify: `frontend/src/components/Header.tsx`

- [ ] **Step 1: Add AgentDropdown import and props**

Modify `frontend/src/components/Header.tsx`:

```typescript
import { formatRelativeTime } from '../utils/time'
import AgentDropdown from './AgentDropdown'
import type { AgentInfo } from '../types'

interface HeaderProps {
  onSync: () => void
  onOpenBlacklist: () => void
  onSearch: (query: string) => void
  onDateFromChange: (date: string) => void
  onDateToChange: (date: string) => void
  onClearDateRange: () => void
  searchQuery: string
  dateFrom: string
  dateTo: string
  syncing: boolean
  lastSyncTime: string | null
  agents: AgentInfo[]
  selectedAgent: string | null
  onAgentSelect: (agentName: string) => void
}

export default function Header({
  onSync,
  onOpenBlacklist,
  onSearch,
  onDateFromChange,
  onDateToChange,
  onClearDateRange,
  searchQuery,
  dateFrom,
  dateTo,
  syncing,
  lastSyncTime,
  agents,
  selectedAgent,
  onAgentSelect
}: HeaderProps) {
```

- [ ] **Step 2: Add AgentDropdown to header layout**

In the same file, modify the button group section (around line 37):

```typescript
          <div className="flex items-center gap-3">
            {lastSyncTime && (
              <span className="text-xs text-gray-500 whitespace-nowrap">
                最近同步：{formatRelativeTime(lastSyncTime)}
              </span>
            )}
            <AgentDropdown
              agents={agents}
              selectedAgent={selectedAgent}
              onSelect={onAgentSelect}
            />
            <button
              onClick={onSync}
              disabled={syncing}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
```

- [ ] **Step 3: Verify TypeScript compilation**

```bash
cd frontend && npm run build
```

Expected: Build succeeds with no errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/Header.tsx
git commit -m "feat: integrate AgentDropdown into Header"
```

---

### Task 3: Add Global Agent State Management in App

**Files:**
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Add selectedAgent state and initialization logic**

Modify `frontend/src/App.tsx`, add state after existing useState declarations (around line 29):

```typescript
  const { pages, groups, total, loading, hasMore, loadMore, refetch } = usePages(searchQuery, dateFrom, dateTo)
  const { agents } = useAgents()
  const [selectedAgent, setSelectedAgent] = useState<string | null>(null)

  // Initialize selected agent from localStorage or first available
  useEffect(() => {
    if (agents.length === 0) return

    const saved = localStorage.getItem('six-sense:selected-agent')
    if (saved && agents.some(a => a.name === saved && a.available)) {
      setSelectedAgent(saved)
    } else {
      const firstAvailable = agents.find(a => a.available)
      if (firstAvailable) {
        setSelectedAgent(firstAvailable.name)
        localStorage.setItem('six-sense:selected-agent', firstAvailable.name)
      }
    }
  }, [agents])
```

- [ ] **Step 2: Add agent selection handler**

Add handler function after the existing handlers (around line 148):

```typescript
  const handleBlacklistPage = async (pageId: number, type: BlacklistType, pattern?: string) => {
    try {
      const response = await api.blacklistPage(pageId, type, pattern)
      await refetch()
      alert(`已加入黑名单，并隐藏 ${response.hidden_pages} 条历史记录`)
    } catch (error) {
      console.error('Add blacklist failed:', error)
      alert('加入黑名单失败，请检查后端服务是否正常运行')
    }
  }

  const handleAgentSelect = (agentName: string) => {
    setSelectedAgent(agentName)
    localStorage.setItem('six-sense:selected-agent', agentName)
  }
```

- [ ] **Step 3: Pass agent props to Header**

Modify the Header component call (around line 164):

```typescript
      <Header
        onSync={handleSync}
        onOpenBlacklist={openBlacklistManager}
        onSearch={handleSearch}
        onDateFromChange={handleDateFromChange}
        onDateToChange={handleDateToChange}
        onClearDateRange={handleClearDateRange}
        searchQuery={searchQuery}
        dateFrom={dateFrom}
        dateTo={dateTo}
        syncing={syncing}
        lastSyncTime={lastSyncTime}
        agents={agents}
        selectedAgent={selectedAgent}
        onAgentSelect={handleAgentSelect}
      />
```

- [ ] **Step 4: Pass selectedAgent to PageList**

Modify the PageList component call (around line 191):

```typescript
            <PageList
              key="page-list"
              groups={groups}
              loading={loading && pages.length === 0}
              agents={agents}
              selectedAgent={selectedAgent}
              onAnalyzeComplete={handleAnalyzeComplete}
              onBlacklist={handleBlacklistPage}
              onLastGroupCollapsedChange={setLastGroupCollapsed}
            />
```

- [ ] **Step 5: Verify TypeScript compilation**

```bash
cd frontend && npm run build
```

Expected: Build succeeds with no errors

- [ ] **Step 6: Commit**

```bash
git add frontend/src/App.tsx
git commit -m "feat: add global agent state management in App"
```

---

### Task 4: Update PageList to Pass selectedAgent to PageCard

**Files:**
- Modify: `frontend/src/components/PageList.tsx`

- [ ] **Step 1: Add selectedAgent prop to PageList**

Modify `frontend/src/components/PageList.tsx`:

```typescript
import type { PageDateGroup, AgentInfo, BlacklistType } from '../types'
import PageCard from './PageCard'
import { useState } from 'react'

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
```

- [ ] **Step 2: Pass selectedAgent to PageCard**

In the same file, modify the PageCard component call (find the map function that renders PageCard):

```typescript
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
```

- [ ] **Step 3: Verify TypeScript compilation**

```bash
cd frontend && npm run build
```

Expected: Build succeeds with no errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/PageList.tsx
git commit -m "feat: pass selectedAgent prop to PageCard"
```

---

### Task 5: Modify PageCard to Use Global Agent

**Files:**
- Modify: `frontend/src/components/PageCard.tsx`

- [ ] **Step 1: Update PageCard props and remove AgentSelector**

Modify `frontend/src/components/PageCard.tsx`:

```typescript
import type { Page, AgentInfo, BlacklistType } from '../types'
import { formatRelativeTime } from '../utils/time'
import { useState } from 'react'
import StreamingText from './StreamingText'
import { useAnalyze } from '../hooks/useAnalyze'

interface PageCardProps {
  page: Page
  agents: AgentInfo[]
  selectedAgent: string | null
  onAnalyzeComplete?: () => void
  onBlacklist?: (pageId: number, type: BlacklistType, pattern?: string) => Promise<void>
}

export default function PageCard({ page, agents, selectedAgent, onAnalyzeComplete, onBlacklist }: PageCardProps) {
  const [showBlacklistMenu, setShowBlacklistMenu] = useState(false)
  const [blacklisting, setBlacklisting] = useState(false)
  const { analyzing, streamText, error, analyze, reset } = useAnalyze()
```

- [ ] **Step 2: Replace agent selection logic with direct analysis**

Remove the old `showAgentSelector` state and handlers, replace with:

```typescript
  const handleAnalyzeClick = () => {
    if (!selectedAgent) {
      alert('请先在页面顶部选择一个 AI Agent')
      return
    }

    const selectedAgentInfo = agents.find(a => a.name === selectedAgent)
    if (!selectedAgentInfo?.available) {
      alert('当前选中的 Agent 不可用，请选择其他 Agent')
      return
    }

    analyze(page.id, selectedAgent, () => {
      onAnalyzeComplete?.()
    })
  }
```

- [ ] **Step 3: Remove AgentSelector component from render**

Find and remove the AgentSelector component at the end of the return statement (should be around line 200+):

```typescript
      {/* Remove this entire block: */}
      {/* {showAgentSelector && (
        <AgentSelector
          agents={agents}
          onSelect={handleAgentSelect}
          onClose={() => setShowAgentSelector(false)}
        />
      )} */}
```

- [ ] **Step 4: Verify TypeScript compilation**

```bash
cd frontend && npm run build
```

Expected: Build succeeds with no errors

- [ ] **Step 5: Test in browser**

Start dev server:
```bash
cd frontend && npm run dev
```

Manual verification:
1. Open http://localhost:5173
2. Verify AgentDropdown appears in header
3. Click dropdown, verify agents list shows
4. Select an agent, verify it persists
5. Click "分析" on a page, verify it uses selected agent without showing dialog
6. Refresh page, verify selected agent persists

Expected: All interactions work as designed

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/PageCard.tsx
git commit -m "feat: use global agent in PageCard, remove AgentSelector dialog"
```

---

### Task 6: Clean Up Unused AgentSelector Component

**Files:**
- Delete: `frontend/src/components/AgentSelector.tsx`

- [ ] **Step 1: Verify AgentSelector is no longer imported**

```bash
cd frontend && grep -r "AgentSelector" src/ --include="*.tsx" --include="*.ts"
```

Expected: No imports found (only the file itself)

- [ ] **Step 2: Delete AgentSelector component**

```bash
rm frontend/src/components/AgentSelector.tsx
```

- [ ] **Step 3: Verify build still works**

```bash
cd frontend && npm run build
```

Expected: Build succeeds with no errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/AgentSelector.tsx
git commit -m "refactor: remove unused AgentSelector component"
```

---

### Task 7: Final Integration Test

**Files:**
- Test: All modified components

- [ ] **Step 1: Start backend and frontend**

```bash
# Terminal 1 - Backend
cd backend-go && GOCACHE=/private/tmp/go-build go run ./cmd/server

# Terminal 2 - Frontend
cd frontend && npm run dev
```

- [ ] **Step 2: Test complete user flow**

Manual test checklist:
1. ✓ Open http://localhost:5173
2. ✓ Verify AgentDropdown shows in header with first available agent selected
3. ✓ Click dropdown, verify both agents show with status indicators
4. ✓ Select different agent, verify selection updates
5. ✓ Refresh page, verify selection persists
6. ✓ Click "分析" on a page, verify analysis starts immediately without dialog
7. ✓ Verify analysis uses the selected agent
8. ✓ Switch agent, analyze another page, verify new agent is used
9. ✓ Test with no available agents (stop backend), verify warning state
10. ✓ Test localStorage persistence across browser sessions

- [ ] **Step 3: Test edge cases**

Edge case checklist:
1. ✓ No agents available - shows warning, disables analysis
2. ✓ Selected agent becomes unavailable - shows warning, prompts to switch
3. ✓ First visit (no localStorage) - auto-selects first available agent
4. ✓ localStorage has invalid agent name - falls back to first available
5. ✓ Click outside dropdown - closes menu
6. ✓ Keyboard navigation works (if implemented)

- [ ] **Step 4: Verify no console errors**

Open browser DevTools Console, verify:
- No React warnings
- No TypeScript errors
- No network errors

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "test: verify global agent selector integration"
```

---

## Self-Review Checklist

**Spec Coverage:**
- ✓ Global agent selector in header
- ✓ Dropdown UI with agent icons and status
- ✓ localStorage persistence
- ✓ Auto-select first available agent
- ✓ Remove per-analysis agent dialog
- ✓ Error handling for no agents / unavailable agents
- ✓ Click outside to close dropdown

**Placeholder Check:**
- ✓ No TBD or TODO items
- ✓ All code blocks are complete
- ✓ All file paths are exact
- ✓ All commands have expected output

**Type Consistency:**
- ✓ AgentInfo type used consistently
- ✓ selectedAgent: string | null used consistently
- ✓ Props match across components
- ✓ localStorage key consistent: 'six-sense:selected-agent'

**Missing Items:**
- Keyboard navigation (listed as optional in spec, not implemented in this plan)
- Accessibility attributes (partially implemented, full keyboard nav deferred)
