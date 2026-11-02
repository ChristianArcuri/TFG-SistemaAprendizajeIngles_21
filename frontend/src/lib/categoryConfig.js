/** Módulos del prototipo TFG: entrevista RRHH, daily meeting y quiz IT. */
export const categoryConfig = {
  hr_interview: {
    mode: 'speaking',
    routeName: 'practice-hr-interview',
    cta: 'Practicar entrevista',
    btnVariant: 'primary',
  },
  daily_meeting: {
    mode: 'daily',
    routeName: 'practice-daily',
    cta: 'Empezar daily',
    btnVariant: 'primary',
  },
  it_quiz: {
    mode: 'quiz',
    routeName: 'practice-quiz',
    cta: 'Resolver quiz',
    btnVariant: 'primary',
  },
}

export function getCategoryConfig(code) {
  return (
    categoryConfig[code] || {
      mode: 'speaking',
      routeName: 'practice-hr-interview',
      cta: 'Practicar entrevista',
      btnVariant: 'primary',
    }
  )
}

export function startRouteForCategory(code) {
  const cfg = getCategoryConfig(code)
  if (cfg.routeName === 'practice-quiz') {
    return { name: 'practice-quiz' }
  }
  if (cfg.routeName === 'practice-daily') {
    return { name: 'practice-daily', params: { categoryCode: code } }
  }
  if (cfg.routeName === 'practice-hr-interview') {
    return { name: 'practice-hr-interview' }
  }
  return { name: cfg.routeName, params: { categoryCode: code } }
}
