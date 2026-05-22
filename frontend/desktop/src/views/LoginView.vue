<template>
  <div class="login-page">
    <form class="login-card" @submit.prevent="submit">
      <h1>Smart Ledger</h1>
      <p class="sub">登录控制台</p>
      <div v-if="error" class="alert alert-error">{{ error }}</div>
      <div class="form-row"><label>用户名</label><input v-model="form.username" required /></div>
      <div class="form-row"><label>密码</label><input v-model="form.password" type="password" required /></div>
      <div class="form-row">
        <label>验证码</label>
        <div class="cap-row">
          <input v-model="form.captchaCode" required placeholder="输入图中字符" />
          <img v-if="captchaImg" :src="captchaImg" alt="captcha" @click="loadCaptcha" title="点击刷新" />
        </div>
      </div>
      <button class="btn-primary" style="width:100%" :disabled="loading">{{ loading ? '登录中…' : '登录' }}</button>
      <p class="hint">默认 admin / admin123</p>
    </form>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const error = ref('')
const captchaId = ref('')
const captchaImg = ref('')
const form = reactive({ username: 'admin', password: '', captchaCode: '' })

async function loadCaptcha() {
  const c = await api.captcha()
  captchaId.value = c.captchaId
  captchaImg.value = c.image.startsWith('data:') ? c.image : `data:image/png;base64,${c.image}`
  form.captchaCode = ''
}

onMounted(async () => {
  await auth.bootstrap()
  if (auth.isLoggedIn) router.replace('/')
  else loadCaptcha()
})

async function submit() {
  loading.value = true
  error.value = ''
  try {
    await auth.login({
      username: form.username,
      password: form.password,
      captchaId: captchaId.value,
      captchaCode: form.captchaCode,
    })
    router.replace('/')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '登录失败'
    await loadCaptcha()
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page { min-height: 100vh; display: flex; align-items: center; justify-content: center; }
.login-card { width: 380px; background: var(--bg-card); border: 1px solid var(--border); border-radius: 12px; padding: 2rem; }
.sub { color: var(--text-muted); margin: 0 0 1.25rem; }
.cap-row { display: flex; gap: 0.5rem; align-items: center; }
.cap-row img { height: 40px; border-radius: 6px; cursor: pointer; border: 1px solid var(--border); }
.hint { text-align: center; font-size: 0.75rem; color: var(--text-muted); margin-top: 1rem; }
</style>
