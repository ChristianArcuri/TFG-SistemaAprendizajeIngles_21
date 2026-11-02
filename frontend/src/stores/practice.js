import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as practiceApi from '../api/practice'

export const usePracticeStore = defineStore('practice', () => {
  const categories = ref([])
  const sessions = ref([])
  const lastResult = ref(null)
  const loading = ref(false)
  const error = ref('')

  async function loadCategories() {
    loading.value = true
    error.value = ''
    try {
      categories.value = await practiceApi.fetchCategories()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function loadHistory() {
    loading.value = true
    error.value = ''
    try {
      const page = await practiceApi.fetchSessions({ limit: 50 })
      sessions.value = page.items || []
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function analyzeAudio({ audioBlob, categoryCode, promptText }) {
    loading.value = true
    error.value = ''
    try {
      const result = await practiceApi.analyzeAudio({
        audioBlob,
        categoryCode,
        promptText,
      })
      lastResult.value = result
      return result
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function analyzeText({ categoryCode, promptText, userText, quizCorrect, quizTotal }) {
    loading.value = true
    error.value = ''
    try {
      const result = await practiceApi.analyzeText({
        categoryCode,
        promptText,
        userText,
        quizCorrect,
        quizTotal,
      })
      lastResult.value = result
      return result
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  function categoryByCode(code) {
    return categories.value.find((c) => c.code === code)
  }

  return {
    categories,
    sessions,
    lastResult,
    loading,
    error,
    loadCategories,
    loadHistory,
    analyzeAudio,
    analyzeText,
    categoryByCode,
  }
})
