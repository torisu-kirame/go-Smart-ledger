<template>
  <div class="page assistant-page">
    <PageHeader :crumbs="crumbs" />

    <div v-if="!aiEnabled" class="alert alert-warn">
      <span>{{ t(`assistant.${aiBlockReason}`) }}</span>
      <router-link to="/settings#ai" class="assistant-settings-link">
        {{ t('assistant.openSettings') }}
      </router-link>
    </div>

    <div class="assistant-layout" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
      <aside class="panel agent-rail">
        <div class="rail-head">
          <h3 v-if="!sidebarCollapsed" class="rail-title">{{ t('assistant.agentsTitle') }}</h3>
          <button
            type="button"
            class="icon-btn btn-ghost btn-sm rail-toggle"
            :title="sidebarCollapsed ? t('assistant.expandSidebar') : t('assistant.collapseSidebar')"
            @click="toggleSidebar"
          >
            <AppIcon :name="sidebarCollapsed ? 'chevron-right' : 'chevron-left'" size="sm" />
          </button>
        </div>

        <ul class="agent-list">
          <li
            v-for="agent in agentsState.agents"
            :key="agent.id"
            class="agent-item"
            :class="{ active: agent.id === agentsState.activeId }"
          >
            <button
              type="button"
              class="agent-item__btn"
              :title="agent.name"
              @click="switchAgent(agent.id)"
            >
              <AppIcon name="sparkles" size="sm" />
              <span v-if="!sidebarCollapsed" class="agent-item__name">{{ agent.name }}</span>
            </button>
            <DeleteButton
              v-if="!sidebarCollapsed && agentsState.agents.length > 1"
              icon-only
              sm
              :title="t('assistant.deleteAgent')"
              @click="deleteAgent(agent.id)"
            />
          </li>
        </ul>

        <button
          type="button"
          class="icon-btn btn-ghost btn-sm rail-add"
          :title="t('assistant.newAgent')"
          :disabled="!aiEnabled"
          @click="openNewAgentModal"
        >
          <AppIcon name="plus" size="sm" />
          <span v-if="!sidebarCollapsed">{{ t('assistant.newAgent') }}</span>
        </button>
      </aside>

      <div class="assistant-main">
        <header class="chat-topbar">
          <div class="chat-topbar__meta">
            <AppIcon name="sparkles" size="sm" class="chat-topbar__icon" />
            <span class="chat-topbar__name">{{ activeAgent?.name || t('assistant.roleAssistant') }}</span>
          </div>
          <div class="chat-topbar__actions">
            <button
              type="button"
              class="icon-btn btn-ghost btn-sm topbar-btn"
              :class="{ 'topbar-btn--active': agentContextReady }"
              :title="t('assistant.ledgerContext')"
              @click="openLedgerModal"
            >
              <AppIcon name="ledger" size="sm" />
            </button>
            <button
              type="button"
              class="icon-btn btn-ghost btn-sm topbar-btn"
              :title="t('assistant.workspaceTitle')"
              :disabled="!activeAgent"
              @click="openWorkspaceModal"
            >
              <AppIcon name="file" size="sm" />
            </button>
          </div>
        </header>

        <div ref="threadEl" class="panel assistant-thread" role="log" aria-live="polite">
          <div v-if="!messages.length" class="thread-empty">
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
              {{ msg.role === 'user' ? t('assistant.roleUser') : activeAgent?.name || t('assistant.roleAssistant') }}
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
          <div class="composer-row">
            <textarea
              v-model="input"
              class="composer-input"
              rows="2"
              :placeholder="t('assistant.inputPlaceholder')"
              :disabled="!aiEnabled || streaming"
              @keydown.enter.exact.prevent="send"
            />
            <div class="composer-actions">
              <button
                v-if="streaming"
                type="button"
                class="icon-btn btn-ghost composer-icon-btn"
                :title="t('assistant.stop')"
                @click="stop"
              >
                <AppIcon name="square" size="sm" />
              </button>
              <button
                type="submit"
                class="icon-btn btn-primary composer-send"
                :title="t('assistant.send')"
                :disabled="!aiEnabled || streaming || !input.trim()"
              >
                <AppIcon name="plane" size="sm" />
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>

    <!-- 账本上下文弹窗：仅右上角关闭 -->
    <div v-if="showLedgerModal" class="modal modal--no-dismiss">
      <div class="modal-card modal-card--wide assistant-modal">
        <header class="assistant-modal__head">
          <h3>{{ t('assistant.ledgerModalTitle') }}</h3>
          <button type="button" class="modal-close-btn" :title="t('assistant.close')" @click="closeLedgerModal">
            <AppIcon name="x" size="sm" />
          </button>
        </header>
        <p class="assistant-modal__hint">{{ t('assistant.contextHint') }}</p>
        <p v-if="activeAgent" class="assistant-modal__agent">
          {{ t('assistant.forAgent') }}：{{ activeAgent.name }}
        </p>
        <div class="form-row">
          <label>{{ t('assistant.ledgerContext') }}</label>
          <select v-model="ledgerModalLedgerId" class="ledger-select" :disabled="loadingLedgers">
            <option value="">{{ t('assistant.noLedger') }}</option>
            <option v-for="lg in ledgers" :key="lg.id" :value="lg.id">
              {{ lg.name || lg.id }}
            </option>
          </select>
        </div>
        <div class="assistant-modal__toolbar">
          <button
            type="button"
            class="icon-btn btn-primary btn-sm"
            :disabled="!ledgerModalLedgerId || contextLoading"
            :title="contextLoading ? t('assistant.loadingContext') : t('assistant.syncContext')"
            @click="syncLedgerContext"
          >
            <AppIcon name="refresh" size="sm" />
            <span>{{ contextLoading ? t('assistant.loadingContext') : t('assistant.syncContext') }}</span>
          </button>
          <button
            v-if="ledgerModalLedgerId"
            type="button"
            class="icon-btn btn-ghost btn-sm"
            :title="t('assistant.clearLedgerContext')"
            @click="clearLedgerContext"
          >
            <AppIcon name="x" size="sm" />
            <span>{{ t('assistant.clearLedgerContext') }}</span>
          </button>
        </div>
        <p v-if="agentContextReady" class="ok-line">{{ t('assistant.contextReady') }}</p>
        <p v-if="contextError" class="err-line">{{ contextError }}</p>
      </div>
    </div>

    <!-- Agent 设定弹窗：仅右上角关闭 -->
    <div v-if="showWorkspaceModal" class="modal modal--no-dismiss">
      <div class="modal-card modal-card--wide assistant-modal assistant-modal--tall">
        <header class="assistant-modal__head">
          <h3>{{ t('assistant.workspaceTitle') }}</h3>
          <button type="button" class="modal-close-btn" :title="t('assistant.close')" @click="closeWorkspaceModal">
            <AppIcon name="x" size="sm" />
          </button>
        </header>
        <p class="assistant-modal__hint">{{ t('assistant.workspaceHint') }}</p>
        <p v-if="activeAgent" class="assistant-modal__agent">
          {{ t('assistant.forAgent') }}：{{ activeAgent.name }}
        </p>
        <div class="workspace-layout">
          <ul class="workspace-files">
            <li
              v-for="name in workspaceFileNames"
              :key="name"
              class="workspace-file"
              :class="{ active: selectedWorkspaceFile === name }"
            >
              <button type="button" @click="selectWorkspaceFile(name)">
                <AppIcon name="file" size="sm" />
                <span>{{ name }}</span>
              </button>
            </li>
          </ul>
          <textarea
            v-model="workspaceDraft"
            class="workspace-editor mono"
            spellcheck="false"
          />
        </div>
        <div class="assistant-modal__toolbar">
          <button
            type="button"
            class="icon-btn btn-ghost btn-sm"
            :title="t('assistant.workspaceReset')"
            @click="resetWorkspaceFiles"
          >
            <AppIcon name="refresh" size="sm" />
            <span>{{ t('assistant.workspaceReset') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 新建 Agent -->
    <div v-if="showAgentModal" class="modal modal--no-dismiss">
      <form class="modal-card assistant-modal" @submit.prevent="submitNewAgent">
        <header class="assistant-modal__head">
          <h3>{{ t('assistant.newAgentTitle') }}</h3>
          <button type="button" class="modal-close-btn" :title="t('assistant.close')" @click="closeAgentModal">
            <AppIcon name="x" size="sm" />
          </button>
        </header>
        <div class="form-row">
          <label>{{ t('assistant.agentName') }}</label>
          <input v-model="newAgentName" maxlength="32" required />
        </div>
        <p class="hint-text">{{ t('assistant.newAgentHint') }}</p>
        <div class="assistant-modal__toolbar">
          <button type="submit" class="icon-btn btn-primary btn-sm">
            <AppIcon name="plus" size="sm" />
            <span>{{ t('assistant.createAgent') }}</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import AppIcon from '../components/AppIcon.vue'
import DeleteButton from '../components/DeleteButton.vue'
import PageHeader from '../components/PageHeader.vue'
import { api } from '../api/http'
import {
  buildLedgerContextPrompt,
  defaultSystemMessages,
  streamChat,
} from '../api/aiChat'
import { useI18n } from '../composables/useI18n'
import { WORKSPACE_FILE_NAMES } from '../utils/agentWorkspace'
import {
  addAgent,
  agentSystemPrompt,
  getActiveAgent,
  loadAgentsState,
  removeAgent,
  resetAgentWorkspaceFiles,
  setActiveAgent,
  updateAgentLedgerContext,
  updateAgentMessages,
  updateAgentWorkspaceFiles,
} from '../utils/aiAgents'
import {
  AI_CONFIG_CHANGED_EVENT,
  aiConfigBlockReason,
  isAiConfigReady,
  loadAiConfig,
} from '../utils/aiConfig'

const SIDEBAR_KEY = 'smart-ledger-assistant-sidebar-collapsed'

const { t } = useI18n()

const crumbs = computed(() => [{ label: t('assistant.title') }])
const configRev = ref(0)
const aiCfg = computed(() => {
  configRev.value
  return loadAiConfig()
})
const aiEnabled = computed(() => isAiConfigReady(aiCfg.value))
const aiBlockReason = computed(() => aiConfigBlockReason(aiCfg.value) || 'disabledHint')

const sidebarCollapsed = ref(localStorage.getItem(SIDEBAR_KEY) === '1')

const agentsState = ref(loadAgentsState())
const activeAgent = computed(() => getActiveAgent(agentsState.value))
const messages = computed(() => activeAgent.value?.messages || [])
const agentContextReady = computed(() => !!activeAgent.value?.ledgerContextText?.trim())

const workspaceFileNames = WORKSPACE_FILE_NAMES
const selectedWorkspaceFile = ref(WORKSPACE_FILE_NAMES[0])
const workspaceDraft = ref('')

const ledgers = ref([])
const loadingLedgers = ref(false)
const ledgerModalLedgerId = ref('')
const contextLoading = ref(false)
const contextError = ref('')

const input = ref('')
const streaming = ref(false)
const chatError = ref('')
const threadEl = ref(null)
let abortCtrl = null

const showAgentModal = ref(false)
const showLedgerModal = ref(false)
const showWorkspaceModal = ref(false)
const newAgentName = ref('')

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
  localStorage.setItem(SIDEBAR_KEY, sidebarCollapsed.value ? '1' : '0')
}

function loadWorkspaceDraft() {
  const agent = activeAgent.value
  if (!agent) return
  const files = agent.workspaceFiles || {}
  workspaceDraft.value = files[selectedWorkspaceFile.value] || ''
}

function selectWorkspaceFile(name) {
  saveWorkspaceDraft()
  selectedWorkspaceFile.value = name
  loadWorkspaceDraft()
}

function saveWorkspaceDraft() {
  const id = agentsState.value.activeId
  if (!id) return
  updateAgentWorkspaceFiles(id, { [selectedWorkspaceFile.value]: workspaceDraft.value })
  refreshAgents()
}

function resetWorkspaceFiles() {
  if (!confirm(t('assistant.workspaceResetConfirm'))) return
  const id = agentsState.value.activeId
  resetAgentWorkspaceFiles(id)
  refreshAgents()
  loadWorkspaceDraft()
}

function refreshAgents() {
  agentsState.value = loadAgentsState()
}

function syncAgentLedgerFields() {
  const agent = activeAgent.value
  ledgerModalLedgerId.value = agent?.ledgerId || ''
  contextError.value = ''
}

function switchAgent(id) {
  if (streaming.value) stop()
  if (showWorkspaceModal.value) saveWorkspaceDraft()
  setActiveAgent(id)
  refreshAgents()
  chatError.value = ''
  syncAgentLedgerFields()
  loadWorkspaceDraft()
  showLedgerModal.value = false
  showWorkspaceModal.value = false
  nextTick(() => scrollToBottom())
}

function deleteAgent(id) {
  if (!confirm(t('assistant.deleteAgentConfirm'))) return
  removeAgent(id)
  refreshAgents()
  syncAgentLedgerFields()
  loadWorkspaceDraft()
}

function openNewAgentModal() {
  newAgentName.value = ''
  showAgentModal.value = true
}

function closeAgentModal() {
  showAgentModal.value = false
}

function submitNewAgent() {
  addAgent(newAgentName.value)
  refreshAgents()
  syncAgentLedgerFields()
  loadWorkspaceDraft()
  closeAgentModal()
}

function openLedgerModal() {
  syncAgentLedgerFields()
  contextError.value = ''
  showLedgerModal.value = true
  if (!ledgers.value.length) loadLedgers()
}

function closeLedgerModal() {
  showLedgerModal.value = false
}

function openWorkspaceModal() {
  loadWorkspaceDraft()
  showWorkspaceModal.value = true
}

function closeWorkspaceModal() {
  saveWorkspaceDraft()
  showWorkspaceModal.value = false
}

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

async function syncLedgerContext() {
  const agentId = agentsState.value.activeId
  if (!agentId || !ledgerModalLedgerId.value) return
  contextLoading.value = true
  contextError.value = ''
  try {
    const data = await api.ragExport(ledgerModalLedgerId.value)
    const name = ledgers.value.find((l) => l.id === ledgerModalLedgerId.value)?.name
    const text = buildLedgerContextPrompt(data, name)
    updateAgentLedgerContext(agentId, ledgerModalLedgerId.value, text)
    refreshAgents()
  } catch (e) {
    contextError.value = e?.message || String(e)
  } finally {
    contextLoading.value = false
  }
}

function clearLedgerContext() {
  const agentId = agentsState.value.activeId
  if (!agentId) return
  updateAgentLedgerContext(agentId, '', '')
  ledgerModalLedgerId.value = ''
  contextError.value = ''
  refreshAgents()
}

async function scrollToBottom() {
  await nextTick()
  if (threadEl.value) threadEl.value.scrollTop = threadEl.value.scrollHeight
}

function stop() {
  abortCtrl?.abort()
}

function persistMessages(next) {
  const id = agentsState.value.activeId
  updateAgentMessages(id, next)
  refreshAgents()
}

async function send() {
  const text = input.value.trim()
  if (!text || streaming.value || !aiEnabled.value || !activeAgent.value) return

  chatError.value = ''
  const next = [...messages.value, { role: 'user', content: text }]
  persistMessages(next)
  input.value = ''
  await scrollToBottom()

  const history = next.map((m) => ({ role: m.role, content: m.content }))
  const systemPrompt = agentSystemPrompt(activeAgent.value)
  const ledgerContext = activeAgent.value.ledgerContextText || ''
  const payload = [
    ...defaultSystemMessages(ledgerContext, systemPrompt),
    ...history,
  ]

  streaming.value = true
  const assistantIdx = next.length
  const withAssistant = [...next, { role: 'assistant', content: '' }]
  persistMessages(withAssistant)
  abortCtrl = new AbortController()

  try {
    await streamChat({
      messages: payload,
      signal: abortCtrl.signal,
      agentUser: `agent:${activeAgent.value.id}`,
      onDelta(delta) {
        const current = loadAgentsState()
        const agent = current.agents.find((a) => a.id === current.activeId)
        if (!agent?.messages[assistantIdx]) return
        agent.messages[assistantIdx].content += delta
        updateAgentMessages(current.activeId, [...agent.messages])
        refreshAgents()
        scrollToBottom()
      },
    })
  } catch (e) {
    const current = loadAgentsState()
    const agent = current.agents.find((a) => a.id === current.activeId)
    let msgs = agent?.messages || []
    if (e?.name === 'AbortError') {
      if (!msgs[assistantIdx]?.content) {
        msgs = msgs.slice(0, assistantIdx)
      }
    } else if (e?.message === 'AI_DISABLED') {
      chatError.value = t('assistant.disabledHint')
      msgs = msgs.slice(0, assistantIdx)
    } else if (e?.message === 'API_KEY_REQUIRED') {
      chatError.value = t('assistant.apiKeyRequired')
      msgs = msgs.slice(0, assistantIdx)
    } else if (e?.message === 'CONNECTION_NOT_VERIFIED') {
      chatError.value = t('assistant.connectionNotVerified')
      msgs = msgs.slice(0, assistantIdx)
    } else if (e?.message === 'AI_EMPTY_RESPONSE') {
      chatError.value = t('assistant.emptyResponse')
      msgs = msgs.slice(0, assistantIdx)
    } else {
      chatError.value = e?.message || String(e)
      msgs = msgs.slice(0, assistantIdx)
    }
    updateAgentMessages(current.activeId, msgs)
    refreshAgents()
  } finally {
    streaming.value = false
    abortCtrl = null
    await scrollToBottom()
  }
}

function refreshAiGate() {
  configRev.value++
}

onMounted(() => {
  loadLedgers()
  syncAgentLedgerFields()
  loadWorkspaceDraft()
  window.addEventListener(AI_CONFIG_CHANGED_EVENT, refreshAiGate)
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') refreshAiGate()
  })
})

