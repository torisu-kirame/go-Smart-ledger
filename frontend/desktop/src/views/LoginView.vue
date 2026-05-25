<template>
  <div class="login-page">
    <button class="theme-switch" @click="onToggleTheme">{{ themeLabel }}</button>
    <form class="login-card" @submit.prevent="submit">
      <h1>Smart Ledger</h1>

      <div class="tabs">
        <button type="button" :class="{ active: mode === 'login' }" @click="mode = 'login'">登录</button>
        <button type="button" :class="{ active: mode === 'register' }" @click="mode = 'register'">注册</button>
      </div>

      <div v-if="error" class="alert alert-error">{{ error }}</div>
      <div v-if="success" class="alert alert-success">{{ success }}</div>

      <div class="form-row"><label>用户名</label><input v-model="form.username" required autocomplete="username" /></div>
      <div class="form-row"><label>密码</label><input v-model="form.password" type="password" required :autocomplete="mode === 'login' ? 'current-password' : 'new-password'" /></div>
      <div v-if="mode === 'register'" class="form-row"><label>确认密码</label><input v-model="form.confirm" type="password" required autocomplete="new-password" /></div>

      <div class="form-row">
        <label>验证码</label>
        <div class="cap-row">
          <input v-model="form.captchaCode" required placeholder="输入图中字符" />
          <img v-if="captchaImg" :src="captchaImg" alt="captcha" @click="loadCaptcha" title="点击刷新" />
        </div>
      </div>

      <button class="btn-primary" style="width:100%" :disabled="loading">
        {{ loading ? (mode === 'login' ? '登录中…' : '注册中…') : (mode === 'login' ? '登录' : '注册并登录') }}
      </button>
    </form>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'
import { cycleTheme, getTheme } from '../utils/theme'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const error = ref('')
const success = ref('')
const mode = ref('login')
const theme = ref(getTheme())
const captchaId = ref('')
const captchaImg = ref('')
const form = reactive({ username: '', password: '', confirm: '', captchaCode: '' })

const THEME_NAMES = {
  'classic-light': '经典白',
  'classic-dark': '经典黑',
  'deep-dark': '深黑',
}

const themeLabel = computed(() => {
  const name = THEME_NAMES[theme.value] || '经典黑'
  return `主题 · ${name}`
})

async function loadCaptcha() {
  const c = await api.captcha()
  captchaId.value = c.captchaId
  captchaImg.value = c.image.startsWith('data:') ? c.image : `data:image/png;base64,${c.image}`
  form.captchaCode = ''
}

watch(mode, () => {
  error.value = ''
  success.value = ''
})

onMounted(async () => {
  await auth.bootstrap()
  if (auth.isLoggedIn) router.replace('/')
  else loadCaptcha()
})

function onToggleTheme() {
  theme.value = cycleTheme(theme.value)
}

async function submit() {
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    if (mode.value === 'register') {
      if (form.password !== form.confirm) {
        error.value = '两次密码不一致'
        return
      }
      if (form.password.length < 6) {
        error.value = '密码至少 6 位'
        return
      }
      await auth.register({
        username: form.username,
        password: form.password,
        captchaId: captchaId.value,
        captchaCode: form.captchaCode,
      })
      success.value = `注册成功，用户 ID：${auth.user?.id}`
      setTimeout(() => router.replace('/'), 600)
    } else {
      await auth.login({
        username: form.username,
        password: form.password,
        captchaId: captchaId.value,
        captchaCode: form.captchaCode,
      })
      router.replace('/')
    }
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : (mode.value === 'login' ? '登录失败' : '注册失败')
    await loadCaptcha()
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page { min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 1rem; }
.login-card { width: 420px; max-width: 100%; background: var(--bg-card); border: 1px solid var(--border); border-radius: 16px; padding: 2rem; box-shadow: var(--shadow-lg); }
.tabs { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
.tabs button { flex: 1; padding: 0.5rem; border-radius: 8px; background: var(--bg); border: 1px solid var(--border); color: var(--text-muted); }
.tabs button.active { background: rgba(61,139,253,.2); border-color: var(--accent); color: var(--accent); }
.cap-row { display: flex; gap: 0.5rem; align-items: center; }
.cap-row img { height: 40px; border-radius: 6px; cursor: pointer; border: 1px solid var(--border); }
.theme-switch { position: fixed; top: 16px; right: 16px; z-index: 10; }
</style>
