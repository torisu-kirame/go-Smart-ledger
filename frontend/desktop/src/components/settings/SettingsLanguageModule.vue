<template>
  <div class="settings-module">
    <div class="form-row">
      <label>{{ t('settings.theme.language') }}</label>
      <AppSelect
        v-model="currentLocale"
        :options="localeOptions"
        @change="onLocaleChange"
      />
    </div>
    <p class="hint">{{ t('settings.language.hint') }}</p>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import AppSelect from '../AppSelect.vue'
import { useI18n } from '../../composables/useI18n'

const { locale, t, setLocale, localeOptions } = useI18n()
const currentLocale = ref(locale.value)

function onLocaleChange(val) {
  setLocale(val)
  currentLocale.value = val
}

watch(locale, (val) => {
  currentLocale.value = val
})
</script>

<style scoped>
.hint {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-muted);
  max-width: var(--field-max);
}
</style>