onUnmounted(() => {
  window.removeEventListener(AI_CONFIG_CHANGED_EVENT, refreshAiGate)
})

watch(
  () => agentsState.value.activeId,
  () => {
    syncAgentLedgerFields()
    loadWorkspaceDraft()
  }
)
</script>

<style scoped>
.assistant-page {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  max-width: none;
  height: calc(100vh - 3rem);
  min-height: 0;
}

.assistant-settings-link {
  margin-left: 0.5rem;
  color: var(--accent);
  text-decoration: underline;
}

.assistant-layout {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: min(220px, 26%) minmax(0, 1fr);
  gap: 0.75rem;
  align-items: stretch;
  transition: grid-template-columns 0.2s ease;
}

.assistant-layout.sidebar-collapsed {
  grid-template-columns: 3.25rem minmax(0, 1fr);
}

@media (max-width: 700px) {
  .assistant-layout {
    grid-template-columns: 3.25rem minmax(0, 1fr);
  }
  .assistant-layout:not(.sidebar-collapsed) {
    grid-template-columns: min(180px, 38%) minmax(0, 1fr);
  }
}

.agent-rail {
  display: flex;
  flex-direction: column;
  padding: 0.55rem;
  min-height: 0;
  overflow: hidden;
}

.rail-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.35rem;
  margin-bottom: 0.5rem;
  min-height: 1.75rem;
}

