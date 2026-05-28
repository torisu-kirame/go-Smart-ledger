<template>
  <section class="detail-card">
    <h3 class="detail-card__title">会计科目表（自定义）</h3>
    <p class="field-hint">可新增、编辑或停用科目；保存前会在本地校验编码与类别。</p>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>编码</th>
            <th>名称</th>
            <th>类别</th>
            <th>启用</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(a, i) in chart.accounts" :key="a.code + '-' + i">
            <td>
              <input v-model="a.code" class="field-sm mono" />
            </td>
            <td>
              <input v-model="a.name" class="field-sm" />
            </td>
            <td>
              <AppSelect v-model="a.category" :options="categoryOptions" sm />
            </td>
            <td>
              <ToggleSwitch
                v-model="a.active"
                :aria-label="`科目 ${a.code} 启用`"
              />
            </td>
            <td>
              <DeleteButton icon-only sm title="删除科目" @click="removeAccount(i)" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="coa-add">
      <input v-model="draft.code" placeholder="编码" class="field-sm mono" />
      <input v-model="draft.name" placeholder="名称" class="field-sm" />
      <AppSelect v-model="draft.category" :options="categoryOptions" sm placeholder="类别" />
      <button type="button" class="btn-ghost" @click="addAccount">添加科目</button>
    </div>

    <div class="actions-row" style="margin-top: 0.85rem">
      <button type="button" class="btn-primary" :disabled="busy" @click="save">保存科目表</button>
      <button type="button" class="btn-ghost" :disabled="busy" @click="reset">恢复默认科目表</button>
    </div>
  </section>
</template>

<script setup>
import { reactive } from 'vue'
import AppSelect from '../../components/AppSelect.vue'
import DeleteButton from '../../components/DeleteButton.vue'
import ToggleSwitch from '../../components/ToggleSwitch.vue'
import { useLedgerAccounting } from '../../composables/provideAccounting'
import { useLedgerDetail } from '../../composables/useLedgerDetail'
import { useNotify } from '../../composables/useNotify'
import {
  ACCOUNT_CATEGORIES,
  validateChartPayload,
} from '../../utils/chartOfAccounts'

const { chart, busy, saveChart, resetChart } = useLedgerAccounting()
const { ledgerId } = useLedgerDetail()
const notify = useNotify()

const categoryOptions = ACCOUNT_CATEGORIES.map((c) => ({ value: c.value, label: c.label }))

const draft = reactive({
  code: '',
  name: '',
  category: 'asset',
  active: true,
})

function addAccount() {
  const code = draft.code.trim()
  const name = draft.name.trim()
  if (!code || !name) {
    notify.error('请填写科目编码与名称')
    return
  }
  if ((chart.value.accounts || []).some((a) => a.code === code)) {
    notify.error('科目编码已存在')
    return
  }
  chart.value.accounts = [
    ...(chart.value.accounts || []),
    { code, name, category: draft.category, active: true },
  ]
  draft.code = ''
  draft.name = ''
  draft.category = 'asset'
  notify.success('已添加科目（请保存）')
}

function removeAccount(index) {
  const next = [...(chart.value.accounts || [])]
  next.splice(index, 1)
  chart.value.accounts = next
}

async function save() {
  const check = validateChartPayload(chart.value)
  if (!check.ok) {
    notify.error(check.message)
    return
  }
  await saveChart(ledgerId)
}

async function reset() {
  await resetChart(ledgerId)
}
</script>

<style scoped>
.detail-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1rem 1.1rem;
  box-shadow: var(--shadow-sm);
}
.detail-card__title {
  margin: 0 0 0.85rem;
  font-size: 0.8125rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.field-hint {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0 0 0.75rem;
}
.coa-add {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.75rem;
  align-items: center;
}
</style>
