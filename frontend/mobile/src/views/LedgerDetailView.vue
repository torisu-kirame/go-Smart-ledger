<template>
  <div class="page page--sub">
    <van-nav-bar
      :title="ledger?.name || '账本详情'"
      left-arrow
      fixed
      placeholder
      safe-area-inset-top
      @click-left="router.back()"
    />

    <van-pull-refresh v-model="refreshing" @refresh="load">
      <van-skeleton v-if="loading && !ledger" title :row="4" />

      <template v-else-if="ledger">
        <van-cell-group inset title="概览">
          <van-cell title="账本 ID" :value="ledger.id" value-class="mono" />
          <van-cell title="类型" :value="ledger.type === 'multi' ? '多人' : '私人'" />
          <van-cell title="最新序号" :value="String(ledger.latestSeq ?? 0)" />
          <van-cell title="锚定状态">
            <template #value>
              <van-tag :type="ledger.anchorStatus === 'synced' ? 'success' : 'warning'" plain>
                {{ ledger.anchorStatus || 'pending' }}
              </van-tag>
            </template>
          </van-cell>
          <van-cell
            v-if="ledger.ledgerAddress"
            title="链上地址"
            :label="ledger.ledgerAddress"
            label-class="mono"
          />
        </van-cell-group>

        <van-cell-group inset title="记一笔">
          <van-field
            v-for="f in entryFields"
            :key="f.key"
            v-model="entryForm[f.key]"
            :label="f.label || f.key"
            :type="f.type === 'number' ? 'number' : 'text'"
            :placeholder="f.label || f.key"
          />
          <div class="entry-actions">
            <van-button type="primary" block round :loading="entryBusy" @click="submitEntry">
              提交分录
            </van-button>
          </div>
        </van-cell-group>

        <h3 class="section-title">最近事件</h3>
        <van-empty v-if="!events.length" description="暂无事件" />
        <van-cell-group v-else inset>
          <van-cell
            v-for="ev in events"
            :key="ev.seq || ev.id"
            :title="eventTitle(ev)"
            :label="eventLabel(ev)"
          />
        </van-cell-group>
      </template>

      <van-empty v-else description="账本不存在或无权访问" />
    </van-pull-refresh>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showFailToast, showSuccessToast } from 'vant'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'

defineOptions({ name: 'LedgerDetailView' })

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const ledger = ref(null)
const events = ref([])
const loading = ref(false)
const refreshing = ref(false)
const entryBusy = ref(false)
const entryForm = reactive({})

const ledgerId = computed(() => route.params.id)

const entryFields = computed(() => {
  const fields = ledger.value?.entrySchema?.fields
  if (Array.isArray(fields) && fields.length) return fields
  return [
    { key: 'amount', label: '金额', type: 'number' },
    { key: 'note', label: '备注', type: 'text' },
    { key: 'date', label: '日期', type: 'text' },
  ]
})

function eventTitle(ev) {
  const kind = ev.kind || ev.type || 'event'
  return `#${ev.seq ?? '—'} · ${kind}`
}

function eventLabel(ev) {
  if (ev.timestamp) return new Date(ev.timestamp).toLocaleString()
  if (ev.data && typeof ev.data === 'object') {
    const parts = Object.entries(ev.data)
      .slice(0, 3)
      .map(([k, v]) => `${k}: ${v}`)
    return parts.join(' · ') || JSON.stringify(ev.data).slice(0, 80)
  }
  return ''
}

async function load() {
  loading.value = true
  try {
    const [l, evRes] = await Promise.all([
      api.getLedger(ledgerId.value),
      api.listLedgerEvents(ledgerId.value).catch(() => ({ events: [] })),
    ])
    ledger.value = l?.ledger || l
    const ev = evRes?.events || evRes || []
    events.value = (Array.isArray(ev) ? ev : []).slice(-20).reverse()
    for (const f of entryFields.value) {
      if (entryForm[f.key] == null) entryForm[f.key] = ''
    }
  } catch (e) {
    showFailToast(e instanceof ApiError ? e.message : '加载失败')
    ledger.value = null
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

async function submitEntry() {
  if (!ledger.value) return
  entryBusy.value = true
  try {
    const data = {}
    for (const f of entryFields.value) {
      const v = entryForm[f.key]
      if (v === '' || v == null) continue
      data[f.key] = f.type === 'number' ? Number(v) : v
    }
    await api.appendEntry(ledgerId.value, {
      signerId: auth.user?.id,
      schemaId: ledger.value.entrySchema?.templateId || 'default',
      data,
    })
    showSuccessToast('已提交')
    for (const k of Object.keys(entryForm)) entryForm[k] = ''
    await load()
  } catch (e) {
    showFailToast(e instanceof ApiError ? e.message : '提交失败')
  } finally {
    entryBusy.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.entry-actions {
  padding: 12px 16px 4px;
}
</style>
