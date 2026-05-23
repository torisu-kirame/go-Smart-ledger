<template>
  <div class="form-grid">
    <div v-for="f in schema.fields" :key="f.key" class="form-row">
      <label>{{ f.label }}<span v-if="f.required" class="req">*</span></label>
      <select
        v-if="f.type === 'user'"
        v-model="model[f.key]"
        :required="f.required"
      >
        <option value="" disabled>请选择</option>
        <option v-for="m in members" :key="m.id" :value="m.id">
          {{ m.nickname || m.username || m.id }}
        </option>
      </select>
      <input
        v-else-if="f.type === 'date'"
        v-model="model[f.key]"
        type="date"
        :required="f.required"
      />
      <input
        v-else-if="f.type === 'number'"
        v-model="model[f.key]"
        type="text"
        inputmode="decimal"
        :placeholder="f.label"
        :required="f.required"
      />
      <input
        v-else
        v-model="model[f.key]"
        type="text"
        :placeholder="f.label"
        :required="f.required"
      />
    </div>
  </div>
</template>

<script setup>
defineProps({
  schema: { type: Object, required: true },
  model: { type: Object, required: true },
  members: { type: Array, default: () => [] },
})
</script>

<style scoped>
.form-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 0.75rem; }
.req { color: var(--danger); margin-left: 0.15rem; }
</style>
