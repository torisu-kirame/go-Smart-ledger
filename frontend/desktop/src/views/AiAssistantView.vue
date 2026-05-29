<template>
  <div class="page assistant-page">
    <PageHeader :crumbs="crumbs" :subtitle="t('assistant.subtitle')" />

    <div v-if="!aiEnabled" class="alert alert-warn">
      <span>{{ t('assistant.disabledHint') }}</span>
      <router-link to="/settings#ai" class="assistant-settings-link">
        {{ t('assistant.openSettings') }}
      </router-link>
    </div>

    <div class="panel assistant-toolbar">
      <div class="toolbar-row">
        <label class="toolbar-label">{{ t('assistant.ledgerContext') }}</label>
        <select v-model="selectedLedgerId" class="ledger-select" :disabled="loadingLedgers">
          <option value="">{{ t('assistant.noLedger') }}</option>
          <option v-for="lg in ledgers" :key="lg.id" :value="lg.id">
            {{ lg.name || lg.id }}
          </option>
        </select>
        <button
          type="button"
          class="btn-ghost btn-sm"
          :disabled="!selectedLedgerId || contextLoading"
          @click="loadLedgerContext"
        >
          <AppIcon name="refresh" size="sm" />
          <span>{{ contextLoading ? t('assistant.loadingContext') : t('assistant.syncContext') }}</span>
        </button>
        <span v-if="contextReady" class="context-badge">{{ t('assistant.contextReady') }}</span>
      </div>
      <p class="toolbar-hint">{{ t('assistant.contextHint') }}</p>
      <p v-if="contextError" class="err-line">{{ contextError }}</p>
    </div>

    <div ref="threadEl" class="panel assistant-thread" role="log" aria-live="polite">
      <div v-if="messages.length === 0" class="thread-empty">
        <AppIcon name="sparkles" size="md" class="thread-empty__icon" />
        <p>{{ t('assistant.emptyHint') }}</p>
      </div>
      <article
        v-for="(msg, i) in messages"
        :key="i"
        class="chat-bubble"
        :class="msg.role === 'user' ? 'chat-bubble--user' : 'chat-bubble--assistant'"
      >
        <header class="chat-bubble__role">
          {{ msg.role === 'user' ? t('assistant.roleUser') : t('assistant.roleAssistant') }}
        </header>
        <div class="chat-bubble__body">{{ msg.content }}</div>
      </article>
      <div
        v-if="streaming && !messages[messages.length - 1]?.content"
        class="chat-bubble chat-bubble--assistant chat-bubble--typing"
      >
        <span class="typing-dot" />
        <span class="typing-dot" />
        <span class="typing-dot" />
      </div>
    </div>

    <div v-if="chatError" class="alert alert-error">{{ chatError }}</div>

    <form class="assistant-composer panel" @submit.prevent="send">
      <textarea
        v-model="input"
        class="composer-input"
        rows="3"
        :placeholder="t('assistant.inputPlaceholder')"
        :disabled="!aiEnabled || streaming"
        @keydown.enter.exact.prevent="send"
      />
      <div class="composer-actions">
        <button
          v-if="streaming"
          type="button"
          class="btn-ghost"
          @click="stop"
        >
          {{ t('assistant.stop') }}
        </button>
        <button
          type="submit"
          class="btn-primary icon-btn"
          :disabled="!aiEnabled || streaming || !input.trim()"
        >
          <AppIcon name="send" size="sm" />
          <span>{{ t('assistant.send') }}</span>
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import AppIcon from '../components/AppIcon.vue'
import PageHeader from '../components/PageHeader.vue'
import { api } from '../api/http'
import {
  buildLedgerContextPrompt,
  defaultSystemMessages,
  streamChat,
} from '../api/aiChat'
import { useI18n } from '../composables/useI18n'
import { loadAiConfig } from '../utils/aiConfig'

const { t } = useI18n()

const crumbs = computed(() => [{ label: t('assistant.title') }])

const aiEnabled = computed(() => loadAiConfig().enabled)
const ledgers = ref([])
const loadingLedgers = ref(false)
const selectedLedgerId = ref('')
const ledgerContext = ref('')
const contextLoading = ref(false)
const contextError = ref('')
const contextReady = ref(false)

const messages = ref([])
const input = ref('')
const streaming = ref(false)
const chatError = ref('')
const threadEl = ref(null)
let abortCtrl = null

async function loadLedgers() {
  loadingLedgers.value = true
  try {
    const list = await api.listLedgers()
    ledgers.value = Array.isArray(list) ? list : list?.items || []
  } catch {
    ledgers.value = []
  } finally {
    loadingLedgers.value = false
  }
}

