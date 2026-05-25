export function formatVisitCount(count: number): string {
  return count.toLocaleString('zh-CN')
}

type VisitCountScope = 'page' | 'domain'

export function getVisitCountBadgeClass(count: number, scope: VisitCountScope = 'page'): string {
  const base = 'inline-flex h-5 items-center rounded border px-1.5 text-xs'
  const thresholds = scope === 'domain'
    ? { active: 5, notable: 15, frequent: 30 }
    : { active: 2, notable: 5, frequent: 10 }

  if (count >= thresholds.frequent) {
    return `${base} border-rose-600 bg-rose-600 font-semibold text-white shadow-sm`
  }

  if (count >= thresholds.notable) {
    return `${base} border-rose-200 bg-rose-50 font-semibold text-rose-700`
  }

  if (count >= thresholds.active) {
    return `${base} border-orange-200 bg-orange-50 font-semibold text-orange-700`
  }

  return `${base} border-gray-200 bg-white font-medium text-gray-500`
}
