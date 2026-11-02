import { computed, ref } from 'vue'

const STORAGE_KEY = 'tfg_sai_seen_prompts'

function loadSeen(categoryCode) {
  try {
    const map = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}')
    return map[categoryCode] ?? []
  } catch {
    return []
  }
}

function markSeen(categoryCode, ids) {
  const raw = localStorage.getItem(STORAGE_KEY)
  const map = raw ? JSON.parse(raw) : {}
  const prev = new Set(map[categoryCode] ?? [])
  ids.forEach((id) => prev.add(id))
  map[categoryCode] = [...prev]
  localStorage.setItem(STORAGE_KEY, JSON.stringify(map))
}

function shuffle(arr) {
  const copy = [...arr]
  for (let i = copy.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[copy[i], copy[j]] = [copy[j], copy[i]]
  }
  return copy
}

/** Sesion de quiz: N preguntas del banco, filtro por tema, progreso y respuestas. */
export function useQuizSession(bank) {
  const categoryCode = bank.categoryCode
  const sessionSize = bank.sessionSize ?? 10

  const activeSubtypeId = ref(bank.subtypes[0]?.id ?? null)
  const sessionQuestions = ref([])
  const sessionIndex = ref(0)
  const selected = ref(null)
  const answered = ref(false)
  const correctCount = ref(0)
  const answersLog = ref([])
  const sessionStarted = ref(false)
  const sessionFinished = ref(false)

  const pool = computed(() => {
    if (!activeSubtypeId.value) return bank.questions
    return bank.questions.filter((q) => q.subtypeId === activeSubtypeId.value)
  })

  const question = computed(() => sessionQuestions.value[sessionIndex.value] ?? null)
  const isLastInSession = computed(
    () => sessionIndex.value >= sessionQuestions.value.length - 1,
  )
  const progressPct = computed(() => {
    if (!sessionQuestions.value.length) return 0
    return Math.round(((sessionIndex.value + (answered.value ? 1 : 0)) / sessionSize) * 100)
  })

  function pickSessionQuestions() {
    const list = pool.value
    if (!list.length) return []
    const seen = new Set(loadSeen(categoryCode))
    const unseen = list.filter((q) => !seen.has(q.id))
    const source = unseen.length >= sessionSize ? unseen : list
    return shuffle(source).slice(0, Math.min(sessionSize, source.length))
  }

  function startSession(subtypeId = activeSubtypeId.value) {
    activeSubtypeId.value = subtypeId
    sessionQuestions.value = pickSessionQuestions()
    sessionIndex.value = 0
    selected.value = null
    answered.value = false
    correctCount.value = 0
    answersLog.value = []
    sessionStarted.value = sessionQuestions.value.length > 0
    sessionFinished.value = false
  }

  function selectSubtype(subtypeId) {
    startSession(subtypeId)
  }

  function chooseOption(index) {
    if (answered.value || !question.value) return
    selected.value = index
    answered.value = true
    const ok = index === question.value.correctIndex
    if (ok) correctCount.value += 1
    answersLog.value.push({
      id: question.value.id,
      subtypeId: question.value.subtypeId,
      situation: question.value.situationEs,
      chosen: question.value.options[index],
      correct: question.value.options[question.value.correctIndex],
      ok,
      skipped: false,
    })
  }

  function advanceSession() {
    if (sessionIndex.value >= sessionQuestions.value.length - 1) {
      finishSession()
      return
    }
    sessionIndex.value += 1
    selected.value = null
    answered.value = false
  }

  function skipQuestion() {
    if (!question.value || sessionFinished.value) return
    if (!answered.value) {
      answersLog.value.push({
        id: question.value.id,
        subtypeId: question.value.subtypeId,
        situation: question.value.situationEs,
        chosen: '(skipped)',
        correct: question.value.options[question.value.correctIndex],
        ok: false,
        skipped: true,
      })
    }
    advanceSession()
  }

  function nextAfterAnswer() {
    if (!answered.value) return
    advanceSession()
  }

  function finishSession() {
    sessionFinished.value = true
    const ids = answersLog.value.map((a) => a.id).filter(Boolean)
    if (ids.length) markSeen(categoryCode, ids)
  }

  function buildSummaryText() {
    const total = sessionQuestions.value.length
    const lines = answersLog.value.map((a, i) => {
      const tag = a.skipped ? 'SKIP' : a.ok ? 'OK' : 'X'
      return `Q${i + 1}: ${tag} — chose "${a.chosen}"`
    })
    return `IT English quiz score: ${correctCount.value}/${total}.\n${lines.join('\n')}`
  }

  return {
    subtypes: bank.subtypes,
    sessionSize,
    activeSubtypeId,
    sessionQuestions,
    sessionIndex,
    sessionStarted,
    sessionFinished,
    question,
    selected,
    answered,
    correctCount,
    answersLog,
    isLastInSession,
    progressPct,
    pool,
    startSession,
    selectSubtype,
    chooseOption,
    nextAfterAnswer,
    skipQuestion,
    buildSummaryText,
  }
}