.rail-title {
  margin: 0;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  overflow: hidden;
  white-space: nowrap;
}

.rail-toggle {
  margin-left: auto;
  flex-shrink: 0;
}

.sidebar-collapsed .rail-head {
  justify-content: center;
}

.agent-list {
  list-style: none;
  margin: 0;
  padding: 0;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.agent-item {
  display: flex;
  align-items: center;
  gap: 0.2rem;
  margin-bottom: 0.2rem;
}

.agent-item.active .agent-item__btn {
  background: var(--accent-soft);
  color: var(--accent);
  border-color: color-mix(in srgb, var(--accent) 45%, var(--border));
}

.agent-item__btn {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.45rem 0.5rem;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text);
  font-size: 0.875rem;
  cursor: pointer;
  text-align: left;
}

.sidebar-collapsed .agent-item__btn {
  justify-content: center;
  padding: 0.45rem;
}

.agent-item__btn:hover {
  background: var(--hover);
}

.agent-item__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rail-add {
  margin-top: 0.35rem;
  width: 100%;
  justify-content: center;
}

.sidebar-collapsed .rail-add span {
  display: none;
}

.assistant-main {
  display: flex;
  flex-direction: column;
  min-height: 0;
  min-width: 0;
  gap: 0.5rem;
}

.chat-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.35rem 0.15rem;
  flex-shrink: 0;
}

