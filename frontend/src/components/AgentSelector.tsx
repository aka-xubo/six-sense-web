import type { AgentInfo } from '../types'

interface AgentSelectorProps {
  agents: AgentInfo[]
  onSelect: (agentName: string) => void
  onClose: () => void
}

export default function AgentSelector({ agents, onSelect, onClose }: AgentSelectorProps) {
  const availableAgents = agents.filter(a => a.available)

  if (availableAgents.length === 0) {
    return (
      <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
        <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            没有可用的 AI Agent
          </h3>
          <p className="text-sm text-gray-500 mb-4">
            请确保已安装 Claude Code 或 Codex CLI
          </p>
          <button
            onClick={onClose}
            className="w-full px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200"
          >
            关闭
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
        <h3 className="text-lg font-medium text-gray-900 mb-4">
          选择 AI Agent
        </h3>

        <div className="space-y-2">
          {availableAgents.map((agent) => (
            <button
              key={agent.name}
              onClick={() => onSelect(agent.name)}
              className="w-full text-left px-4 py-3 border border-gray-200 rounded-lg hover:bg-blue-50 hover:border-blue-300 transition-colors"
            >
              <div className="flex items-center justify-between">
                <div>
                  <div className="font-medium text-gray-900">
                    {agent.display_name}
                  </div>
                  {agent.version && (
                    <div className="text-sm text-gray-500">
                      版本 {agent.version}
                    </div>
                  )}
                </div>
                <svg className="h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clipRule="evenodd" />
                </svg>
              </div>
            </button>
          ))}
        </div>

        <button
          onClick={onClose}
          className="w-full mt-4 px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200"
        >
          取消
        </button>
      </div>
    </div>
  )
}
