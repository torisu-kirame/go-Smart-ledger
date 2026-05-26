<template>
  <div class="settings-module">
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
          :hint="t('settings.personal.avatarHint')"
          :disabled="uploading"
          @file="$emit('upload-avatar', $event)"
        />
      </div>

      <dl class="info-list">
        <div><dt>{{ t('settings.personal.userId') }}</dt><dd class="mono">{{ profile.id }}</dd></div>
        <div><dt>{{ t('settings.personal.username') }}</dt><dd>{{ profile.username }}</dd></div>
        <div><dt>{{ t('settings.personal.createdAt') }}</dt><dd>{{ formatTime(profile.createdAt) }}</dd></div>
      </dl>

      <form class="nickname-form" @submit.prevent="$emit('save-nickname')">
        <div class="form-row">
          <label>{{ t('settings.personal.nickname') }}</label>
          <input
            :value="nickname"
            maxlength="32"
            :placeholder="t('settings.personal.nicknamePh')"
            @input="$emit('update:nickname', $event.target.value)"
          />
        </div>
        <button class="btn-primary" type="submit" :disabled="saving">
          {{ saving ? t('settings.personal.saving') : t('settings.personal.saveNickname') }}
        </button>
      </form>
    </template>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import FileUploadZone from '../FileUploadZone.vue'
import { useI18n } from '../../composables/useI18n'

const props = defineProps({
  profile: { type: Object, default: null },
  nickname: { type: String, default: '' },
  loading: { type: Boolean, default: false },
  saving: { type: Boolean, default: false },
  uploading: { type: Boolean, default: false },
  avatarBust: { type: Number, default: 0 },
})

defineEmits(['update:nickname', 'save-nickname', 'upload-avatar'])

const { locale, t } = useI18n()
const avatarFailed = ref(false)

const avatarSrc = computed(() => {
  if (!props.profile || avatarFailed.value) {
    return defaultAvatar(props.profile?.nickname || '?')
  }
  const url = props.profile.avatarUrl || `/api/v1/users/${props.profile.id}/avatar`
  return `${url}?t=${props.avatarBust}`
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
</script>

<style scoped>
.avatar-block {
  display: flex;
  gap: 1.25rem;
  align-items: flex-start;
  margin: 0 0 1.25rem;
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
</style>
