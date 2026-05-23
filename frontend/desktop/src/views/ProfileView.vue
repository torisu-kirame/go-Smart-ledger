<template>
  <div class="profile-page">
    <h2>个人中心</h2>
    <p class="muted">管理昵称与头像</p>

    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="success" class="alert alert-success">{{ success }}</div>

    <div v-if="loading" class="muted">加载中…</div>

    <div v-else-if="profile" class="card profile-card">
      <div class="avatar-block">
        <div class="avatar-wrap">
          <img :src="avatarSrc" :alt="profile.nickname" @error="onAvatarError" />
        </div>
        <div class="avatar-actions">
          <label class="btn-primary file-btn">
            更换头像
            <input type="file" accept="image/png,image/jpeg,image/webp,image/gif" hidden @change="onPickAvatar" />
          </label>
          <p class="hint">支持 JPG/PNG/WebP/GIF，最大 2MB</p>
        </div>
      </div>

      <dl class="info-list">
        <div><dt>用户 ID</dt><dd class="mono">{{ profile.id }}</dd></div>
        <div><dt>登录名</dt><dd>{{ profile.username }}</dd></div>
        <div><dt>注册时间</dt><dd>{{ formatTime(profile.createdAt) }}</dd></div>
      </dl>

      <form class="nickname-form" @submit.prevent="saveNickname">
        <div class="form-row">
          <label>昵称</label>
          <input v-model="nickname" maxlength="32" placeholder="展示名称，最多 32 字" />
        </div>
        <button class="btn-primary" type="submit" :disabled="saving">{{ saving ? '保存中…' : '保存昵称' }}</button>
      </form>

      <section class="danger-zone">
        <h3>注销账号</h3>
        <p class="hint">注销后不可恢复，将删除个人资料、好友关系与头像。须输入当前登录名与密码确认。</p>
        <div class="form-row">
          <label>登录名</label>
          <input v-model="deleteForm.username" autocomplete="username" />
        </div>
        <div class="form-row">
          <label>密码</label>
          <input v-model="deleteForm.password" type="password" autocomplete="current-password" />
        </div>
        <button class="btn-danger" type="button" :disabled="deleting" @click="onDeleteAccount">
          {{ deleting ? '注销中…' : '永久注销账号' }}
        </button>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'

const router = useRouter()

const auth = useAuthStore()
const profile = ref(null)
const nickname = ref('')
const loading = ref(true)
const saving = ref(false)
const uploading = ref(false)
const deleting = ref(false)
const deleteForm = ref({ username: '', password: '' })
const error = ref('')
const success = ref('')
const avatarBust = ref(0)
const avatarFailed = ref(false)

const avatarSrc = computed(() => {
  if (!profile.value || avatarFailed.value) return defaultAvatar(profile.value?.nickname || '?')
  const url = profile.value.avatarUrl || `/api/v1/users/${profile.value.id}/avatar`
  return `${url}?t=${avatarBust.value}`
})

function defaultAvatar(text) {
  const ch = (text || '?').charAt(0).toUpperCase()
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="120" height="120"><rect fill="#1a2230" width="120" height="120"/><text x="50%" y="54%" dominant-baseline="middle" text-anchor="middle" fill="#3d8bfd" font-size="48" font-family="sans-serif">${ch}</text></svg>`
  return `data:image/svg+xml,${encodeURIComponent(svg)}`
}

function onAvatarError() {
  avatarFailed.value = true
}

function formatTime(t) {
  if (!t) return '-'
  return new Date(t).toLocaleString()
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    profile.value = await api.getProfile()
    nickname.value = profile.value.nickname || profile.value.username
    deleteForm.value.username = profile.value.username
    avatarFailed.value = false
    avatarBust.value = Date.now()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function saveNickname() {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    profile.value = await api.updateProfile(nickname.value.trim())
    auth.user = { ...auth.user, nickname: profile.value.nickname, avatarUrl: profile.value.avatarUrl }
    success.value = '昵称已更新'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '保存失败'
  } finally {
    saving.value = false
  }
}

async function onPickAvatar(ev) {
  const file = ev.target.files?.[0]
  ev.target.value = ''
  if (!file) return
  if (file.size > 2 * 1024 * 1024) {
    error.value = '图片不能超过 2MB'
    return
  }
  uploading.value = true
  error.value = ''
  success.value = ''
  try {
    profile.value = await api.uploadAvatar(file)
    auth.user = { ...auth.user, avatarUrl: profile.value.avatarUrl, nickname: profile.value.nickname }
    avatarFailed.value = false
    avatarBust.value = Date.now()
    success.value = '头像已更新'
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '上传失败'
  } finally {
    uploading.value = false
  }
}

async function onDeleteAccount() {
  if (!confirm('确定永久注销？此操作不可撤销。')) return
  deleting.value = true
  error.value = ''
  success.value = ''
  try {
    await api.deleteAccount(deleteForm.value.username.trim(), deleteForm.value.password)
    auth.accessToken = null
    auth.user = null
    router.replace('/login')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '注销失败'
  } finally {
    deleting.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.profile-page { max-width: 520px; }
.muted { color: var(--text-muted); font-size: 0.875rem; }
.card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius); padding: 1.5rem; margin-top: 1rem; }
.avatar-block { display: flex; gap: 1.25rem; align-items: center; margin-bottom: 1.5rem; flex-wrap: wrap; }
.avatar-wrap {
  width: 96px; height: 96px; border-radius: 50%; overflow: hidden;
  border: 2px solid var(--border); background: var(--bg);
}
.avatar-wrap img { width: 100%; height: 100%; object-fit: cover; }
.file-btn { display: inline-block; cursor: pointer; }
.hint { font-size: 0.75rem; color: var(--text-muted); margin: 0.5rem 0 0; }
.info-list { margin: 0 0 1.5rem; }
.info-list > div { display: flex; gap: 1rem; padding: 0.4rem 0; border-bottom: 1px solid var(--border); font-size: 0.9rem; }
.info-list dt { width: 5rem; color: var(--text-muted); margin: 0; }
.info-list dd { margin: 0; flex: 1; }
.mono { font-family: ui-monospace, monospace; color: var(--accent); }
.nickname-form { border-top: 1px solid var(--border); padding-top: 1rem; }
.danger-zone { margin-top: 1.5rem; padding-top: 1rem; border-top: 1px solid rgba(248,113,113,.35); }
.danger-zone h3 { color: var(--danger); margin: 0 0 0.5rem; font-size: 1rem; }
.btn-danger {
  width: 100%; margin-top: 0.5rem; padding: 0.55rem;
  background: transparent; border: 1px solid var(--danger); color: var(--danger); font-weight: 600;
}
.btn-danger:hover:not(:disabled) { background: rgba(248,113,113,.12); }
</style>
