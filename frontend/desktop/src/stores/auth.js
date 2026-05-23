import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, configureAuth } from '../api/http'

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref(null)
  const user = ref(null)
  const loading = ref(true)

  const isLoggedIn = computed(() => !!user.value)

  function setSession(token, u) {
    accessToken.value = token
    user.value = u
  }

  async function refresh() {
    try {
      const res = await api.refresh()
      accessToken.value = res.accessToken
      user.value = res.user
      return true
    } catch {
      accessToken.value = null
      user.value = null
      return false
    }
  }

  async function bootstrap() {
    loading.value = true
    configureAuth(() => accessToken.value, refresh)
    await refresh()
    loading.value = false
  }

  async function login(form) {
    const res = await api.login(form)
    setSession(res.accessToken, res.user)
  }

  async function register(form) {
    const res = await api.register(form)
    setSession(res.accessToken, res.user)
  }

  async function logout() {
    try {
      await api.logout()
    } finally {
      accessToken.value = null
      user.value = null
    }
  }

  return { accessToken, user, loading, isLoggedIn, bootstrap, login, register, logout, refresh }
})
