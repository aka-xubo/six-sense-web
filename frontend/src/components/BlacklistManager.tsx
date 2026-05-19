import type { BlacklistEntry } from '../types'

interface BlacklistManagerProps {
  entries: BlacklistEntry[]
  loading: boolean
  deletingId: number | null
  onDelete: (entry: BlacklistEntry) => void
  onClose: () => void
}

const TYPE_LABELS: Record<BlacklistEntry['type'], string> = {
  url: 'URL',
  domain: '域名',
  path: '域名路径'
}

export default function BlacklistManager({
  entries,
  loading,
  deletingId,
  onDelete,
  onClose
}: BlacklistManagerProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-gray-500 bg-opacity-75">
      <div className="mx-4 w-full max-w-3xl overflow-hidden rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h3 className="text-lg font-medium text-gray-900">黑名单</h3>
          <button
            type="button"
            onClick={onClose}
            className="text-sm font-medium text-gray-500 hover:text-gray-800"
          >
            关闭
          </button>
        </div>

        <div className="max-h-[70vh] overflow-y-auto">
          {loading ? (
            <div className="flex justify-center py-12">
              <svg className="h-6 w-6 animate-spin text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
            </div>
          ) : entries.length === 0 ? (
            <div className="px-6 py-12 text-center text-sm text-gray-500">暂无黑名单规则</div>
          ) : (
            <div className="divide-y divide-gray-100">
              {entries.map((entry) => (
                <div key={entry.id} className="flex items-start justify-between gap-4 px-6 py-4">
                  <div className="min-w-0 flex-1">
                    <div className="mb-1 flex items-center gap-2">
                      <span className="rounded bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700">
                        {TYPE_LABELS[entry.type]}
                      </span>
                      <span className="text-xs text-gray-400">
                        {new Date(entry.created_at).toLocaleString()}
                      </span>
                    </div>
                    <div className="break-all text-sm text-gray-800">{entry.pattern}</div>
                  </div>
                  <button
                    type="button"
                    onClick={() => onDelete(entry)}
                    disabled={deletingId === entry.id}
                    className="flex-shrink-0 rounded px-3 py-1 text-sm font-medium text-red-600 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {deletingId === entry.id ? '移除中...' : '移除'}
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
