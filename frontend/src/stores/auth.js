import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '../api/auth'

const TOKEN_KEY = 'tfg_sai_token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem(TOKEN_KEY) || '')
  const user = ref(null)
  const profile = ref(null)
  const loading = ref(false)
  const error = ref('')

  const isAuthenticated = computed(() => Boolean(token.value))

  const hasProfile = computed(
    () => Boolean(profile.value?.englishLevel || profile.value?.goal),
  )

  function setSession(authData) {
    token.value = authData.token
    user.value = authData.user
    localStorage.setItem(TOKEN_KEY, authData.token)
  }

  function clearSession() {
    token.value = ''
    user.value = null
    profile.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  async function register(payload) {
    loading.value = true
    error.value = ''
    try {
      const data = await authApi.register(payload)
      setSession(data)
      await loadMe()
      return data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function login(payload) {
    loading.value = true
    error.value = ''
    try {
      const data = await authApi.login(payload)
      setSession(data)
      await loadMe()
      return data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function loadMe() {
    if (!token.value) return
    const data = await authApi.fetchMe()
    user.value = data.user
    profile.value = data.profile || null
    return data
  }

  async function saveProfile(payload) {
    loading.value = true
    error.value = ''
    try {
      profile.value = await authApi.updateProfile(payload)
      return profile.value
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  function logout() {
    clearSession()
  }

  return {
    token,
    user,
    profile,
    loading,
    error,
    isAuthenticated,
    hasProfile,
    register,
    login,
    loadMe,
    saveProfile,
    logout,
  }
})
