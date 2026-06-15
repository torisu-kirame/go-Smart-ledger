<template>
  <div class="settings-module">
    <div class="profile-bar panel">
      <div class="form-row profile-bar__select">
        <label>{{ t('settings.ai.profileSelect') }}</label>
        <AppSelect
          v-model="activeProfileId"
          :options="profileSelectOptions"
          @change="onProfileSelect"
        />
      </div>
      <p v-if="activeProfile?.description" class="profile-desc">
        {{ activeProfile.description }}
      </p>
      <div class="actions-row profile-bar__actions">
        <button type="button" class="btn-primary btn-sm" @click="openSaveModal">
          {{ t('settings.ai.saveProfile') }}
        </button>
        <button type="button" class="btn-ghost btn-sm" @click="openRenameModal">
          {{ t('settings.ai.renameProfile') }}
        </button>
        <button
          type="button"
          class="btn-ghost btn-sm btn-danger-text"
          :disabled="profilesState.profiles.length <= 1"
          @click="confirmDeleteProfile"
        >
          {{ t('settings.ai.deleteProfile') }}
        </button>
      </div>
    </div>

    <label class="inline-check">
      <input v-model="ai.enabled" type="checkbox" @change="onEnabledChange" />
      {{ t('settings.ai.enabled') }}
    </label>

    <div v-if="showApiKeyWarn" class="warn-panel panel">
      <p class="hint-title">{{ t('settings.ai.apiKeyWarnTitle') }}</p>
      <p class="hint-text">{{ t('settings.ai.apiKeyWarnText') }}</p>
    </div>

    <div v-if="showOfflineHint" class="offline-panel panel">
      <p class="hint-title">{{ t('settings.ai.offlineTitle') }}</p>
      <p class="hint-text">{{ t('settings.ai.offlineHint') }}</p>
      <ul class="offline-reqs">
        <li>{{ t('settings.ai.offlineReq1') }}</li>
        <li>{{ t('settings.ai.offlineReq2') }}</li>
        <li>{{ t('settings.ai.offlineReq3') }}</li>
        <li>{{ t('settings.ai.offlineReq4') }}</li>
      </ul>
      <button type="button" class="btn-ghost btn-sm" @click="applyOfflineDockerPreset">
        {{ t('settings.ai.offlineApplyPreset') }}
      </button>
    </div>

    <div class="form-row">
      <label>{{ t('settings.ai.provider') }}</label>
      <AppSelect
        v-model="ai.provider"
        :options="providerOptions"
        @change="onProviderChange"
      />
    </div>

    <div class="form-row">
      <label>{{ t('settings.ai.baseUrl') }}</label>
      <input v-model="ai.baseUrl" type="url" @change="persistAi" />
    </div>

    <div class="form-row">
      <label>{{ t('settings.ai.chatModel') }}</label>
      <div class="model-row">
        <AppSelect
          v-model="modelPick"
          :options="modelOptions"
          @change="onModelPick"
        />
        <input
          v-model="ai.chatModel"
          class="model-custom"
          :placeholder="t('settings.ai.modelCustomPh')"
          @change="persistAi"
        />
      </div>
    </div>

    <div class="form-row">
      <label>{{ t('settings.ai.apiKey') }}</label>
      <input v-model="ai.apiKey" type="password" autocomplete="off" @change="persistAi" />
      <p class="field-hint">{{ t('settings.ai.apiKeyHint') }}</p>
    </div>

    <div class="test-row panel">
      <button
        type="button"
        class="btn-primary btn-sm"
        :disabled="testingConnection || !canTestConnection"
        @click="runTestConnection"
      >
        {{ testingConnection ? t('settings.ai.testing') : t('settings.ai.testConnection') }}
      </button>
      <p v-if="ai.connectionVerified" class="ok-line test-status">
        {{ t('settings.ai.testOk') }}
      </p>
      <p v-else-if="ai.enabled" class="hint-text test-status">
        {{ t('settings.ai.testRequired') }}
      </p>
      <p v-if="testError" class="err-line">{{ testError }}</p>
    </div>

    <p class="hint-text agent-hint">{{ t('settings.ai.agentBackendHint') }}</p>

    <div v-if="showProfileModal" class="modal">
      <form class="modal-card profile-modal" @submit.prevent="submitProfileModal">
        <h3>{{ profileModalTitle }}</h3>
        <div class="form-row">
          <label>{{ t('settings.ai.profileName') }}</label>
          <input v-model="profileForm.name" type="text" required maxlength="64" />
        </div>
        <div class="form-row">
          <label>{{ t('settings.ai.profileDesc') }}</label>
          <textarea
            v-model="profileForm.description"
            rows="3"
            maxlength="200"
            :placeholder="t('settings.ai.profileDescPh')"
          />
        </div>
        <div class="modal-actions">
          <button type="button" class="btn-ghost" @click="closeProfileModal">
            {{ t('settings.ai.modalCancel') }}
          </button>
          <button type="submit" class="btn-primary">
            {{ t('settings.ai.modalSave') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import AppSelect from '../AppSelect.vue'
import { useI18n } from '../../composables/useI18n'
import { testAiConnection } from '../../api/aiChat'
import {
  AI_PROVIDERS,
  applyProfile,
  applyProviderDefaults,
  deleteProfile,
  getActiveProfile,
  isOfflineProvider,
  loadAiConfig,
  loadProfilesState,
  markConnectionUnverified,
  needsApiKey,
  offlineDockerConfig,
  providerModelOptions,
  renameProfile,
  saveAiConfig,
  saveProfileAs,
} from '../../utils/aiConfig'

const { t } = useI18n()
const profilesState = ref(loadProfilesState())
const activeProfileId = ref(profilesState.value.activeProfileId)
const ai = ref(loadAiConfig())
const modelPick = ref(ai.value.chatModel)
const testingConnection = ref(false)
const testError = ref('')

const canTestConnection = computed(() => {
  if (needsApiKey(ai.value) && !ai.value.apiKey?.trim()) return false
  if (!ai.value.baseUrl?.trim()) return false
  if (!ai.value.chatModel?.trim()) return false
  return true
})

const showProfileModal = ref(false)
const profileModalMode = ref('save')
const profileForm = ref({ name: '', description: '' })

const activeProfile = computed(() => getActiveProfile(profilesState.value))

const profileSelectOptions = computed(() =>
  profilesState.value.profiles.map((p) => ({
    value: p.id,
    label: p.name,
  }))
)

const profileModalTitle = computed(() =>
  profileModalMode.value === 'rename'
    ? t('settings.ai.renameProfile')
    : t('settings.ai.saveProfile')
)

const providerOptions = computed(() =>
  Object.entries(AI_PROVIDERS).map(([value, p]) => ({ value, label: p.label }))
)

const modelOptions = computed(() => providerModelOptions(ai.value.provider))

const showOfflineHint = computed(() => isOfflineProvider(ai.value.provider))

const showApiKeyWarn = computed(
  () => ai.value.enabled && needsApiKey(ai.value) && !ai.value.apiKey?.trim()
)

watch(
  () => ai.value.chatModel,
  (m) => {
    if (modelOptions.value.some((o) => o.value === m)) modelPick.value = m
  }
)

function refreshProfiles() {
  profilesState.value = loadProfilesState()
  activeProfileId.value = profilesState.value.activeProfileId
}

function loadActiveConfig() {
  ai.value = loadAiConfig()
  modelPick.value = ai.value.chatModel
}

function onEnabledChange() {
  if (ai.value.enabled && needsApiKey(ai.value) && !ai.value.apiKey?.trim()) {
    // 仍允许勾选，由警告条提示填写 API Key
  }
  persistAi()
}

function applyOfflineDockerPreset() {
  ai.value = { ...ai.value, ...offlineDockerConfig() }
  modelPick.value = ai.value.chatModel
  persistAi()
}

function persistAi(clearVerified = true) {
  if (clearVerified) {
    ai.value = markConnectionUnverified(ai.value)
    testError.value = ''
  }
  saveAiConfig(ai.value)
  refreshProfiles()
}

async function runTestConnection() {
  testError.value = ''
  testingConnection.value = true
  persistAi(false)
  try {
    await testAiConnection(ai.value)
    refreshProfiles()
    loadActiveConfig()
  } catch (e) {
    testError.value = e?.message || t('settings.ai.testFail')
  } finally {
    testingConnection.value = false
  }
}

function onProfileSelect() {
  applyProfile(activeProfileId.value)
  refreshProfiles()
  loadActiveConfig()
}

function onProviderChange() {
  ai.value = applyProviderDefaults(ai.value, ai.value.provider)
  modelPick.value = ai.value.chatModel
  persistAi()
}

function onModelPick() {
  if (modelPick.value) {
    ai.value.chatModel = modelPick.value
    persistAi()
  }
}

function openSaveModal() {
  profileModalMode.value = 'save'
  profileForm.value = {
    name: '',
    description: activeProfile.value?.description || '',
  }
  showProfileModal.value = true
}

function openRenameModal() {
  profileModalMode.value = 'rename'
  profileForm.value = {
    name: activeProfile.value?.name || '',
    description: activeProfile.value?.description || '',
  }
  showProfileModal.value = true
}

function closeProfileModal() {
  showProfileModal.value = false
}

function submitProfileModal() {
  const name = profileForm.value.name.trim()
  if (!name) return

  if (profileModalMode.value === 'rename') {
    renameProfile(activeProfileId.value, name, profileForm.value.description)
    refreshProfiles()
  } else {
    persistAi()
    saveProfileAs(name, profileForm.value.description, ai.value)
    refreshProfiles()
    activeProfileId.value = profilesState.value.activeProfileId
    loadActiveConfig()
  }
  closeProfileModal()
}

function confirmDeleteProfile() {
  if (profilesState.value.profiles.length <= 1) return
  const name = activeProfile.value?.name || ''
  if (!window.confirm(t('settings.ai.deleteProfileConfirm').replace('{name}', name))) return
  deleteProfile(activeProfileId.value)
  refreshProfiles()
  activeProfileId.value = profilesState.value.activeProfileId
  loadActiveConfig()
}
</script>

<style scoped>
.offline-panel,
.warn-panel {
  margin: 1rem 0;
  padding: 0.85rem 1rem;
  border-radius: var(--radius-sm);
}

.test-row {
  margin: 1rem 0;
  padding: 0.85rem 1rem;
  border-radius: var(--radius-sm);
  background: color-mix(in srgb, var(--accent) 4%, var(--bg-card));
}

.test-status {
  margin: 0.5rem 0 0;
}

.warn-panel {
  background: color-mix(in srgb, var(--danger) 8%, var(--bg-card));
  border: 1px solid color-mix(in srgb, var(--danger) 25%, var(--border));
}

.offline-panel {
  background: color-mix(in srgb, var(--accent) 5%, var(--bg-card));
  border: 1px dashed var(--border);
}

.offline-reqs {
  margin: 0 0 0.75rem;
  padding-left: 1.25rem;
  font-size: 0.82rem;
  color: var(--text-muted);
  line-height: 1.55;
}

.offline-reqs li {
  margin-bottom: 0.25rem;
}

.profile-bar {
  margin-bottom: 1.25rem;
  padding: 0.85rem 1rem;
  background: color-mix(in srgb, var(--accent) 5%, var(--bg-card));
}

.profile-bar__select {
  margin-bottom: 0.5rem;
}

.profile-desc {
  margin: 0 0 0.65rem;
  font-size: 0.82rem;
  color: var(--text-muted);
  line-height: 1.5;
}

.profile-bar__actions {
  margin-top: 0.25rem;
}

.btn-danger-text:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.btn-danger-text:not(:disabled) {
  color: var(--danger);
}

.model-row {
  display: grid;
  gap: 0.5rem;
}

.model-custom {
  width: 100%;
}

.field-hint {
  margin: 0.25rem 0 0;
  font-size: 0.78rem;
  color: var(--text-muted);
}

.agent-hint {
  margin-top: 0.75rem;
  color: var(--text-muted, #64748b);
  font-size: 0.875rem;
}

.ok-line {
  color: var(--success);
  font-size: 0.875rem;
  margin: 0.5rem 0 0;
}

.err-line {
  color: var(--danger);
  font-size: 0.875rem;
  margin: 0.5rem 0 0;
}

.hint-title {
  margin: 0 0 0.35rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text);
}

.hint-text {
  margin: 0 0 0.65rem;
  font-size: 0.82rem;
  color: var(--text-muted);
  line-height: 1.5;
}

.btn-sm {
  font-size: 0.8125rem;
  padding: 0.35rem 0.65rem;
}

.modal {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
}

.modal-card.profile-modal {
  width: min(100%, 440px);
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.25rem 1.5rem;
}

.profile-modal h3 {
  margin: 0 0 1rem;
  font-size: 1.05rem;
}

.modal-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 1rem;
}
</style>
