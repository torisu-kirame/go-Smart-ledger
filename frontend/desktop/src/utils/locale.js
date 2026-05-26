const LOCALE_KEY = 'smart-ledger-locale'
export const LOCALE_CHANGE = 'smart-ledger-locale-change'

export const LOCALES = {
  zh: 'zh',
  en: 'en',
}

export function getLocale() {
  if (typeof window === 'undefined') return LOCALES.zh
  const saved = window.localStorage.getItem(LOCALE_KEY)
  if (saved === LOCALES.zh || saved === LOCALES.en) return saved
  const lang = navigator.language || ''
  return lang.toLowerCase().startsWith('zh') ? LOCALES.zh : LOCALES.en
}

export function applyLocale(locale) {
  if (typeof document === 'undefined') return
  document.documentElement.lang = locale === LOCALES.en ? 'en' : 'zh-CN'
}

export function initLocale() {
  const locale = getLocale()
  applyLocale(locale)
  return locale
}

export function setLocale(locale) {
  const next = locale === LOCALES.en ? LOCALES.en : LOCALES.zh
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(LOCALE_KEY, next)
    window.dispatchEvent(new CustomEvent(LOCALE_CHANGE, { detail: next }))
  }
  applyLocale(next)
  return next
}

export function dashboardPath(locale, themeId = 'classic-dark') {
  const base = locale === LOCALES.en ? '/dashboard/' : '/explorer-zh/'
  const theme = encodeURIComponent(themeId || 'classic-dark')
  const join = base.includes('?') ? '&' : '?'
  return `${base}${join}theme=${theme}`
}
