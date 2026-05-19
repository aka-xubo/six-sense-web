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
        className={`inline-flex items-center gap-2 px-3 py-2 border rounded-md text-sm font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 ${
          !hasAvailableAgents || !selectedAgentInfo
            ? 'border-yellow-400 bg-yellow-50 text-yellow-800'
            : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
        }`}
      >
        {selectedAgentInfo ? (
          <>
            <span className="text-base">{getAgentIcon(selectedAgentInfo.name)}</spa       <span>{selectedAgentInfo.display_name}</span>
            <span
              className={`inline-block w-2 h-2 rounded-full ${
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
