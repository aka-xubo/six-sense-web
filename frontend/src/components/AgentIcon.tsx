const agentIconUrls: Record<string, string> = {
  claude: new URL('../images/claude-color.svg', import.meta.url).href,
  codex: new URL('../images/codex-color.svg', import.meta.url).href
}

const agentLabels: Record<string, string> = {
  claude: 'Claude Code',
  codex: 'OpenAI Codex'
}

interface AgentIconProps {
  agentName: string
  className?: string
  showFallbackText?: boolean
}

export default function AgentIcon({
  agentName,
  className = 'h-4 w-4',
  showFallbackText = false
}: AgentIconProps) {
  const iconUrl = agentIconUrls[agentName]
  const label = agentLabels[agentName] ?? agentName

  if (!iconUrl) {
    return showFallbackText ? (
      <span className="text-xs font-medium text-gray-500">{label}</span>
    ) : (
      <span aria-label={label} title={label} className="inline-block h-2 w-2 rounded-full bg-gray-400" />
    )
  }

  return (
    <img
      src={iconUrl}
      alt={label}
      title={label}
      className={`${className} flex-shrink-0`}
    />
  )
}
