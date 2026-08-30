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

      <!-- 表格样式：仿 Excel / xlsx 工作表 -->
      <div v-else-if="viewMode === 'table'" class="xlsx-sheet">
        <div class="xlsx-sheet__bar" aria-hidden="true">
          <span class="xlsx-sheet__tab">{{ sheetTabLabel }}</span>
        </div>
        <div class="table-wrap xlsx-sheet__scroll">
          <table class="xlsx-table">
            <thead>
              <tr class="xlsx-letters">
                <th class="xlsx-corner" scope="col" />
                <th
                  v-for="(col, ci) in columns"
                  :key="`L-${col.key}`"
                  class="xlsx-letter"
                  scope="col"
                >
                  {{ colLetter(ci) }}
                </th>
              </tr>
              <tr class="xlsx-titles">
                <th class="xlsx-row-head" scope="col" />
                <th
                  v-for="col in columns"
                  :key="`T-${col.key}`"
                  class="xlsx-title"
                  scope="col"
                >
                  {{ col.label }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, ri) in rows"
                :key="row.seq"
                class="xlsx-row"
                :class="{ 'xlsx-row--locked': row.locked }"
              >
                <th class="xlsx-row-head" scope="row">{{ ri + 1 }}</th>
                <td
                  v-for="col in columns"
                  :key="col.key"
                  class="xlsx-cell"
                  :class="{
                    'xlsx-cell--seq': col.key === '_seq',
                    'xlsx-cell--num': isNumericCol(col),
                  }"
                  :title="cellText(row, col)"
                >
                  {{ cellText(row, col) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
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
const sheetTabLabel = computed(() => {
  const n = (props.schema?.fields || []).length
  return n ? `Sheet · ${n} 列` : 'Sheet'
})

const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${columns.value.length}, minmax(0, 1fr))`,
}))

function setView(mode) {
  viewMode.value = mode
  setContentViewMode(mode)
}

/** Excel 列标：0→A, 25→Z, 26→AA */
function colLetter(index) {
  let n = index + 1
  let s = ''
  while (n > 0) {
    n -= 1
    s = String.fromCharCode(65 + (n % 26)) + s
    n = Math.floor(n / 26)
  }
  return s
}

function isNumericCol(col) {
  const t = col.field?.type
  return t === 'number' || t === 'amount' || t === 'money'
}

function cellText(row, col) {
  if (col.key === '_seq') return String(row.seq ?? '')
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
.ledger-content--expanded .xlsx-sheet,
.ledger-content--expanded .ledger-rows {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.ledger-content--expanded .xlsx-sheet__scroll,
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

/* —— 工作表样式：网格线 + 跟随项目主题色 —— */
.xlsx-sheet {
  --xlsx-grid: color-mix(in srgb, var(--border) 85%, var(--text-muted));
  --xlsx-letter-bg: var(--bg-elevated);
  --xlsx-title-bg: color-mix(in srgb, var(--accent-soft) 55%, var(--bg-elevated));
  --xlsx-gutter-bg: var(--bg-elevated);
  --xlsx-cell-bg: var(--bg-card);
  --xlsx-text: var(--text);
  --xlsx-muted: var(--text-muted);
  --xlsx-hover: var(--hover);
  --xlsx-select: var(--accent-soft);
  --xlsx-accent: var(--accent);
  --xlsx-accent-dim: var(--accent-dim);
  --xlsx-bar: var(--bg-elevated);
  border: 1px solid var(--xlsx-grid);
  border-radius: 2px;
  overflow: hidden;
  background: var(--xlsx-cell-bg);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--accent) 8%, transparent);
  min-height: 0;
}
.xlsx-sheet__bar {
  display: flex;
  align-items: flex-end;
  gap: 0;
  padding: 0.35rem 0.35rem 0;
  background: var(--xlsx-bar);
  border-bottom: 1px solid var(--xlsx-grid);
  min-height: 1.75rem;
}
.xlsx-sheet__tab {
  display: inline-flex;
  align-items: center;
  padding: 0.28rem 0.85rem 0.32rem;
  background: var(--xlsx-accent);
  color: #fff;
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: 2px 2px 0 0;
  box-shadow: inset 0 -2px 0 var(--xlsx-accent-dim);
  letter-spacing: 0.01em;
}
.xlsx-sheet__scroll {
  overflow: auto;
  border: none;
  border-radius: 0;
  background: var(--xlsx-cell-bg);
}
.xlsx-table {
  width: max-content;
  min-width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
  font-size: 0.8125rem;
  font-family: Calibri, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  color: var(--xlsx-text);
  background: var(--xlsx-cell-bg);
}
.xlsx-table th,
.xlsx-table td {
  border: 1px solid var(--xlsx-grid);
  padding: 0;
  text-align: left;
  vertical-align: middle;
  white-space: nowrap;
  text-transform: none;
  letter-spacing: normal;
  font-weight: 400;
  color: var(--xlsx-text);
  background: var(--xlsx-cell-bg);
}
.xlsx-corner,
.xlsx-row-head,
.xlsx-letter {
  position: sticky;
  z-index: 3;
  width: 2.5rem;
  min-width: 2.5rem;
  max-width: 2.5rem;
  text-align: center !important;
  font-size: 0.7rem;
  font-weight: 600 !important;
  color: var(--xlsx-muted) !important;
  background: var(--xlsx-gutter-bg) !important;
  user-select: none;
}
.xlsx-corner {
  left: 0;
  top: 0;
  z-index: 5;
}
.xlsx-letter {
  top: 0;
  z-index: 4;
  height: 1.35rem;
  background: var(--xlsx-letter-bg) !important;
  border-bottom-color: var(--xlsx-grid) !important;
}
.xlsx-title {
  position: sticky;
  top: 1.35rem;
  z-index: 4;
  height: 1.7rem;
  padding: 0 0.55rem !important;
  background: var(--xlsx-title-bg) !important;
  font-weight: 600 !important;
  font-size: 0.75rem !important;
  color: var(--xlsx-text) !important;
  box-shadow: inset 0 -2px 0 var(--xlsx-accent);
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 11rem;
  min-width: 5.5rem;
}
.xlsx-titles .xlsx-row-head {
  top: 1.35rem;
  left: 0;
  z-index: 5;
  box-shadow: inset 0 -2px 0 var(--xlsx-accent);
}
.xlsx-row-head {
  left: 0;
  z-index: 3;
  height: 1.55rem;
  font-variant-numeric: tabular-nums;
}
.xlsx-cell {
  height: 1.55rem;
  padding: 0 0.55rem !important;
  max-width: 14rem;
  min-width: 5.5rem;
  overflow: hidden;
  text-overflow: ellipsis;
  background: var(--xlsx-cell-bg) !important;
}
.xlsx-cell--seq {
  font-variant-numeric: tabular-nums;
  font-family: ui-monospace, Consolas, 'Courier New', monospace;
  font-size: 0.75rem;
  color: var(--xlsx-muted) !important;
  text-align: right !important;
  min-width: 3.5rem;
  max-width: 5rem;
}
.xlsx-cell--num {
  text-align: right !important;
  font-variant-numeric: tabular-nums;
}
.xlsx-row:hover .xlsx-cell {
  background: var(--xlsx-hover) !important;
}
.xlsx-row:hover .xlsx-row-head {
  background: var(--xlsx-select) !important;
  color: var(--xlsx-text) !important;
}
.xlsx-row--locked .xlsx-cell {
  color: var(--xlsx-muted) !important;
  font-style: italic;
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
