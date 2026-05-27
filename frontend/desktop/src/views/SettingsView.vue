<template>
  <div class="page settings-page">
    <PageHeader :crumbs="headerCrumbs" :subtitle="t('settings.subtitle')" />

    <div class="settings-layout">
      <SettingsModuleNav :active="activeSection" @select="selectSection" />

      <section class="settings-panel panel" :aria-labelledby="panelTitleId">
        <header class="module-header">
          <h3 :id="panelTitleId">{{ moduleTitle }}</h3>
          <p class="module-desc">{{ moduleDesc }}</p>
        </header>

        <SettingsAppearanceModule v-if="activeSection === 'appearance'" />
        <SettingsLanguageModule v-else-if="activeSection === 'language'" />
        <SettingsAiModule v-else-if="activeSection === 'ai'" />
        <SettingsAccountModule
          v-else-if="activeSection === 'account'"
          :profile="profile"
          :nickname="nickname"
          :loading="loading"
          :saving="saving"
          :uploading="uploading"
          :avatar-bust="avatarBust"
          @update:nickname="nickname = $event"
          @save-nickname="saveNickname"
          @upload-avatar="onPickAvatar"
        />
        <SettingsSecurityModule
          v-else-if="activeSection === 'security'"
          :delete-form="deleteForm"
          :deleting="deleting"
          @update:delete-form="deleteForm = $event"
          @delete-account="onDeleteAccount"
        />
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SettingsAccountModule from '../components/settings/SettingsAccountModule.vue'
import SettingsAiModule from '../components/settings/SettingsAiModule.vue'
import SettingsAppearanceModule from '../components/settings/SettingsAppearanceModule.vue'
import SettingsLanguageModule from '../components/settings/SettingsLanguageModule.vue'
import PageHeader from '../components/PageHeader.vue'
import SettingsModuleNav from '../components/settings/SettingsModuleNav.vue'
import SettingsSecurityModule from '../components/settings/SettingsSecurityModule.vue'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'
import { useI18n } from '../composables/useI18n'
import { normalizeSettingsHash } from '../utils/settingsSections'
import { useNotifyStore } from '../stores/notify'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const { t } = useI18n()
const notify = useNotifyStore()

const activeSection = ref(normalizeSettingsHash(route.hash))
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

const panelTitleId = computed(() => `settings-panel-${activeSection.value}`)
const moduleTitle = computed(() => t(`settings.modules.${activeSection.value}.title`))
const moduleDesc = computed(() => t(`settings.modules.${activeSection.value}.desc`))

const headerCrumbs = computed(() => [
  { label: t('settings.title'), to: { path: '/settings', hash: '#appearance' } },
  { label: t(`settings.nav.${activeSection.value}`) },
])

function selectSection(id) {
  activeSection.value = id
  router.replace({ path: '/settings', hash: `#${id}` })
}

function syncHashFromRoute() {
  activeSection.value = normalizeSettingsHash(route.hash)
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    profile.value = await api.getProfile()
    nickname.value = profile.value.nickname || profile.value.username
    deleteForm.value.username = profile.value.username
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

onMounted(() => {
  load()
  syncHashFromRoute()
})

watch(() => route.hash, syncHashFromRoute)

watch(success, (v) => {
  if (!v) return
  notify.success(v)
  success.value = ''
})

watch(error, (v) => {
  if (!v) return
  notify.error(v)
  error.value = ''
})
</script>

<style scoped>
.settings-page {
  max-width: 52rem;
}

.settings-layout {
  display: flex;
  gap: 1.25rem;
  align-items: flex-start;
}

.settings-panel {
  flex: 1;
  min-width: 0;
}

.module-header {
  margin-bottom: 1.25rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border);
}

.module-header h3 {
  margin: 0;
  font-size: 1.1rem;
}

.module-desc {
  margin: 0.35rem 0 0;
  font-size: 0.85rem;
  color: var(--text-muted);
}

@media (max-width: 640px) {
  .settings-layout {
    flex-direction: column;
  }
}
</style>
