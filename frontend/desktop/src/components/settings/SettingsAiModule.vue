<template>
  <div class="settings-module">
    <label class="inline-check">
      <input v-model="ai.enabled" type="checkbox" @change="persistAi" />
      {{ t('settings.ai.enabled') }}
    </label>
    <div class="form-row">
      <label>{{ t('settings.ai.provider') }}</label>
      <AppSelect
        v-model="ai.provider"
        :options="aiProviderOptions"
        @change="persistAi"
      />
    </div>
    <div class="form-row">
      <label>{{ t('settings.ai.baseUrl') }}</label>
      <input v-model="ai.baseUrl" type="url" @change="persistAi" />
    </div>
    <div class="form-row">
      <label>{{ t('settings.ai.chatModel') }}</label>
      <input v-model="ai.chatModel" @change="persistAi" />
    </div>
    <div class="form-row">
      <label>{{ t('settings.ai.embedModel') }}</label>
      <input v-model="ai.embedModel" @change="persistAi" />
    </div>
    <div class="form-row">
      <label>{{ t('settings.ai.apiKey') }}</label>
      <input v-model="ai.apiKey" type="password" autocomplete="off" @change="persistAi" />
    </div>
    <div class="form-row">
      <label>{{ t('settings.ai.gateway') }}</label>
      <input v-model="ai.openclawGateway" type="url" @change="persistAi" />
    </div>
    <div class="actions-row">
      <button type="button" class="btn-ghost" @click="copyOpenClawConfig">
        {{ t('settings.ai.copyConfig') }}
      </button>
    </div>
    <p v-if="aiCopied" class="ok-line">{{ t('settings.ai.copied') }}</p>
    <p v-if="copyError" class="err-line">{{ copyError }}</p>

    <div class="docker-hint panel">
      <p class="hint-title">{{ t('settings.ai.dockerTitle') }}</p>
      <p class="hint-text">{{ t('settings.ai.dockerHint') }}</p>
      <button type="button" class="btn-ghost btn-sm" @click="applyDockerDefaults">
        {{ t('settings.ai.dockerApply') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import AppSelect from '../AppSelect.vue'
import { useI18n } from '../../composables/useI18n'
import { exportOpenClawSnippet, loadAiConfig, saveAiConfig } from '../../utils/aiConfig'

const { t } = useI18n()
const ai = ref(loadAiConfig())
const aiCopied = ref(false)
const copyError = ref('')

const aiProviderOptions = computed(() => [
  { value: 'ollama', label: 'Ollama' },
  { value: 'openai', label: 'OpenAI 兼容' },
  { value: 'lmstudio', label: 'LM Studio' },
])

function persistAi() {
  saveAiConfig(ai.value)
  aiCopied.value = false
  copyError.value = ''
}

async function copyOpenClawConfig() {
  persistAi()
  const text = exportOpenClawSnippet(ai.value)
  try {
    await navigator.clipboard.writeText(text)
    aiCopied.value = true
    copyError.value = ''
  } catch {
    copyError.value = t('settings.ai.copyFail')
  }
}

function applyDockerDefaults() {
  ai.value = {
    ...ai.value,
    provider: 'ollama',
    baseUrl: 'http://127.0.0.1:11434/v1',
    chatModel: 'llama3.2',
    embedModel: 'nomic-embed-text',
    apiKey: 'ollama',
    openclawGateway: 'http://127.0.0.1:18789',
  }
  persistAi()
  aiCopied.value = false
  copyError.value = ''
}
</script>

<style scoped>
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

.docker-hint {
  margin-top: 1.25rem;
  padding: 0.85rem 1rem;
  background: color-mix(in srgb, var(--accent) 5%, var(--bg-card));
  border-style: dashed;
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
</style>