.chat-topbar__meta {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  min-width: 0;
}

.chat-topbar__icon {
  color: var(--accent);
  flex-shrink: 0;
}

.chat-topbar__name {
  font-weight: 600;
  font-size: 0.95rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-topbar__actions {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  flex-shrink: 0;
}

.topbar-btn--active {
  color: var(--success);
  background: color-mix(in srgb, var(--success) 12%, transparent);
}

.assistant-thread {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem;
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
  max-width: min(92%, 42rem);
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
  flex-shrink: 0;
  padding: 0.65rem 0.75rem;
}

.composer-row {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
}

.composer-input {
  flex: 1;
  resize: none;
  min-height: 2.75rem;
  max-height: 8rem;
  padding: 0.6rem 0.7rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text);
  font: inherit;
  line-height: 1.45;
}

.composer-actions {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  flex-shrink: 0;
}

.composer-icon-btn,
.composer-send {
  width: 2.5rem;
  height: 2.5rem;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
}

.composer-send:disabled {
  opacity: 0.45;
}

.modal--no-dismiss {
  pointer-events: auto;
}

.assistant-modal {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  width: min(100%, 36rem);
}

.assistant-modal--tall {
  width: min(100%, 52rem);
  max-height: 88vh;
}

.assistant-modal__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.assistant-modal__head h3 {
  margin: 0;
  font-size: 1.05rem;
}

