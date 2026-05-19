import { useState } from 'react'

interface AnalyzeState {
  analyzing: boolean
  streamText: string
  error: string | null
  completed: boolean
  insights: any | null
}

export function useAnalyze() {
  const [state, setState] = useState<AnalyzeState>({
    analyzing: false,
    streamText: '',
    error: null,
    completed: false,
    insights: null
  })

  const analyze = (pageId: number, agentName: string, onComplete?: () => void) => {
    let receivedServerError = false

    setState({
      analyzing: true,
      streamText: '',
      error: null,
      completed: false,
      insights: null
    })

    // Create EventSource for SSE
    const eventSource = new EventSource(
      `/api/analyze?page_id=${pageId}&agent_name=${agentName}`
    )

    eventSource.addEventListener('status', (e) => {
      const data = JSON.parse(e.data)
      console.log('Status:', data.message)
    })

    eventSource.addEventListener('delta', (e) => {
      const data = JSON.parse(e.data)
      setState(prev => ({
        ...prev,
        streamText: prev.streamText + data.text
      }))
    })

    eventSource.addEventListener('complete', (e) => {
      const data = JSON.parse(e.data)
      setState(prev => ({
        ...prev,
        analyzing: false,
        completed: true,
        insights: data
      }))
      eventSource.close()
      onComplete?.()
    })

    eventSource.addEventListener('analysis_error', (e: MessageEvent) => {
      receivedServerError = true
      const data = e.data ? JSON.parse(e.data) : { error: 'Unknown error' }
      setState(prev => ({
        ...prev,
        analyzing: false,
        error: data.error || 'Analysis failed'
      }))
      eventSource.close()
    })

    eventSource.onerror = () => {
      if (receivedServerError) {
        return
      }

      setState(prev => ({
        ...prev,
        analyzing: false,
        error: 'Connection error'
      }))
      eventSource.close()
    }

    return () => {
      eventSource.close()
    }
  }

  const reset = () => {
    setState({
      analyzing: false,
      streamText: '',
      error: null,
      completed: false,
      insights: null
    })
  }

  return {
    ...state,
    analyze,
    reset
  }
}
