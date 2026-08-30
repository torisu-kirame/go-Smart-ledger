<template>
  <div
    class="ledger-content panel"
    :class="{
      'ledger-content--expanded': expanded,
      'ledger-content--editing': editMode,
    }"
  >
    <div class="panel-head">
      <h3>
        账本内容 ({{ displayRows.length }})
        <span v-if="editMode" class="edit-badge">编辑中</span>
      </h3>
      <div class="view-toggle" role="group" aria-label="展示样式">
        <button
          type="button"
          class="view-toggle__btn"
          :class="{ active: viewMode === 'table' }"
          title="表格样式（类似 xlsx）"
          :disabled="editMode"
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
          :disabled="editMode"
          @click="setView('rows')"
        >
          <AppIcon name="rows" size="sm" />
          <span>列表</span>
        </button>
      </div>
    </div>

    <div class="ledger-content__body">
      <div v-if="loading" class="muted empty-hint">加载中…</div>
      <div
        v-else-if="!editMode && !displayRows.length"
        class="muted empty-hint"
      >
        暂无记账记录，点击右下角「记一笔」添加
      </div>

      <!-- 表格样式：仿 Excel / xlsx 工作表 -->
      <div v-else-if="editMode || viewMode === 'table'" class="xlsx-sheet">
        <div class="xlsx-sheet__bar" aria-hidden="true">
          <span class="xlsx-sheet__tab">{{ sheetTabLabel }}</span>
        </div>
        <div class="table-wrap xlsx-sheet__scroll">
          <table class="xlsx-table" :class="{ 'xlsx-table--edit': editMode }">
            <thead>
              <tr class="xlsx-letters">
                <th class="xlsx-corner" scope="col" />
                <th
                  v-for="(col, ci) in editColumns"
                  :key="`L-${col.key}`"
                  class="xlsx-letter"
                  scope="col"
                >
                  {{ colLetter(ci) }}
                </th>
                <th v-if="editMode" class="xlsx-letter xlsx-letter--action" scope="col" />
                <th v-if="editMode" class="xlsx-letter xlsx-letter--add" scope="col" />
              </tr>
              <tr class="xlsx-titles">
                <th class="xlsx-row-head" scope="col" />
                <th
                  v-for="col in editColumns"
                  :key="`T-${col.key}`"
                  class="xlsx-title"
                  scope="col"
                >
                  <template v-if="editMode && col.field">
                    <input
                      v-model="col.field.label"
                      class="xlsx-title-input"
                      @input="markDirty"
                    />
                  </template>
                  <template v-else>{{ col.label }}</template>
                </th>
                <th v-if="editMode" class="xlsx-title xlsx-title--action" scope="col" />
                <th v-if="editMode" class="xlsx-title xlsx-title--add" scope="col" />
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, ri) in displayRows"
                :key="row._key"
                class="xlsx-row"
                :class="{
                  'xlsx-row--locked': row.locked,
                  'xlsx-row--dragging': dragRowIndex === ri,
                  'xlsx-row--drag-over': dragOverIndex === ri,
                }"
              >
                <th
                  class="xlsx-row-head xlsx-row-head--drag"
                  scope="row"
                  :title="canReorderRows ? '长按拖动排序' : ''"
                  @pointerdown="onRowHeadPointerDown($event, ri)"
                >
                  {{ ri + 1 }}
                </th>
                <td
                  v-for="col in editColumns"
                  :key="col.key"
                  class="xlsx-cell"
                  :class="{
                    'xlsx-cell--seq': col.key === '_seq',
                    'xlsx-cell--num': isNumericCol(col),
                  }"
                  :title="editMode ? '' : cellText(row, col)"
                >
                  <template v-if="editMode && col.field && !row.locked">
                    <input
                      class="xlsx-cell-input"
                      :value="row.cells[col.key] ?? ''"
                      @input="onCellInput(row, col.key, $event.target.value)"
                    />
                  </template>
                  <template v-else>{{ cellText(row, col) }}</template>
                </td>
                <td v-if="editMode" class="xlsx-cell xlsx-cell--del">
                  <button
                    type="button"
                    class="xlsx-del-btn"
                    title="删除行"
                    @click="emitDeleteRow(ri)"
                  >
                    ×
                  </button>
                </td>
                <td
                  v-if="editMode && ri === 0"
                  class="xlsx-cell xlsx-cell--add-col"
                  :rowspan="Math.max(displayRows.length, 1)"
                >
                  <button
                    type="button"
                    class="xlsx-add-btn"
                    title="添加字段"
                    @click="emitAddColumn"
                  >
                    +
                  </button>
                </td>
              </tr>
              <tr v-if="editMode && !displayRows.length" class="xlsx-row">
                <td class="xlsx-cell" :colspan="editColumns.length + 1" />
                <td class="xlsx-cell xlsx-cell--del" />
                <td class="xlsx-cell xlsx-cell--add-col">
                  <button
                    type="button"
                    class="xlsx-add-btn"
                    title="添加字段"
                    @click="emitAddColumn"
                  >
                    +
                  </button>
                </td>
              </tr>
              <tr v-if="editMode" class="xlsx-row xlsx-row--add">
                <td
                  class="xlsx-cell xlsx-cell--add-row"
                  :colspan="editColumns.length + 2"
                >
                  <button
                    type="button"
                    class="xlsx-add-btn xlsx-add-btn--row"
                    title="添加行"
                    @click="emitAddRow"
                  >
                    +
                  </button>
                </td>
                <td class="xlsx-cell xlsx-cell--save">
                  <button
                    type="button"
                    class="xlsx-save-btn"
                    title="保存并退出编辑"
                    :disabled="saving"
                    @click="emitSave"
                  >
                    <AppIcon name="check" size="sm" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 列表样式：仅横向分割线，列间距留白，值用 pill -->
      <div v-else class="ledger-rows">
        <div class="ledger-rows__head" :style="gridStyle">
          <span v-for="col in columns" :key="col.key" class="ledger-rows__label">{{
            col.label
          }}</span>
        </div>
        <div
          v-for="row in displayRows"
          :key="row._key || row.seq"
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
  rowOrder: { type: Array, default: null },
  loading: { type: Boolean, default: false },
  expanded: { type: Boolean, default: false },
  editMode: { type: Boolean, default: false },
  /** draft rows when editing: { _key, seq?, cells, locked?, dirty? } */
  draftRows: { type: Array, default: null },
  draftFields: { type: Array, default: null },
  saving: { type: Boolean, default: false },
  canReorderRows: { type: Boolean, default: true },
})

