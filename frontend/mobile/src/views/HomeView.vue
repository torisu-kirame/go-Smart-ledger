<template>
  <div class="page">
    <van-nav-bar title="Smart Ledger" fixed placeholder safe-area-inset-top />

    <van-pull-refresh v-model="refreshing" @refresh="load">
      <div class="stat-grid">
        <div class="stat-card">
          <h4>链状态</h4>
          <div class="val" :class="chainOnline ? 'ok' : 'bad'">{{ chainOnline ? '在线' : '离线' }}</div>
        </div>
        <div class="stat-card">
          <h4>区块高度</h4>
          <div class="val mono">{{ chainHeight }}</div>
        </div>
        <div class="stat-card">
          <h4>我的账本</h4>
          <div class="val">{{ ledgers.length }}</div>
        </div>
        <div class="stat-card">
          <h4>好友</h4>
          <div class="val">{{ friendCount }}</div>
        </div>
      </div>

      <h3 class="section-title">快捷入口</h3>
      <van-grid :column-num="4" :border="false" clickable>
        <van-grid-item icon="balance-list-o" text="账本" to="/ledgers" />
        <van-grid-item icon="plus" text="新建" @click="goCreateLedger" />
        <van-grid-item icon="friends-o" text="协作" to="/collab" />
        <van-grid-item icon="setting-o" text="设置" to="/profile" />
      </van-grid>

      <h3 class="section-title">使用提示</h3>
      <van-cell-group inset>
        <van-cell title="简单流水" label="自定义模板字段，支持 Excel 导入" />
        <van-cell title="多人协作" label="邀请成员、提议与审批分录" />
        <van-cell title="链上锚定" label="封账后可在链浏览器查验" />
      </van-cell-group>
    </van-pull-refresh>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showFailToast } from 'vant'
import { api, ApiError } from '../api/http'

defineOptions({ name: 'HomeView' })

const router = useRouter()
const refreshing = ref(false)
const chainOnline = ref(false)
const chainHeight = ref('—')
const ledgers = ref([])
const friendCount = ref(0)

async function load() {
  refreshing.value = true
  try {
    const [chain, ledgerList, friends] = await Promise.all([
      api.chainStatus().catch(() => null),
      api.listLedgers().catch(() => ({ ledgers: [] })),
      api.listFriends().catch(() => ({ friends: [] })),
    ])
    chainOnline.value = !!chain?.online
    chainHeight.value = chain?.height != null ? String(chain.height) : '—'
    ledgers.value = ledgerList?.ledgers || ledgerList || []
    if (!Array.isArray(ledgers.value)) ledgers.value = []
    const fl = friends?.friends || friends || []
    friendCount.value = Array.isArray(fl) ? fl.length : 0
  } catch (e) {
    if (e instanceof ApiError) showFailToast(e.message)
  } finally {
    refreshing.value = false
  }
}

function goCreateLedger() {
  router.push({ path: '/ledgers', query: { create: '1' } })
}

onMounted(load)
</script>
