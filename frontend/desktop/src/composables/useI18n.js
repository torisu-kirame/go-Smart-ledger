import { computed, ref } from 'vue'
import { translate } from '../i18n/messages'
import { getLocale, LOCALE_CHANGE, setLocale as persistLocale } from '../utils/locale'

const locale = ref(getLocale())

if (typeof window !== 'undefined') {
  window.addEventListener(LOCALE_CHANGE, (e) => {
    locale.value = e.detail
  })
}

export function useI18n() {
  const t = (key) => translate(locale.value, key)

  const localeOptions = computed(() => [
    { value: 'zh', label: t('settings.theme.langZh') },
    { value: 'en', label: t('settings.theme.langEn') },
  ])

  function setLocale(next) {
    locale.value = persistLocale(next)
  }

  return { locale, t, setLocale, localeOptions }
}
