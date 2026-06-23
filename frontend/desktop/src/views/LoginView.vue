<template>
  <div class="login-page">
    <button class="theme-switch icon-btn icon-btn--ghost" type="button" @click="onToggleTheme">
      <AppIcon name="palette" size="sm" />
      <span>{{ themeLabel }}</span>
    </button>

    <form class="login-card" @submit.prevent="submit">
      <div class="login-brand">
        <span class="login-brand-icon" aria-hidden="true">
          <AppIcon name="brand" size="xl" />
        </span>
        <div>
          <h1>Smart Ledger</h1>
          <p class="login-tagline">智能账本 · 多人协作 · 链上锚定</p>
        </div>
      </div>

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
          <button type="button" class="captcha-btn" title="点击刷新" @click="loadCaptcha">
            <img v-if="captchaImg" :src="captchaImg" alt="captcha" />
            <AppIcon v-else name="refresh" size="sm" />
          </button>
        </div>
      </div>

      <button class="btn-primary login-submit" type="submit" :disabled="loading">
        {{ loading ? (mode === 'login' ? '登录中…' : '注册中…') : (mode === 'login' ? '登录' : '注册并登录') }}
      </button>

      <p v-if="mode === 'login'" class="login-hint">首次使用请用默认账号 <strong>admin</strong> / <strong>admin123</strong>，或切换到「注册」创建新用户。</p>
    </form>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppIcon from '../components/AppIcon.vue'
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
  'classic-dark': '深蓝',
  'deep-dark': '深黑',
}

const themeLabel = computed(() => {
  const name = THEME_NAMES[theme.value] || '深蓝'
  return `主题 · ${name}`
})

async function loadCaptcha() {
  try {
    const c = await api.captcha()
    captchaId.value = c.captchaId
    captchaImg.value = c.image.startsWith('data:') ? c.image : `data:image/png;base64,${c.image}`
    form.captchaCode = ''
  } catch (e) {
    captchaImg.value = ''
    error.value = e instanceof ApiError ? e.message : '验证码加载失败，请稍后重试'
  }
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
  if (!captchaId.value) {
    error.value = '验证码未加载，请点击刷新后重试'
    await loadCaptcha()
    return
  }
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
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
}

.login-card {
  width: 440px;
  max-width: 100%;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 2rem;
  box-shadow: var(--shadow-lg);
}

.login-brand {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.login-brand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 3.25rem;
  height: 3.25rem;
  border-radius: 14px;
  background: linear-gradient(145deg, var(--accent-soft), color-mix(in srgb, var(--accent) 20%, transparent));
  color: var(--accent);
  flex-shrink: 0;
}

.login-brand h1 {
  margin: 0;
  font-size: 1.35rem;
}

.login-tagline {
  margin: 0.25rem 0 0;
  font-size: 0.8rem;
  color: var(--text-muted);
}

.tabs {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1.25rem;
  padding: 0.25rem;
  background: var(--bg);
  border-radius: 10px;
  border: 1px solid var(--border);
}

.tabs button {
  flex: 1;
  padding: 0.5rem;
  border-radius: 8px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-muted);
  font-weight: 600;
}

.tabs button.active {
  background: var(--accent-soft);
  border-color: color-mix(in srgb, var(--accent) 35%, transparent);
  color: var(--accent);
}

.cap-row {
  display: flex;
  gap: 0.5rem;
  align-items: stretch;
  max-width: var(--field-max);
}

.cap-row input {
  flex: 1;
}

.captcha-btn {
  padding: 0;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--bg);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 6.5rem;
  color: var(--text-muted);
}

.captcha-btn img {
  height: 40px;
  width: auto;
  border-radius: 8px;
  display: block;
}

.captcha-btn:hover {
  border-color: var(--accent);
}

.login-submit {
  width: 100%;
  margin-top: 0.25rem;
}

.login-hint {
  margin: 1rem 0 0;
  font-size: 0.78rem;
  line-height: 1.5;
  color: var(--text-muted);
  text-align: center;
}

.login-hint strong {
  color: var(--text);
  font-weight: 600;
}

.theme-switch {
  position: fixed;
  top: 16px;
  right: 16px;
  z-index: 10;
}
</style>
