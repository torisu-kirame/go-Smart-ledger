<template>
  <div class="page">
    <van-nav-bar title="我的" fixed placeholder safe-area-inset-top />

    <div class="profile-header">
      <van-image round width="64" height="64" :src="avatarUrl" fit="cover">
        <template #error>
          <div class="avatar-fallback">{{ initials }}</div>
        </template>
      </van-image>
      <div class="profile-meta">
        <h2>{{ auth.user?.nickname || auth.user?.username || '用户' }}</h2>
        <p class="mono">{{ auth.user?.id }}</p>
      </div>
    </div>

    <van-cell-group inset title="账号">
      <van-field v-model="nickname" label="昵称" placeholder="设置昵称" />
      <van-cell title="保存昵称" is-link @click="saveNickname" />
      <van-cell title="退出登录" is-link class="danger-cell" @click="logout" />
    </van-cell-group>

    <van-cell-group inset title="服务器（APK / 跨域）">
      <van-field
        v-model="apiBaseInput"
        label="API 基址"
        type="textarea"
        rows="2"
        autosize
        :placeholder="serverHint"
      />
      <van-cell title="保存并应用" is-link @click="saveServer" />
      <van-cell title="恢复默认同源" is-link @click="resetServer" />
    </van-cell-group>

    <van-cell-group inset title="关于">
      <van-cell title="版本" value="0.1.0 mobile" />
      <van-cell title="运行环境" :value="platformLabel" />
      <van-cell title="当前 API" :label="currentApi" label-class="mono" />
    </van-cell-group>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showConfirmDialog, showFailToast, showSuccessToast } from 'vant'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'
import {
  defaultServerHint,
  getApiBase,
  isNativeApp,
  saveApiBase,
} from '../utils/apiBase'

defineOptions({ name: 'ProfileView' })

const router = useRouter()
const auth = useAuthStore()
const nickname = ref('')
const apiBaseInput = ref('')

const avatarUrl = computed(() => {
  if (!auth.user?.id) return ''
  return api.userAvatarUrl(auth.user.id)
})

const initials = computed(() => {
  const n = auth.user?.nickname || auth.user?.username || '?'
  return n.slice(0, 1).toUpperCase()
})

const platformLabel = computed(() => (isNativeApp() ? '原生 App' : '移动 Web'))
const currentApi = computed(() => getApiBase())
const serverHint = computed(() => defaultServerHint())

onMounted(() => {
  nickname.value = auth.user?.nickname || ''
  apiBaseInput.value = getApiBase() === '/api/v1' ? '' : getApiBase()
})

async function saveNickname() {
  try {
    await api.updateProfile(nickname.value.trim())
    await auth.refresh()
    showSuccessToast('已保存')
  } catch (e) {
    showFailToast(e instanceof ApiError ? e.message : '保存失败')
  }
}

async function saveServer() {
  const v = apiBaseInput.value.trim()
  if (v && !/^https?:\/\/.+\/api\/v1\/?$/i.test(v)) {
    showFailToast('格式示例：http://192.168.1.10:25175/api/v1')
    return
  }
  await saveApiBase(v)
  showSuccessToast('已保存，请重新登录以刷新会话')
  try {
    await auth.refresh()
  } catch {
    /* ignore */
  }
}

async function resetServer() {
  apiBaseInput.value = ''
  await saveApiBase('')
  showSuccessToast('已恢复默认同源')
}

async function logout() {
  try {
    await showConfirmDialog({ title: '退出登录', message: '确定退出？' })
    await auth.logout()
    router.replace('/login')
  } catch (e) {
    if (e !== 'cancel') showFailToast('退出失败')
  }
}
</script>

<style scoped>
.profile-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 16px;
  background: var(--sl-card);
  margin-bottom: 12px;
}

.profile-meta h2 {
  margin: 0;
  font-size: 20px;
}

.profile-meta p {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--sl-muted);
  word-break: break-all;
}

.avatar-fallback {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--sl-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 700;
}

.danger-cell :deep(.van-cell__title) {
  color: var(--sl-danger);
}
</style>