const emit = defineEmits([
  'add-column',
  'add-row',
  'delete-row',
  'save',
  'dirty',
  'reorder-rows',
  'update-cell',
  'update-field-label',
])

const viewMode = ref(getContentViewMode())
const rows = ref([])
const dragRowIndex = ref(-1)
const dragOverIndex = ref(-1)
let longPressTimer = null
let longPressArmed = false
let pointerStartY = 0

const columns = computed(() => contentColumns(props.schema))
const editColumns = computed(() => {
  if (props.editMode && props.draftFields) {
    const cols = [{ key: '_seq', label: '#', fixed: true }]
    for (const f of props.draftFields) {
      cols.push({ key: f.key, label: f.label || f.key, field: f })
    }
    return cols
  }
  return columns.value
})

const displayRows = computed(() => {
  if (props.editMode && props.draftRows) {
    return props.draftRows
  }
  return rows.value.map((r) => ({ ...r, _key: `s-${r.seq}` }))
})

const sheetTabLabel = computed(() => {
  const n = editColumns.value.filter((c) => c.field).length
  return n ? `Sheet · ${n} 列` : 'Sheet'
})

const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${columns.value.length}, minmax(0, 1fr))`,
}))

function setView(mode) {
  if (props.editMode) return
  viewMode.value = mode
  setContentViewMode(mode)
}

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
  if (col.key === '_seq') return row.seq != null ? String(row.seq) : ''
  if (!col.field) return '—'
  if (props.editMode) return row.cells?.[col.key] || '—'
  return displayCell(row, col.field, props.members)
}

function markDirty() {
  emit('dirty')
}

function onCellInput(row, key, value) {
  emit('update-cell', { key: row._key, fieldKey: key, value })
  markDirty()
}

function emitAddColumn() {
  emit('add-column')
}
function emitAddRow() {
  emit('add-row')
}
function emitDeleteRow(ri) {
  emit('delete-row', ri)
}
function emitSave() {
  emit('save')
}

function clearLongPress() {
  if (longPressTimer) {
    clearTimeout(longPressTimer)
    longPressTimer = null
  }
  longPressArmed = false
}

function onRowHeadPointerDown(e, ri) {
  if (!props.canReorderRows || displayRows.value.length < 2) return
  if (e.button != null && e.button !== 0) return
  pointerStartY = e.clientY
  clearLongPress()
  longPressTimer = setTimeout(() => {
    longPressArmed = true
    dragRowIndex.value = ri
    try {
      e.target.setPointerCapture?.(e.pointerId)
    } catch {
      /* ignore */
    }
  }, 420)
  const onMove = (ev) => {
    if (!longPressArmed) {
      if (Math.abs(ev.clientY - pointerStartY) > 8) clearLongPress()
      return
    }
    const el = document.elementFromPoint(ev.clientX, ev.clientY)
    const th = el?.closest?.('.xlsx-row')
    if (!th) return
    const tbody = th.parentElement
    if (!tbody) return
    const idx = [...tbody.querySelectorAll('.xlsx-row:not(.xlsx-row--add)')].indexOf(th)
    if (idx >= 0) dragOverIndex.value = idx
  }
  const onUp = () => {
    window.removeEventListener('pointermove', onMove)
    window.removeEventListener('pointerup', onUp)
    window.removeEventListener('pointercancel', onUp)
    const from = dragRowIndex.value
    const to = dragOverIndex.value
    clearLongPress()
    dragRowIndex.value = -1
    dragOverIndex.value = -1
    if (from >= 0 && to >= 0 && from !== to) {
      emit('reorder-rows', { from, to })
      if (props.editMode) markDirty()
    }
  }
  window.addEventListener('pointermove', onMove)
  window.addEventListener('pointerup', onUp)
  window.addEventListener('pointercancel', onUp)
}

async function refreshRows() {
  const tid = props.tableId || null
  rows.value = await buildEntryRows(
    props.events,
    props.schema,
    props.groupKey,
    tid,
    props.rowOrder
  )
}

watch(
  () => [props.events, props.schema, props.groupKey, props.tableId, props.rowOrder],
  () => {
    refreshRows()
  },
  { immediate: true, deep: true }
)

watch(
  () => props.editMode,
  (on) => {
    if (on) viewMode.value = 'table'
  }
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
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}
.edit-badge {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: var(--accent-soft);
  color: var(--accent);
  text-transform: none;
  letter-spacing: 0;
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
.view-toggle__btn:hover:not(:disabled) {
  color: var(--text);
  background: var(--hover);
}
.view-toggle__btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
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
.xlsx-letter--add,
.xlsx-title--add,
.xlsx-cell--add-col {
  width: 2.25rem;
  min-width: 2.25rem;
  max-width: 2.25rem;
  padding: 0 !important;
  text-align: center !important;
  background: color-mix(in srgb, var(--xlsx-gutter-bg) 65%, var(--xlsx-cell-bg)) !important;
}
.xlsx-letter--action,
.xlsx-title--action {
  width: 2.1rem;
  min-width: 2.1rem;
  max-width: 2.1rem;
}
.xlsx-cell--add-col {
  vertical-align: middle;
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
.xlsx-title-input {
  width: 100%;
  border: none;
  background: transparent;
  color: inherit;
  font: inherit;
  font-weight: 600;
  padding: 0;
  outline: none;
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
.xlsx-row-head--drag {
  cursor: grab;
  touch-action: none;
}
.xlsx-row--dragging {
  opacity: 0.55;
}
.xlsx-row--drag-over .xlsx-cell,
.xlsx-row--drag-over .xlsx-row-head {
  box-shadow: inset 0 2px 0 var(--xlsx-accent);
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
.xlsx-cell-input {
  width: 100%;
  height: 100%;
  border: none;
  background: transparent;
  color: inherit;
  font: inherit;
  padding: 0;
  outline: none;
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
.xlsx-cell--del,
.xlsx-cell--save,
.xlsx-cell--add-col-spacer {
  width: 2.1rem;
  min-width: 2.1rem;
  max-width: 2.1rem;
  padding: 0 !important;
  text-align: center !important;
}
.xlsx-cell--add-row {
  text-align: center !important;
  background: color-mix(in srgb, var(--xlsx-gutter-bg) 70%, var(--xlsx-cell-bg)) !important;
  min-width: 8rem;
  height: 2rem;
}
.xlsx-add-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  min-height: 1.55rem;
  border: none;
  background: transparent;
  color: var(--xlsx-muted);
  font-size: 1.1rem;
  font-weight: 500;
  cursor: pointer;
  line-height: 1;
}
.xlsx-add-btn:hover {
  color: var(--xlsx-accent);
  background: var(--xlsx-select);
}
.xlsx-add-btn--row {
  min-height: 2rem;
  font-size: 1.25rem;
}
.xlsx-del-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  border: none;
  background: transparent;
  color: var(--xlsx-muted);
  font-size: 1.05rem;
  line-height: 1;
  cursor: pointer;
}
.xlsx-del-btn:hover {
  color: #e85d5d;
  background: color-mix(in srgb, #e85d5d 18%, transparent);
}
.xlsx-save-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  min-height: 2rem;
  border: none;
  background: color-mix(in srgb, var(--xlsx-accent) 22%, transparent);
  color: var(--xlsx-accent);
  cursor: pointer;
}
.xlsx-save-btn:hover:not(:disabled) {
  background: var(--xlsx-accent);
  color: #fff;
}
.xlsx-save-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
.xlsx-row--add:hover .xlsx-cell {
  background: inherit !important;
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
