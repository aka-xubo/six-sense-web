interface StreamingTextProps {
  text: string
}

export default function StreamingText({ text }: StreamingTextProps) {
  return (
    <div className="mt-3 p-3 bg-gray-50 rounded-md border border-gray-200">
      <div className="flex items-start space-x-2">
        <svg className="animate-spin h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <div className="flex-1 min-w-0">
          <div className="text-sm text-gray-700 whitespace-pre-wrap break-words font-mono">
            {text}
            <span className="inline-block w-2 h-4 bg-blue-600 animate-pulse ml-1"></span>
          </div>
        </div>
      </div>
    </div>
  )
}
