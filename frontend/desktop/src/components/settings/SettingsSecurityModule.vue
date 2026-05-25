<template>
  <div class="settings-module danger-zone">
    <p class="warn-text">{{ t('settings.security.warn') }}</p>
    <div class="form-row">
      <label>{{ t('settings.personal.username') }}</label>
      <input
        :value="deleteForm.username"
        autocomplete="username"
        @input="patchDelete('username', $event.target.value)"
      />
    </div>
    <div class="form-row">
      <label>{{ t('settings.personal.password') }}</label>
      <input
        :value="deleteForm.password"
        type="password"
        autocomplete="current-password"
        @input="patchDelete('password', $event.target.value)"
      />
    </div>
    <DeleteButton
      :disabled="deleting"
      :label="deleting ? t('settings.personal.deleting') : t('settings.personal.deleteBtn')"
      @click="$emit('delete-account')"
    />
  </div>
</template>

<script setup>
import DeleteButton from '../DeleteButton.vue'
import { useI18n } from '../../composables/useI18n'

const props = defineProps({
  deleteForm: { type: Object, required: true },
  deleting: { type: Boolean, default: false },
})

const emit = defineEmits(['update:deleteForm', 'delete-account'])

const { t } = useI18n()

function patchDelete(field, value) {
  emit('update:deleteForm', { ...props.deleteForm, [field]: value })
}
</script>

<style scoped>
.danger-zone {
  border: 1px solid color-mix(in srgb, var(--danger) 35%, transparent);
  border-radius: var(--radius);
  padding: 1rem 1.1rem;
  background: color-mix(in srgb, var(--danger) 6%, var(--bg-card));
}

.warn-text {
  margin: 0 0 1rem;
  font-size: 0.875rem;
  color: var(--text-muted);
  line-height: 1.5;
}

.danger-zone .btn-delete {
  margin-top: 0.5rem;
}
</style>
