import { useState, useRef, useEffect } from 'react'
import type { AgentInfo } from '../types'
import AgentIcon from './AgentIcon'

interface AgentDropdownProps {
  agents: AgentInfo[]
  selectedAgent: string | null
  onSelect: (agentName: string) => void
  loading?: boolean
  error?: string | null
  onRefresh?: () => void
}

export default function AgentDropdown({
  agents,
  selectedAgent,
  onSelect,
  loading = false,
  error = null,
  onRefresh
}: AgentDropdownProps) {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const buttonRef = useRef<HTMLButtonElement>(null)
  const [dropdownPosition, setDropdownPosition] = useState({ top: 0, right: 0 })

  // Update dropdown position when opened
  useEffect(() => {
    if (isOpen && buttonRef.current) {
      const rect = buttonRef.current.getBoundingClientRect()
      setDropdownPosition({
        top: rect.bottom + 8,
        right: window.innerWidth - rect.right
      })
    }
  }, [isOpen])

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
  const selectedAgentUnavailable = Boolean(selectedAgentInfo && !selectedAgentInfo.available)

  const handleToggle = () => {
    setIsOpen(!isOpen)
  }

  const handleSelect = (agentName: string) => {
    const agent = agents.find(item => item.name === agentName)
    if (!agent?.available) {
      return
    }

    onSelect(agentName)
    setIsOpen(false)
  }

  const getButtonStateClass = () => {
    if (error) return 'border-red-300 bg-red-50 text-red-800'
    if (loading) return 'border-gray-300 bg-white text-gray-500'
    if (!hasAvailableAgents || !selectedAgentInfo || selectedAgentUnavailable) {
      return 'border-yellow-400 bg-yellow-50 text-yellow-800'
    }
    return 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
  }

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Dropdown Button */}
      <button
        ref={buttonRef}
        onClick={handleToggle}
        aria-label="选择 AI Agent"
        className={`inline-flex min-w-[180px] items-center justify-between gap-2 px-3 py-2 border rounded-md text-sm font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 ${getButtonStateClass()}`}
      >
        <span className="inline-flex min-w-0 items-center gap-2">
          {loading ? (
            <>
              <svg className="h-4 w-4 animate-spin text-gray-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>检测 Agent...</span>
            </>
          ) : error ? (
            <>
              <span className="text-base">!</span>
              <span>检测失败</span>
            </>
          ) : selectedAgentInfo ? (
            <>
              <AgentIcon agentName={selectedAgentInfo.name} className="h-4 w-4" />
              <span className="truncate">{selectedAgentInfo.display_name}</span>
              {selectedAgentUnavailable && <span className="text-xs">不可用</span>}
            </>
          ) : (
            <>
              <span className="text-base">!</span>
              <span>未检测到 Agent</span>
            </>
          )}
          {!loading && !error && selectedAgentInfo && (
            <span
              className={`inline-block h-2 w-2 flex-shrink-0 rounded-full ${
                selectedAgentInfo.available ? 'bg-green-500' : 'bg-gray-400'
              }`}
            />
          )}
        </span>
        <svg
          className={`h-4 w-4 flex-shrink-0 transition-transform ${isOpen ? 'rotate-180' : ''}`}
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
          className="fixed w-64 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-[100]"
          style={{
            top: `${dropdownPosition.top}px`,
            right: `${dropdownPosition.right}px`
          }}
        >
          <div className="border-b border-gray-100 px-4 py-3">
            <div className="flex items-center justify-between gap-3">
              <div>
                <div className="text-sm font-medium text-gray-900">AI Agent</div>
                <div className="mt-0.5 text-xs text-gray-500">
                  {loading ? '正在检测本地 CLI' : error ? '无法完成检测' : `${agents.filter(agent => agent.available).length} 个可用`}
                </div>
              </div>
              {onRefresh && (
                <button
                  type="button"
                  onClick={(event) => {
                    event.stopPropagation()
                    onRefresh()
                  }}
                  disabled={loading}
                  className="rounded-md border border-gray-300 px-2 py-1 text-xs font-medium text-gray-600 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  刷新
                </button>
              )}
            </div>
            {error && (
              <div className="mt-2 rounded-md bg-red-50 px-2 py-1.5 text-xs text-red-700">
                {error}
              </div>
            )}
          </div>
          {loading ? (
            <div className="px-4 py-4 text-sm text-gray-500">
              正在重新检测本地 Agent 状态...
            </div>
          ) : agents.map((agent) => (
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
                    <AgentIcon agentName={agent.name} className="h-5 w-5" />
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
                  {!agent.available && (
                    <div className="mt-1 text-xs text-gray-400">
                      请确认命令已安装且可正常运行
                    </div>
                  )}
                </div>
              </button>
            ))}
          {!loading && agents.length === 0 && (
            <div className="px-4 py-3 text-sm text-gray-500">
              没有返回任何 Agent 配置
            </div>
          )}
        </div>
      )}
    </div>
  )
}
