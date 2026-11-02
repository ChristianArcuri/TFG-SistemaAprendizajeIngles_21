import { computed, ref } from 'vue'

const STORAGE_KEY = 'tfg_sai_seen_prompts'

function loadSeenMap() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : {}
  } catch {
    return {}
  }
}

function saveSeenMap(map) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(map))
}

/** Banco de prompts con subtipos, navegación y progreso visto (localStorage). */
export function usePracticePrompts(bank, categoryCode) {
  const activeSubtypeId = ref(bank.subtypes[0]?.id ?? null)
  const currentPrompt = ref(null)
  const currentIndex = ref(0)

  const activeSubtype = computed(
    () => bank.subtypes.find((s) => s.id === activeSubtypeId.value) ?? null,
  )

  const pool = computed(() => {
    if (!activeSubtypeId.value) return bank.prompts
    return bank.prompts.filter((p) => p.subtypeId === activeSubtypeId.value)
  })

  const progressLabel = computed(() => {
    if (!currentPrompt.value || !pool.value.length) return ''
    return `${currentIndex.value + 1} de ${pool.value.length}`
  })

  const seenVersion = ref(0)

  function isSeen(promptId) {
    seenVersion.value
    return (loadSeenMap()[categoryCode] ?? []).includes(promptId)
  }

  function markSeen(promptId) {
    if (!promptId) return
    const map = loadSeenMap()
    const list = map[categoryCode] ?? []
    if (!list.includes(promptId)) {
      map[categoryCode] = [...list, promptId]
      saveSeenMap(map)
      seenVersion.value++
    }
  }

  function setCurrent(prompt) {
    if (!prompt) return
    currentPrompt.value = prompt
    currentIndex.value = Math.max(0, pool.value.findIndex((p) => p.id === prompt.id))
  }

  function pickRandomFrom(list) {
    if (!list.length) return null
    const unseen = list.filter((p) => !isSeen(p.id))
    const source = unseen.length ? unseen : list
    return source[Math.floor(Math.random() * source.length)]
  }

  function init() {
    const first = pickRandomFrom(pool.value) ?? pool.value[0] ?? null
    setCurrent(first)
  }

  function selectSubtype(subtypeId) {
    activeSubtypeId.value = subtypeId
    const first = pickRandomFrom(pool.value) ?? pool.value[0] ?? null
    setCurrent(first)
    return first
  }

  function nextPrompt() {
    const list = pool.value
    if (!list.length) return null
    const nextIdx = (currentIndex.value + 1) % list.length
    setCurrent(list[nextIdx])
    return currentPrompt.value
  }

  function completeCurrent() {
    if (currentPrompt.value) markSeen(currentPrompt.value.id)
  }

  return {
    subtypes: bank.subtypes,
    activeSubtypeId,
    activeSubtype,
    currentPrompt,
    pool,
    progressLabel,
    isSeen,
    markSeen,
    init,
    selectSubtype,
    nextPrompt,
    completeCurrent,
  }
}
