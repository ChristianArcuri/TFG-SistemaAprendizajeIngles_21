function avg(nums) {
  if (!nums.length) return 0
  return Math.round(nums.reduce((a, b) => a + b, 0) / nums.length)
}

function uniqueDays(sessions) {
  const days = new Set()
  for (const s of sessions) {
    const d = new Date(s.createdAt)
    days.add(`${d.getFullYear()}-${d.getMonth()}-${d.getDate()}`)
  }
  return days.size
}

function values(sessions, key) {
  return sessions
    .map((s) => s.scores?.[key])
    .filter((n) => typeof n === 'number')
}

/** Promedios por competencia a partir del historial guardado. */
export function computeProgressStats(sessions) {
  if (!sessions?.length) {
    return { streakDays: 0, comprehensionPct: 0, vocabularyPct: 0, communicationPct: 0 }
  }

  return {
    streakDays: uniqueDays(sessions),
    comprehensionPct: avg(values(sessions, 'comprehension')),
    vocabularyPct: avg(values(sessions, 'vocabulary')),
    communicationPct: avg(values(sessions, 'communication')),
  }
}

export function overallFromScores(scores) {
  if (!scores) return 0
  const nums = [scores.comprehension, scores.vocabulary, scores.communication].filter(
    (n) => typeof n === 'number',
  )
  return avg(nums)
}
