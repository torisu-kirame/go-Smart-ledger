<template>
  <div class="ledger-content panel" :class="{ 'ledger-content--expanded': expanded }">
    <div class="panel-head">
      <h3>账本内容 ({{ rows.length }})</h3>
      <div class="view-toggle" role="group" aria-label="展示样式">
        <button
          type="button"
          class="view-toggle__btn"
          :class="{ active: viewMode === 'table' }"
          title="表格样式（类似 xlsx）"
          @click="setView('table')"
        >
          <AppIcon name="grid" size="sm" />
          <span>表格</span>
        </button>
        <button
          type="button"
          class="view-toggle__btn"
          :class="{ active: viewMode === 'rows' }"
          title="分行列表，字段分列展示"
          @click="setView('rows')"
        >
          <AppIcon name="rows" size="sm" />
          <span>列表</span>
        </button>
      </div>
    </div>

    <div class="ledger-content__body">
      <div v-if="loading" class="muted empty-hint">加载中…</div>
      <div v-else-if="!rows.length" class="muted empty-hint">暂无记账记录，点击右下角「记一笔」添加</div>

      <!-- 表格样式：完整网格边框 -->
      <div v-else-if="viewMode === 'table'" class="table-wrap ledger-table">
      <table>
        <thead>
          <tr>
            <th v-for="col in columns" :key="col.key">{{ col.label }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rows" :key="row.seq">
            <td v-for="col in columns" :key="col.key" :class="{ mono: col.key === '_seq' }">
              {{ cellText(row, col) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 列表样式：仅横向分割线，列间距留白，值用 pill -->
    <div v-else class="ledger-rows">
      <div
        class="ledger-rows__head"
        :style="gridStyle"
      >
        <span v-for="col in columns" :key="col.key" class="ledger-rows__label">{{ col.label }}</span>
      </div>
      <div
        v-for="row in rows"
        :key="row.seq"
        class="ledger-rows__item"
        :style="gridStyle"
      >
        <span
          v-for="col in columns"
          :key="col.key"
          class="ledger-rows__cell"
          :class="{ 'ledger-rows__cell--primary': col.key === '_seq' }"
        >
          <span v-if="col.key === '_seq'" class="ledger-rows__seq">#{{ row.seq }}</span>
          <span v-else class="ledger-rows__pill">{{ cellText(row, col) }}</span>
        </span>
      </div>
    </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import AppIcon from './AppIcon.vue'
import {
  buildEntryRows,
  contentColumns,
  displayCell,
  getContentViewMode,
  setContentViewMode,
} from '../utils/ledgerEntries'

const props = defineProps({
  events: { type: Array, default: () => [] },
  schema: { type: Object, required: true },
  members: { type: Array, default: () => [] },
  groupKey: { type: String, default: '' },
  tableId: { type: String, default: '' },
  loading: { type: Boolean, default: false },
  /** 占满查看页主区域高度 */
  expanded: { type: Boolean, default: false },
})

const viewMode = ref(getContentViewMode())
const rows = ref([])

const columns = computed(() => contentColumns(props.schema))

const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${columns.value.length}, minmax(0, 1fr))`,
}))

function setView(mode) {
  viewMode.value = mode
  setContentViewMode(mode)
}

function cellText(row, col) {
  if (col.key === '_seq') return row.locked ? `#${row.seq}` : `#${row.seq}`
  if (!col.field) return '—'
  return displayCell(row, col.field, props.members)
}

async function refreshRows() {
  const tid = props.tableId || null
  rows.value = await buildEntryRows(props.events, props.schema, props.groupKey, tid)
}

watch(
  () => [props.events, props.schema, props.groupKey, props.tableId],
  () => {
    refreshRows()
  },
  { immediate: true, deep: true }
)
</script>

<style scoped>
.ledger-content .panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.85rem;
  flex-wrap: wrap;
}
.ledger-content .panel-head h3 {
  margin: 0;
}
.view-toggle {
  display: inline-flex;
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow: hidden;
  background: var(--bg-elevated);
}
.view-toggle__btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.4rem 0.75rem;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}
.view-toggle__btn + .view-toggle__btn {
  border-left: 1px solid var(--border);
}
.view-toggle__btn:hover {
  color: var(--text);
  background: var(--hover);
}
.view-toggle__btn.active {
  color: var(--accent);
  background: var(--accent-soft);
}
.ledger-content--expanded {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  margin-bottom: 0;
}
.ledger-content__body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.ledger-content--expanded .ledger-content__body {
  min-height: min(58vh, 720px);
}
.ledger-content--expanded .empty-hint {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}
.ledger-content--expanded .ledger-table,
.ledger-content--expanded .ledger-rows {
  flex: 1;
  min-height: 0;
  overflow: auto;
}
.empty-hint {
  padding: 1.5rem;
  text-align: center;
  font-size: 0.875rem;
}
.ledger-table table {
  min-width: 100%;
}
.ledger-table th,
.ledger-table td {
  white-space: nowrap;
}
.ledger-rows {
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--bg-card);
}
.ledger-rows__head,
.ledger-rows__item {
  display: grid;
  column-gap: 1.75rem;
  align-items: center;
  padding: 0.9rem 1.25rem;
}
.ledger-rows__head {
  border-bottom: 1px solid var(--border);
  background: var(--bg-elevated);
}
.ledger-rows__label {
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
}
.ledger-rows__item {
  border-bottom: 1px solid var(--border);
}
.ledger-rows__item:last-child {
  border-bottom: none;
}
.ledger-rows__item:hover {
  background: var(--hover);
}
.ledger-rows__cell {
  min-width: 0;
}
.ledger-rows__cell--primary .ledger-rows__seq {
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--text);
}
.ledger-rows__pill {
  display: inline-flex;
  max-width: 100%;
  padding: 0.32rem 0.6rem;
  border-radius: 6px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.8125rem;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
