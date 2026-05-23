<template>
  <div class="page settings-page">
    <h2>{{ t('settings.title') }}</h2>

    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="success" class="alert alert-success">{{ success }}</div>

    <section class="settings-section panel">
      <h3>{{ t('settings.theme.title') }}</h3>

      <div class="form-row">
        <label>{{ t('settings.theme.language') }}</label>
        <AppSelect
          v-model="currentLocale"
          :options="localeOptions"
          @change="onLocaleChange"
        />
      </div>

      <div class="form-row">
        <label>{{ t('settings.theme.mode') }}</label>
        <div class="segmented">
          <button
            type="button"
            class="seg-btn"
            :class="{ active: themeMode === 'light' }"
            @click="setMode('light')"
          >
            {{ t('settings.theme.light') }}
          </button>
          <button
            type="button"
            class="seg-btn"
            :class="{ active: themeMode === 'dark' }"
            @click="setMode('dark')"
          >
            {{ t('settings.theme.dark') }}
          </button>
        </div>
      </div>

      <div class="form-row">
        <label>{{ t('settings.theme.accent') }}</label>
        <div class="accent-grid">
          <button
            v-for="p in ACCENT_PRESETS"
            :key="p.id"
            type="button"
            class="accent-chip"
            :class="{ active: accentId === p.id }"
            :title="accentName(p)"
            @click="setAccentColor(p.id)"
          >
            <span class="swatch" :style="{ background: p.swatch }" />
            <span>{{ accentName(p) }}</span>
          </button>
        </div>
      </div>
    </section>

    <section id="personal" class="settings-section panel">
      <h3>{{ t('settings.personal.title') }}</h3>

      <div v-if="loading" class="muted">{{ t('settings.personal.loading') }}</div>

      <template v-else-if="profile">
        <div class="avatar-block">
          <div class="avatar-wrap">
            <img :src="avatarSrc" :alt="profile.nickname" @error="onAvatarError" />
          </div>
          <FileUploadZone
            compact
            accept="image/png,image/jpeg,image/webp,image/gif"
            :title="t('settings.personal.avatarTitle')"
            :disabled="uploading"
            @file="onPickAvatar"
          />
        </div>

        <dl class="info-list">
          <div><dt>{{ t('settings.personal.userId') }}</dt><dd class="mono">{{ profile.id }}</dd></div>
          <div><dt>{{ t('settings.personal.username') }}</dt><dd>{{ profile.username }}</dd></div>
          <div><dt>{{ t('settings.personal.createdAt') }}</dt><dd>{{ formatTime(profile.createdAt) }}</dd></div>
        </dl>

        <form class="nickname-form" @submit.prevent="saveNickname">
          <div class="form-row">
            <label>{{ t('settings.personal.nickname') }}</label>
            <input v-model="nickname" maxlength="32" :placeholder="t('settings.personal.nicknamePh')" />
          </div>
          <button class="btn-primary" type="submit" :disabled="saving">
            {{ saving ? t('settings.personal.saving') : t('settings.personal.saveNickname') }}
          </button>
        </form>

        <section class="danger-zone">
          <h4>{{ t('settings.personal.deleteTitle') }}</h4>
          <div class="form-row">
            <label>{{ t('settings.personal.username') }}</label>
            <input v-model="deleteForm.username" autocomplete="username" />
          </div>
          <div class="form-row">
            <label>{{ t('settings.personal.password') }}</label>
            <input v-model="deleteForm.password" type="password" autocomplete="current-password" />
          </div>
          <button class="btn-danger" type="button" :disabled="deleting" @click="onDeleteAccount">
            {{ deleting ? t('settings.personal.deleting') : t('settings.personal.deleteBtn') }}
          </button>
        </section>
      </template>
    </section>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppSelect from '../components/AppSelect.vue'
import FileUploadZone from '../components/FileUploadZone.vue'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'
import { useI18n } from '../composables/useI18n'
import { ACCENT_PRESETS, getAccent, getTheme, setAccent, setTheme } from '../utils/theme'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const { locale, t, setLocale, localeOptions } = useI18n()

const currentLocale = ref(locale.value)
const themeMode = ref(getTheme())
const accentId = ref(getAccent())
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

