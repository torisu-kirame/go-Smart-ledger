<template>
  <div class="page">
    <van-nav-bar title="协作" fixed placeholder safe-area-inset-top />

    <van-tabs v-model:active="tab" sticky offset-top="46">
      <van-tab title="好友" name="friends">
        <van-pull-refresh v-model="refreshFriends" @refresh="loadFriends">
          <van-cell-group v-if="incoming.length" inset title="收到的请求">
            <van-cell
              v-for="r in incoming"
              :key="r.fromUserId"
              :title="r.fromUserId"
              label="请求加为好友"
            >
              <template #value>
                <van-button size="mini" type="primary" @click="accept(r.fromUserId)">同意</van-button>
                <van-button size="mini" style="margin-left: 6px" @click="reject(r.fromUserId)">拒绝</van-button>
              </template>
            </van-cell>
          </van-cell-group>

          <van-empty v-if="!friends.length && !incoming.length" description="暂无好友" />

          <van-cell-group v-if="friends.length" inset title="好友列表">
            <van-cell v-for="f in friends" :key="f.friendId || f.id" :title="displayFriend(f)">
              <template #right-icon>
                <van-icon name="delete-o" @click="removeFriend(f)" />
              </template>
            </van-cell>
          </van-cell-group>

          <div class="fab-area">
            <van-button round type="primary" icon="plus" @click="showAdd = true">添加好友</van-button>
          </div>
        </van-pull-refresh>
      </van-tab>

      <van-tab title="团队" name="teams">
        <van-pull-refresh v-model="refreshTeams" @refresh="loadTeams">
          <van-empty v-if="!teams.length" description="暂无团队" />
          <van-cell-group v-else inset>
            <van-cell
              v-for="t in teams"
              :key="t.id"
              :title="t.name"
              :label="`成员 ${t.memberCount ?? '—'}`"
            />
          </van-cell-group>
        </van-pull-refresh>
      </van-tab>
    </van-tabs>

    <van-dialog v-model:show="showAdd" title="添加好友" show-cancel-button @confirm="addFriend">
      <van-field v-model="addUserId" label="用户 ID" placeholder="对方用户 ID" />
    </van-dialog>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { showConfirmDialog, showFailToast, showSuccessToast } from 'vant'
import { api, ApiError } from '../api/http'

defineOptions({ name: 'CollabView' })

const tab = ref('friends')
const friends = ref([])
const incoming = ref([])
const teams = ref([])
const refreshFriends = ref(false)
const refreshTeams = ref(false)
const showAdd = ref(false)
const addUserId = ref('')

function displayFriend(f) {
  return f.nickname || f.username || f.friendId || f.id
}

async function loadFriends() {
  try {
    const [fl, inc] = await Promise.all([
      api.listFriends(),
      api.listIncomingFriendRequests(),
    ])
    friends.value = fl?.friends || fl || []
    incoming.value = inc?.requests || inc || []
    if (!Array.isArray(friends.value)) friends.value = []
    if (!Array.isArray(incoming.value)) incoming.value = []
  } catch (e) {
    showFailToast(e instanceof ApiError ? e.message : '加载失败')
  } finally {
    refreshFriends.value = false
  }
}

async function loadTeams() {
  try {
    const res = await api.listTeams()
    teams.value = res?.teams || res || []
    if (!Array.isArray(teams.value)) teams.value = []
  } catch (e) {
    showFailToast(e instanceof ApiError ? e.message : '加载失败')
  } finally {
    refreshTeams.value = false
  }
}

async function accept(fromUserId) {
  try {
    await api.acceptFriendRequest(fromUserId)
    showSuccessToast('已添加')
    loadFriends()
  } catch (e) {
    showFailToast(e.message)
  }
}

async function reject(fromUserId) {
  try {
    await api.rejectFriendRequest(fromUserId)
    loadFriends()
  } catch (e) {
    showFailToast(e.message)
  }
}

async function addFriend() {
  const id = addUserId.value.trim()
  if (!id) {
    showFailToast('请输入用户 ID')
    return
  }
  try {
    await api.addFriend(id)
    showSuccessToast('请求已发送')
    addUserId.value = ''
    loadFriends()
  } catch (e) {
    showFailToast(e instanceof ApiError ? e.message : '发送失败')
  }
}

async function removeFriend(f) {
  try {
    await showConfirmDialog({ title: '删除好友', message: '确定删除该好友？' })
    await api.removeFriend(f.friendId || f.id)
    showSuccessToast('已删除')
    loadFriends()
  } catch (e) {
    if (e !== 'cancel') showFailToast(e.message || '操作失败')
  }
}

onMounted(() => {
  loadFriends()
  loadTeams()
})
</script>

<style scoped>
.fab-area {
  display: flex;
  justify-content: center;
  padding: 24px 16px;
}
</style>
