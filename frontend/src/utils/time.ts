const WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六']

function toLocalDate(dateString: string): Date {
  return new Date(dateString)
}

function startOfLocalDay(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate())
}

function getCalendarDayDiff(date: Date, now: Date): number {
  const dateStart = startOfLocalDay(date)
  const nowStart = startOfLocalDay(now)
  return Math.floor((nowStart.getTime() - dateStart.getTime()) / (1000 * 60 * 60 * 24))
}

export function getDateGroupKey(dateString: string): string {
  const date = toLocalDate(dateString)
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

export function formatDateGroupHeader(dateString: string): string {
  const date = toLocalDate(dateString)
  const dayDiff = getCalendarDayDiff(date, new Date())
  const fullDate = `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日星期${WEEKDAYS[date.getDay()]}`

  if (dayDiff === 0) {
    return `今天 - ${fullDate}`
  }

  if (dayDiff === 1) {
    return `昨天 - ${fullDate}`
  }

  return fullDate
}

export function formatShortDate(dateString: string): string {
  const date = toLocalDate(dateString)
  return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`
}

/**
 * Format relative time (e.g., "2小时前", "昨天").
 */
export function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSeconds = Math.floor(diffMs / 1000)
  const diffMinutes = Math.floor(diffSeconds / 60)
  const diffHours = Math.floor(diffMinutes / 60)
  const calendarDayDiff = getCalendarDayDiff(date, now)

  if (diffSeconds < 60) {
    return '刚刚'
  } else if (diffMinutes < 60) {
    return `${diffMinutes}分钟前`
  } else if (calendarDayDiff === 0) {
    return `${diffHours}小时前`
  } else if (calendarDayDiff === 1) {
    return '昨天'
  } else if (calendarDayDiff < 7) {
    return `${calendarDayDiff}天前`
  } else if (calendarDayDiff < 30) {
    const weeks = Math.floor(calendarDayDiff / 7)
    return `${weeks}周前`
  } else if (calendarDayDiff < 365) {
    const months = Math.floor(calendarDayDiff / 30)
    return `${months}个月前`
  } else {
    const years = Math.floor(calendarDayDiff / 365)
    return `${years}年前`
  }
}

/**
 * Get time level for color coding
 */
export function getTimeLevel(dateString: string): string {
  const date = toLocalDate(dateString)
  const diffDays = getCalendarDayDiff(date, new Date())

  if (diffDays === 0) return 'today'
  if (diffDays === 1) return 'yesterday'
  if (diffDays < 7) return 'week'
  if (diffDays < 30) return 'month'
  return 'old'
}