async function loadLedgerContext() {
  if (!selectedLedgerId.value) return
  contextLoading.value = true
  contextError.value = ''
  contextReady.value = false
  try {
    const data = await api.ragExport(selectedLedgerId.value)
    const name = ledgers.value.find((l) => l.id === selectedLedgerId.value)?.name
    ledgerContext.value = buildLedgerContextPrompt(data, name)
    contextReady.value = !!ledgerContext.value
  } catch (e) {
    contextError.value = e?.message || String(e)
    ledgerContext.value = ''
  } finally {
    contextLoading.value = false
  }
}

watch(selectedLedgerId, () => {
  contextReady.value = false
  ledgerContext.value = ''
  contextError.value = ''
})

async function scrollToBottom() {
  await nextTick()
  const el = threadEl.value
  if (el) el.scrollTop = el.scrollHeight
}

function stop() {
  abortCtrl?.abort()
}

async function send() {
  const text = input.value.trim()
  if (!text || streaming.value || !aiEnabled.value) return

  chatError.value = ''
  messages.value.push({ role: 'user', content: text })
  input.value = ''
  await scrollToBottom()

  const history = messages.value.map((m) => ({ role: m.role, content: m.content }))
  const payload = [...defaultSystemMessages(ledgerContext.value), ...history]

  streaming.value = true
  const assistantIdx = messages.value.length
  messages.value.push({ role: 'assistant', content: '' })
  abortCtrl = new AbortController()

  try {
    await streamChat({
      messages: payload,
      signal: abortCtrl.signal,
      onDelta(delta) {
        messages.value[assistantIdx].content += delta
        scrollToBottom()
      },
    })
  } catch (e) {
    if (e?.name === 'AbortError') {
      if (!messages.value[assistantIdx].content) {
        messages.value.splice(assistantIdx, 1)
      }
    } else if (e?.message === 'AI_DISABLED') {
      chatError.value = t('assistant.disabledHint')
      messages.value.splice(assistantIdx, 1)
    } else {
      chatError.value = e?.message || String(e)
      messages.value.splice(assistantIdx, 1)
    }
  } finally {
    streaming.value = false
    abortCtrl = null
    await scrollToBottom()
  }
}

onMounted(() => {
  loadLedgers()
})
</script>

<style scoped>
.assistant-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: var(--page-max);
  min-height: calc(100vh - 3rem);
}

.assistant-settings-link {
  margin-left: 0.5rem;
  color: var(--accent);
  text-decoration: underline;
}

.assistant-toolbar .toolbar-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
}

.toolbar-label {
  font-size: 0.875rem;
  color: var(--text-muted);
}

.ledger-select {
  min-width: 12rem;
  max-width: 20rem;
  padding: 0.4rem 0.6rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text);
}

.context-badge {
  font-size: 0.8rem;
  color: var(--success);
}

.toolbar-hint {
  margin: 0.5rem 0 0;
  font-size: 0.8rem;
  color: var(--text-muted);
}

.assistant-thread {
  flex: 1;
  min-height: 280px;
  max-height: min(56vh, 520px);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.thread-empty {
  margin: auto;
  text-align: center;
  color: var(--text-muted);
  padding: 2rem 1rem;
}

.thread-empty__icon {
  opacity: 0.6;
  margin-bottom: 0.5rem;
}

.chat-bubble {
  max-width: 92%;
  padding: 0.65rem 0.85rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
}

.chat-bubble--user {
  align-self: flex-end;
  background: var(--accent-soft);
  border-color: transparent;
}

.chat-bubble--assistant {
  align-self: flex-start;
  background: var(--bg-elevated);
}

.chat-bubble__role {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
  margin-bottom: 0.35rem;
}

.chat-bubble__body {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.925rem;
  line-height: 1.5;
}

.chat-bubble--typing {
  display: flex;
  gap: 0.35rem;
  padding: 0.75rem 1rem;
}

.typing-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--text-muted);
  animation: typing 1.2s infinite ease-in-out;
}

.typing-dot:nth-child(2) {
  animation-delay: 0.15s;
}
.typing-dot:nth-child(3) {
  animation-delay: 0.3s;
}

@keyframes typing {
  0%,
  80%,
  100% {
    opacity: 0.3;
    transform: translateY(0);
  }
  40% {
    opacity: 1;
    transform: translateY(-3px);
  }
}

.assistant-composer {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.composer-input {
  width: 100%;
  resize: vertical;
  min-height: 4.5rem;
  padding: 0.65rem 0.75rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text);
  font: inherit;
}

.composer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.err-line {
  margin: 0.35rem 0 0;
  color: var(--danger);
  font-size: 0.85rem;
}
</style>