.modal-close-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  padding: 0;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-elevated);
  color: var(--text-muted);
  cursor: pointer;
}

.modal-close-btn:hover {
  color: var(--text);
  background: var(--hover);
}

.assistant-modal__hint,
.assistant-modal__agent {
  margin: 0;
  font-size: 0.82rem;
  color: var(--text-muted);
}

.assistant-modal__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.25rem;
}

.ledger-select {
  width: 100%;
  padding: 0.45rem 0.6rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text);
}

.workspace-layout {
  display: grid;
  grid-template-columns: min(168px, 30%) minmax(0, 1fr);
  gap: 0.75rem;
  min-height: 16rem;
  flex: 1;
}

.workspace-files {
  list-style: none;
  margin: 0;
  padding: 0;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  overflow: auto;
  max-height: 22rem;
}

.workspace-file button {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.35rem;
  text-align: left;
  padding: 0.45rem 0.55rem;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.75rem;
  cursor: pointer;
}

.workspace-file.active button {
  background: var(--accent-soft);
  color: var(--accent);
  font-weight: 600;
}

.workspace-editor {
  width: 100%;
  min-height: 16rem;
  max-height: 22rem;
  padding: 0.65rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg);
  color: var(--text);
  font-size: 0.78rem;
  line-height: 1.45;
  resize: vertical;
}

.ok-line {
  margin: 0;
  color: var(--success);
  font-size: 0.85rem;
}

.err-line {
  margin: 0;
  color: var(--danger);
  font-size: 0.85rem;
}

.hint-text {
  margin: 0;
  font-size: 0.82rem;
  color: var(--text-muted);
}
</style>
