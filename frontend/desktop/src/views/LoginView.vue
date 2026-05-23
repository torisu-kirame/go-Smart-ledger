<template>
  <div class="login-page">
    <button class="theme-switch" @click="onToggleTheme">{{ themeLabel }}</button>
    <form class="login-card" @submit.prevent="submit">
      <h1>Smart Ledger</h1>
      <p class="sub">用户端 · 登录与注册</p>

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
      <p class="hint">登录后可在「个人中心」管理资料，在「好友」页通过用户 ID 添加好友</p>
    </form>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'
import { getTheme, toggleTheme } from '../utils/theme'

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

const themeLabel = computed(() => (theme.value === 'light' ? '深色' : '浅色'))

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
  theme.value = toggleTheme(theme.value)
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
.sub { color: var(--text-muted); margin: 0 0 1rem; }
.tabs { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
.tabs button { flex: 1; padding: 0.5rem; border-radius: 8px; background: var(--bg); border: 1px solid var(--border); color: var(--text-muted); }
.tabs button.active { background: rgba(61,139,253,.2); border-color: var(--accent); color: var(--accent); }
.cap-row { display: flex; gap: 0.5rem; align-items: center; }
.cap-row img { height: 40px; border-radius: 6px; cursor: pointer; border: 1px solid var(--border); }
.hint { text-align: center; font-size: 0.75rem; color: var(--text-muted); margin-top: 1rem; line-height: 1.4; }
.theme-switch { position: fixed; top: 16px; right: 16px; z-index: 10; }
</style>
