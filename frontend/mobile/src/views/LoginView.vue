<template>
  <div class="page page--login">
    <div class="login-hero">
      <div class="brand-icon">📒</div>
      <h1>Smart Ledger</h1>
      <p>智能账本 · 移动版</p>
    </div>

    <van-form class="login-form" @submit="submit">
      <van-tabs v-model:active="mode" shrink animated swipeable>
        <van-tab title="登录" name="login" />
        <van-tab title="注册" name="register" />
      </van-tabs>

      <van-cell-group inset>
        <van-field
          v-model="form.username"
          label="用户名"
          placeholder="请输入用户名"
          required
          autocomplete="username"
        />
        <van-field
          v-model="form.password"
          type="password"
          label="密码"
          placeholder="请输入密码"
          required
          autocomplete="current-password"
        />
        <van-field
          v-if="mode === 'register'"
          v-model="form.confirm"
          type="password"
          label="确认密码"
          placeholder="再次输入密码"
          required
          autocomplete="new-password"
        />
        <van-field v-model="form.captchaCode" label="验证码" placeholder="输入图中字符" required>
          <template #button>
            <img v-if="captchaImg" class="captcha-img" :src="captchaImg" alt="captcha" @click="loadCaptcha" />
            <van-button v-else size="small" type="primary" plain @click="loadCaptcha">获取</van-button>
          </template>
        </van-field>
      </van-cell-group>

      <div class="login-actions">
        <van-button round block type="primary" native-type="submit" :loading="loading">
          {{ mode === 'login' ? '登录' : '注册并登录' }}
        </van-button>
      </div>
    </van-form>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showFailToast, showSuccessToast } from 'vant'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'

defineOptions({ name: 'LoginView' })

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const mode = ref('login')
const captchaId = ref('')
const captchaImg = ref('')
const form = reactive({ username: '', password: '', confirm: '', captchaCode: '' })

async function loadCaptcha() {
  try {
    const c = await api.captcha()
    captchaId.value = c.captchaId
    captchaImg.value = c.image.startsWith('data:') ? c.image : `data:image/png;base64,${c.image}`
    form.captchaCode = ''
  } catch (e) {
    showFailToast(e.message || '验证码加载失败')
  }
}

watch(mode, () => {
  form.confirm = ''
})

onMounted(async () => {
  await auth.bootstrap()
  if (auth.isLoggedIn) {
    router.replace('/')
    return
  }
  loadCaptcha()
})

async function submit() {
  if (mode.value === 'register' && form.password !== form.confirm) {
    showFailToast('两次密码不一致')
    return
  }
  loading.value = true
  try {
    const payload = {
      username: form.username.trim(),
      password: form.password,
      captchaId: captchaId.value,
      captchaCode: form.captchaCode.trim(),
    }
    if (mode.value === 'login') await auth.login(payload)
    else await auth.register(payload)
    showSuccessToast('欢迎回来')
    router.replace('/')
  } catch (e) {
    showFailToast(e instanceof ApiError ? e.message : '操作失败')
    loadCaptcha()
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-hero {
  padding: calc(var(--sl-safe-top) + 48px) 24px 24px;
  text-align: center;
  background: linear-gradient(160deg, #1a56db 0%, #3b82f6 100%);
  color: #fff;
}

.brand-icon {
  font-size: 48px;
  margin-bottom: 8px;
}

.login-hero h1 {
  margin: 0;
  font-size: 28px;
}

.login-hero p {
  margin: 8px 0 0;
  opacity: 0.9;
  font-size: 14px;
}

.login-form {
  margin-top: -16px;
}

.login-form :deep(.van-tabs) {
  background: transparent;
}

.login-form :deep(.van-tabs__nav) {
  background: var(--sl-card);
  margin: 0 16px;
  border-radius: 12px 12px 0 0;
}

.login-actions {
  padding: 24px 16px;
}

.captcha-img {
  height: 36px;
  border-radius: 6px;
  cursor: pointer;
}
</style>
