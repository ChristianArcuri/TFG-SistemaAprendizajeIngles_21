import api from './client'

export async function fetchCategories() {
  const { data } = await api.get('/api/v1/practice/categories')
  return data.data
}

export async function fetchPrompts(categoryCode) {
  const { data } = await api.get('/api/v1/practice/prompts', {
    params: { category: categoryCode },
  })
  return data.data
}

export async function createSession(payload) {
  const { data } = await api.post('/api/v1/practice/sessions', payload)
  return data.data
}

export async function fetchSessions({ limit = 20, offset = 0 } = {}) {
  const { data } = await api.get('/api/v1/practice/sessions', {
    params: { limit, offset },
  })
  return data.data
}

export async function fetchSession(sessionId) {
  const { data } = await api.get(`/api/v1/practice/sessions/${sessionId}`)
  return data.data
}

export async function analyzeAudio({ audioBlob, categoryCode, promptText }) {
  const form = new FormData()
  form.append('audio', audioBlob, audioBlob.name || 'recording.webm')
  form.append('categoryCode', categoryCode)
  if (promptText) {
    form.append('promptText', promptText)
  }

  const { data } = await api.post('/api/v1/practice/analyze-audio', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000,
  })
  return data.data
}

export async function analyzeText({ categoryCode, promptText, userText, quizCorrect, quizTotal }) {
  const body = { categoryCode, promptText, userText }
  if (quizTotal) {
    body.quizCorrect = quizCorrect
    body.quizTotal = quizTotal
  }
  const { data } = await api.post('/api/v1/practice/analyze-text', body, { timeout: 120000 })
  return data.data
}

export async function submitQuizResult({ correct, total, missedTopics }) {
  const { data } = await api.post('/api/v1/practice/quiz-results', {
    correct,
    total,
    missedTopics,
  })
  return data.data
}

export async function fetchRecommendation() {
  const { data } = await api.get('/api/v1/practice/recommendation')
  return data.data
}
