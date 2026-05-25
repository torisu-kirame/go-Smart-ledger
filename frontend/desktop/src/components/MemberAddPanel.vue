<template>
  <div class="member-add">
    <div class="tabs">
      <button type="button" class="tab" :class="{ active: mode === 'id' }" @click="mode = 'id'">
        输入用户 ID
      </button>
      <button type="button" class="tab" :class="{ active: mode === 'friends' }" @click="openFriends">
        从好友选择
      </button>
    </div>

    <div v-if="mode === 'id'" class="pane">
      <template v-if="multiple">
        <div v-for="(row, i) in idRows" :key="'id-' + i" class="member-row">
          <input v-model="row.id" placeholder="成员用户 ID" @input="emitMerged" />
          <DeleteButton
            v-if="idRows.length > 1"
            icon-only
            sm
            title="删除该行"
            @click="removeIdRow(i)"
          />
        </div>
        <button type="button" class="btn-ghost" @click="addIdRow">+ 输入 ID</button>
      </template>
      <div v-else class="member-row">
        <input v-model="singleId" placeholder="用户 ID" @input="emitSingle" />
      </div>
    </div>

    <div v-else class="pane">
      <button v-if="!friendsLoaded" type="button" class="btn-ghost" :disabled="loadingFriends" @click="loadFriends">
        {{ loadingFriends ? '加载中…' : '加载好友列表' }}
      </button>
      <p v-else-if="friendError" class="err">{{ friendError }}</p>
      <p v-else-if="!availableFriends.length" class="muted">暂无好友，请先在「好友」页添加</p>
      <div v-else class="friend-list">
        <label
          v-for="f in availableFriends"
          :key="f.id"
          class="check-row"
          :class="{ picked: isPicked(f.id) }"
        >
          <input
            :type="multiple ? 'checkbox' : 'radio'"
            :name="multiple ? undefined : 'pick-friend'"
            :checked="isPicked(f.id)"
            @change="toggleFriend(f.id, $event)"
          />
          <img class="mini-avatar" :src="avatarUrl(f.id)" alt="" />
          <span class="friend-label">{{ f.nickname || f.username }}</span>
          <span class="mono friend-id">{{ f.id }}</span>
        </label>
      </div>
    </div>

    <p v-if="multiple && mergedPreview.length" class="picked-summary">
      已选 {{ mergedPreview.length }} 人：
      <span class="mono">{{ mergedPreview.join('、') }}</span>
    </p>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import DeleteButton from './DeleteButton.vue'
import { api } from '../api/http'

const props = defineProps({
  modelValue: {
    type: [Array, String],
    default: () => [],
  },
  multiple: {
    type: Boolean,
    default: true,
  },
  excludeIds: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['update:modelValue'])

const mode = ref('id')
const idRows = ref([{ id: '' }])
const singleId = ref('')
const friends = ref([])
const picked = ref([])
const loadingFriends = ref(false)
const friendsLoaded = ref(false)
const friendError = ref('')

const excludeSet = computed(() => new Set((props.excludeIds || []).filter(Boolean)))

const availableFriends = computed(() =>
  friends.value.filter((f) => f.id && !excludeSet.value.has(f.id))
)

const mergedPreview = computed(() => {
  if (!props.multiple) return []
  return normalizeIds([
    ...idRows.value.map((r) => r.id),
    ...picked.value,
  ])
})

function avatarUrl(userId) {
  return api.userAvatarUrl(userId)
}

function normalizeIds(list) {
  const out = []
  const seen = new Set()
  for (const raw of list) {
    const id = String(raw || '').trim()
    if (!id || excludeSet.value.has(id) || seen.has(id)) continue
    seen.add(id)
    out.push(id)
  }
  return out
}

function isPicked(id) {
  return picked.value.includes(id)
}

function emitMerged() {
  emit('update:modelValue', mergedPreview.value)
}

function emitSingle() {
  const id = singleId.value.trim()
  emit('update:modelValue', excludeSet.value.has(id) ? '' : id)
}

function addIdRow() {
  idRows.value.push({ id: '' })
}

function removeIdRow(i) {
  idRows.value.splice(i, 1)
  emitMerged()
}

function toggleFriend(id, ev) {
  if (props.multiple) {
    if (ev.target.checked) {
      if (!picked.value.includes(id)) picked.value.push(id)
    } else {
      picked.value = picked.value.filter((x) => x !== id)
    }
    emitMerged()
    return
  }
  if (ev.target.checked) {
    picked.value = [id]
    singleId.value = id
    emit('update:modelValue', id)
  }
}

async function loadFriends() {
  loadingFriends.value = true
  friendError.value = ''
  try {
    const res = await api.listFriends()
    friends.value = res.friends || res || []
    friendsLoaded.value = true
  } catch (e) {
    friendError.value = e?.message || '加载好友失败'
  } finally {
    loadingFriends.value = false
  }
}

function openFriends() {
  mode.value = 'friends'
  if (!friendsLoaded.value) loadFriends()
}

watch(
  () => props.modelValue,
  (v) => {
    if (props.multiple) {
      const ids = Array.isArray(v) ? v : []
      if (!ids.length) return
      picked.value = ids.filter((id) => availableFriends.value.some((f) => f.id === id))
      const manual = ids.filter((id) => !picked.value.includes(id))
      idRows.value = manual.length ? manual.map((id) => ({ id })) : [{ id: '' }]
    } else {
      singleId.value = typeof v === 'string' ? v : ''
      picked.value = singleId.value ? [singleId.value] : []
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.member-add {
  display: grid;
  gap: 0.75rem;
}
.tabs {
  display: flex;
  gap: 0.5rem;
}
.tab {
  padding: 0.35rem 0.75rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.875rem;
}
.tab.active {
  background: var(--accent-soft);
  border-color: var(--accent);
  color: var(--accent);
  font-weight: 600;
}
.pane {
  display: grid;
  gap: 0.5rem;
}
.member-row {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
}
.member-row input {
  flex: 1;
  min-width: 10rem;
}
.friend-list {
  display: grid;
  gap: 0.35rem;
  max-height: 220px;
  overflow-y: auto;
}
.check-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.45rem 0.55rem;
  border-radius: 8px;
  border: 1px solid transparent;
  cursor: pointer;
}
.check-row:hover {
  background: var(--hover);
}
.check-row.picked {
  border-color: var(--accent);
  background: var(--accent-soft);
}
.mini-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--border);
}
.friend-label {
  font-weight: 500;
}
.friend-id {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-left: auto;
}
.picked-summary {
  font-size: 0.875rem;
  color: var(--text-muted);
  margin: 0;
  word-break: break-all;
}
.err {
  color: var(--danger);
  font-size: 0.875rem;
}
.muted {
  color: var(--text-muted);
  font-size: 0.875rem;
}
</style>