const ACCENT_I18N = {
  blue: { zh: '天蓝', en: 'Blue' },
  teal: { zh: '青绿', en: 'Teal' },
  violet: { zh: '紫罗兰', en: 'Violet' },
  amber: { zh: '琥珀', en: 'Amber' },
  rose: { zh: '玫红', en: 'Rose' },
}

function accentName(p) {
  return ACCENT_I18N[p.id]?.[locale.value] ?? p.name
}

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

function formatTime(time) {
  if (!time) return '-'
  return new Date(time).toLocaleString(locale.value === 'en' ? 'en-US' : 'zh-CN')
}

function setMode(mode) {
  themeMode.value = setTheme(mode)
}

function setAccentColor(id) {
  accentId.value = setAccent(id)
}

function onLocaleChange(val) {
  setLocale(val)
  currentLocale.value = val
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
    error.value = e instanceof ApiError ? e.message : t('settings.loadFail')
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
    success.value = t('settings.saveOk')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t('settings.saveFail')
  } finally {
    saving.value = false
  }
}

async function onPickAvatar(file) {
  if (!file) return
  if (file.size > 2 * 1024 * 1024) {
    error.value = t('settings.avatarSize')
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
    success.value = t('settings.avatarOk')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t('settings.avatarFail')
  } finally {
    uploading.value = false
  }
}

async function onDeleteAccount() {
  if (!confirm(t('settings.deleteConfirm'))) return
  deleting.value = true
  error.value = ''
  success.value = ''
  try {
    await api.deleteAccount(deleteForm.value.username.trim(), deleteForm.value.password)
    auth.accessToken = null
    auth.user = null
    router.replace('/login')
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t('settings.deleteFail')
  } finally {
    deleting.value = false
  }
}

function scrollToHash() {
  if (route.hash !== '#personal') return
  nextTick(() => {
    document.getElementById('personal')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  })
}

watch(locale, (val) => {
  currentLocale.value = val
})

onMounted(() => {
  load()
  scrollToHash()
})

watch(() => route.hash, scrollToHash)
</script>

<style scoped>
.settings-page {
  max-width: 36rem;
}

.settings-section {
  margin-top: 1rem;
}

.settings-section h3 {
  margin: 0 0 0.35rem;
  font-size: 1.05rem;
}

.segmented {
  display: inline-flex;
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow: hidden;
  background: var(--bg);
}

.seg-btn {
  padding: 0.45rem 1rem;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font: inherit;
  font-weight: 600;
  cursor: pointer;
}

.seg-btn:hover {
  background: var(--hover);
  color: var(--text);
}

.seg-btn.active {
  background: var(--accent-soft);
  color: var(--accent);
}

.accent-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.accent-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-muted);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
}

.accent-chip.active {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-soft);
}

.swatch {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.avatar-block {
  display: flex;
  gap: 1.25rem;
  align-items: flex-start;
  margin: 1rem 0 1.25rem;
  flex-wrap: wrap;
}

.avatar-wrap {
  width: 88px;
  height: 88px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid var(--border);
  background: var(--bg);
  flex-shrink: 0;
}

.avatar-wrap img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.info-list {
  margin: 0 0 1.25rem;
}

.info-list > div {
  display: flex;
  gap: 1rem;
  padding: 0.4rem 0;
  border-bottom: 1px solid var(--border);
  font-size: 0.9rem;
}

.info-list dt {
  width: 5rem;
  color: var(--text-muted);
  margin: 0;
}

.info-list dd {
  margin: 0;
  flex: 1;
}

.mono {
  font-family: ui-monospace, monospace;
  color: var(--accent);
  word-break: break-all;
}

.nickname-form {
  border-top: 1px solid var(--border);
  padding-top: 1rem;
}

.danger-zone {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid color-mix(in srgb, var(--danger) 40%, transparent);
}

.danger-zone h4 {
  color: var(--danger);
  margin: 0 0 0.5rem;
  font-size: 0.95rem;
}

.btn-danger {
  margin-top: 0.5rem;
  padding: 0.55rem 1rem;
  background: transparent;
  border: 1px solid var(--danger);
  color: var(--danger);
  font-weight: 600;
  border-radius: 10px;
  cursor: pointer;
}

.btn-danger:hover:not(:disabled) {
  background: color-mix(in srgb, var(--danger) 12%, transparent);
}
</style>
