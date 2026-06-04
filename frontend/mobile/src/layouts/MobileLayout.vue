<template>
  <div class="mobile-shell">
    <router-view v-slot="{ Component, route: r }">
      <keep-alive :include="['HomeView', 'LedgersView', 'CollabView', 'ProfileView']">
        <component :is="Component" :key="r.name" />
      </keep-alive>
    </router-view>

    <van-tabbar
      v-if="showTabbar"
      v-model="activeTab"
      route
      safe-area-inset-bottom
      active-color="#1a56db"
      inactive-color="#6b7280"
    >
      <van-tabbar-item replace to="/" icon="home-o" name="home">首页</van-tabbar-item>
      <van-tabbar-item replace to="/ledgers" icon="balance-list-o" name="ledgers">账本</van-tabbar-item>
      <van-tabbar-item replace to="/collab" icon="friends-o" name="collab">协作</van-tabbar-item>
      <van-tabbar-item replace to="/profile" icon="user-o" name="profile">我的</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const showTabbar = computed(() => !route.meta.hideTabbar)
const activeTab = computed(() => route.meta.tab || '')
</script>

<style scoped>
.mobile-shell {
  min-height: 100%;
  background: var(--sl-bg);
}
</style>
