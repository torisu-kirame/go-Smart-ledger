<template>
  <div class="page">
    <van-nav-bar title="账本" fixed placeholder safe-area-inset-top>
      <template #right>
        <van-icon name="plus" size="20" @click="showCreate = true" />
      </template>
    </van-nav-bar>

    <van-pull-refresh v-model="refreshing" @refresh="load">
      <van-notice-bar
        v-if="incomingInvites.length"
        left-icon="info-o"
        :text="`您有 ${incomingInvites.length} 条账本邀请`"
        color="#1a56db"
        background="#e8effc"
      />

      <van-cell-group v-if="incomingInvites.length" inset title="收到的邀请">
        <van-cell
          v-for="inv in incomingInvites"
          :key="inv.ledgerId + inv.inviterId"
          :title="inviteName(inv)"
          :label="`邀请人 ${inv.inviterId || '—'}`"
        >
          <template #value>
            <van-button size="small" type="primary" :loading="inviteBusy" @click="acceptInvite(inv)">
              接受
            </van-button>
          </template>
        </van-cell>
      </van-cell-group>

      <van-empty v-if="!list.length && !loading" description="暂无账本，点击右上角创建" />

      <van-cell-group v-else inset title="我的账本">
        <van-cell
          v-for="l in list"
          :key="l.id"
          is-link
          :title="l.name"
          :label="ledgerLabel(l)"
          @click="goDetail(l.id)"
        >
          <template #value>
            <van-tag :type="l.anchorStatus === 'synced' ? 'success' : 'warning'" plain>
              {{ l.anchorStatus || 'pending' }}
            </van-tag>
          </template>
        </van-cell>
      </van-cell-group>
    </van-pull-refresh>

    <van-dialog
      v-model:show="showCreate"
      title="创建账本"
      show-cancel-button
      :before-close="beforeCreateClose"
    >
      <van-cell-group inset>
        <van-field v-model="form.name" label="名称" placeholder="账本名称" required />
        <van-field label="类型">
          <template #input>
            <van-radio-group v-model="form.type" direction="horizontal">
              <van-radio name="private">私人</van-radio>
              <van-radio name="multi">多人</van-radio>
            </van-radio-group>
          </template>
        </van-field>
      </van-cell-group>
    </van-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showFailToast, showSuccessToast } from 'vant'
import { api, ApiError } from '../api/http'
import { useAuthStore } from '../stores/auth'

defineOptions({ name: 'LedgersView' })

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const list = ref([])
const incomingInvites = ref([])
const loading = ref(false)
const refreshing = ref(false)
const showCreate = ref(false)
const inviteBusy = ref(false)
const form = reactive({
  name: '',
  type: 'private',
})

function ledgerLabel(l) {
  const type = l.type === 'multi' ? '多人' : '私人'
  return `${type} · 流水 · 序号 ${l.latestSeq ?? 0}`
}

function inviteName(inv) {
  const found = list.value.find((x) => x.id === inv.ledgerId)
  return found?.name || `账本 ${inv.ledgerId}`
}

function goDetail(id) {
  router.push(`/ledgers/${id}`)
}

async function load() {
  loading.value = true
  try {
    const [ledgersRes, invitesRes] = await Promise.all([
      api.listLedgers(),
      api.listMyInvites().catch(() => ({ invites: [] })),
    ])
    list.value = ledgersRes?.ledgers || ledgersRes || []
    if (!Array.isArray(list.value)) list.value = []
    incomingInvites.value = invitesRes?.invites || invitesRes || []
    if (!Array.isArray(incomingInvites.value)) incomingInvites.value = []
  } catch (e) {
    showFailToast(e instanceof ApiError ? e.message : '加载失败')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

async function acceptInvite(inv) {
  inviteBusy.value = true
  try {
    await api.acceptLedgerInvite(inv.ledgerId)
    showSuccessToast('已加入账本')
    await load()
  } catch (e) {
    showFailToast(e instanceof ApiError ? e.message : '接受失败')
  } finally {
    inviteBusy.value = false
  }
}

async function beforeCreateClose(action) {
  if (action !== 'confirm') return true
  if (!form.name.trim()) {
    showFailToast('请输入账本名称')
    return false
  }
  try {
    await api.createLedger({
      name: form.name.trim(),
      type: form.type,
      bookkeepingMode: 'simple',
      entrySchema: { templateId: 'custom', fields: [] },
      multiTableEnabled: true,
      creatorId: auth.user?.id,
    })
    showSuccessToast('创建成功')
    form.name = ''
    await load()
    return true
  } catch (e) {
    showFailToast(e instanceof ApiError ? e.message : '创建失败')
    return false
  }
}

watch(
  () => route.query.create,
  (v) => {
    if (v) showCreate.value = true
  },
  { immediate: true }
)

onMounted(load)
</script>
