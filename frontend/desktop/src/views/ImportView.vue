<template>
  <div class="page">
    <PageHeader :crumbs="crumbs" />
    <section class="panel redirect-panel">
      <h3>导入已移至账本内</h3>
      <p class="muted">
        请在「账本管理」打开目标账本，使用顶部 <strong>导入</strong> 页上传 Excel 或 CSV。
        系统会按文件列自动建表并开启多表模式。
      </p>
      <router-link to="/ledgers" class="btn-primary">前往账本管理</router-link>
    </section>
  </div>
</template>

<script setup>
import PageHeader from '../components/PageHeader.vue'
import { usePageCrumbs } from '../composables/usePageCrumbs'

const { crumbs } = usePageCrumbs()
</script>

<style scoped>
.redirect-panel {
  padding: 1.25rem 1.35rem;
  max-width: 36rem;
}
.redirect-panel h3 {
  margin: 0 0 0.5rem;
}
.redirect-panel .muted {
  margin: 0 0 1rem;
  line-height: 1.5;
}
</style>
