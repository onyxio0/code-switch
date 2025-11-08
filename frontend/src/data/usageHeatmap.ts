import type { HeatmapStat } from '../services/logs'

export type UsageHeatmapDay = {
  label: string
  dateKey: string
  requests: number
  inputTokens: number
  outputTokens: number
  reasoningTokens: number
  cost: number
  intensity: number
}

export type UsageHeatmapWeek = UsageHeatmapDay[]

export const DAYS_PER_WEEK = 7
export const DEFAULT_USAGE_WEEKS = 30
const LEVELS = 4

const dateLabelFormatter = new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric' })

const toDateKey = (date: Date) => {
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${year}-${month}-${day}`
}

const clampWeeks = (weeks?: number) => (weeks && weeks > 0 ? weeks : DEFAULT_USAGE_WEEKS)

const intensityForCount = (count: number, maxCount: number) => {
  if (count <= 0 || maxCount <= 0) return 0
  const ratio = count / maxCount
  return Math.min(LEVELS, Math.max(1, Math.ceil(ratio * LEVELS)))
}

const calculateStartDate = (totalDays: number) => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  today.setDate(today.getDate() - totalDays + 1)
  return today
}

export const generateFallbackUsageHeatmap = (weeks = DEFAULT_USAGE_WEEKS): UsageHeatmapWeek[] => {
  const normalizedWeeks = clampWeeks(weeks)
  const totalDays = normalizedWeeks * DAYS_PER_WEEK
  const baseDate = calculateStartDate(totalDays)
  const weeksData: UsageHeatmapWeek[] = []

  for (let w = 0; w < normalizedWeeks; w++) {
    const week: UsageHeatmapWeek = []
    for (let d = 0; d < DAYS_PER_WEEK; d++) {
      const offset = w * DAYS_PER_WEEK + d
      const date = new Date(baseDate)
      date.setDate(baseDate.getDate() + offset)
      const index = w * DAYS_PER_WEEK + d
      const hash = Math.abs(Math.sin(index * 12.9898 + date.getDate()) * 1000)
      const requests = Math.floor(hash % 20)
      const inputTokens = requests * 120 + Math.floor(hash % 40)
      const outputTokens = Math.floor(inputTokens * 0.6)
      const reasoningTokens = Math.floor(requests * 15)
      const intensity = Math.min(LEVELS, Math.floor(requests / 5))
      const cost = Number(((inputTokens + outputTokens) * 0.000002).toFixed(6))
      week.push({
        label: dateLabelFormatter.format(date),
        dateKey: toDateKey(date),
        requests,
        inputTokens,
        outputTokens,
        reasoningTokens,
        cost,
        intensity,
      })
    }
    weeksData.push(week)
  }

  return weeksData
}

export const buildUsageHeatmapMatrix = (
  stats: HeatmapStat[] = [],
  weeks = DEFAULT_USAGE_WEEKS,
): UsageHeatmapWeek[] => {
  const normalizedWeeks = clampWeeks(weeks)
  const totalDays = normalizedWeeks * DAYS_PER_WEEK
  const startDate = calculateStartDate(totalDays)
  type StatBucket = {
    requests: number
    inputTokens: number
    outputTokens: number
    reasoningTokens: number
    cost: number
  }
  const statsMap = new Map<string, StatBucket>()

  stats.forEach((stat) => {
    if (!stat) return
    const key = stat.day?.slice(0, 10)
    if (!key) return
    statsMap.set(key, {
      requests: Number(stat.total_requests) || 0,
      inputTokens: Number(stat.input_tokens) || 0,
      outputTokens: Number(stat.output_tokens) || 0,
      reasoningTokens: Number(stat.reasoning_tokens) || 0,
      cost: Number(stat.total_cost) || 0,
    })
  })

  let maxCount = 0
  statsMap.forEach((bucket) => {
    if (bucket.requests > maxCount) {
      maxCount = bucket.requests
    }
  })

  const weeksData: UsageHeatmapWeek[] = []
  for (let w = 0; w < normalizedWeeks; w++) {
    const week: UsageHeatmapWeek = []
    for (let d = 0; d < DAYS_PER_WEEK; d++) {
      const offset = w * DAYS_PER_WEEK + d
      const date = new Date(startDate)
      date.setDate(startDate.getDate() + offset)
      const key = toDateKey(date)
      const bucket = statsMap.get(key) ?? {
        requests: 0,
        inputTokens: 0,
        outputTokens: 0,
        reasoningTokens: 0,
        cost: 0,
      }
      week.push({
        label: dateLabelFormatter.format(date),
        dateKey: key,
        requests: bucket.requests,
        inputTokens: bucket.inputTokens,
        outputTokens: bucket.outputTokens,
        reasoningTokens: bucket.reasoningTokens,
        cost: bucket.cost,
        intensity: intensityForCount(bucket.requests, maxCount),
      })
    }
    weeksData.push(week)
  }

  return weeksData
}
