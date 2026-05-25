const THEME_KEY = 'smart-ledger-theme'
const ACCENT_KEY = 'smart-ledger-accent'

/** 界面主题（非系统昼夜跟随） */
export const THEME_PRESETS = [
  { id: 'classic-light', i18nKey: 'classicLight' },
  { id: 'classic-dark', i18nKey: 'classicDark' },
  { id: 'deep-dark', i18nKey: 'deepDark' },
]

const VALID_THEMES = new Set(THEME_PRESETS.map((p) => p.id))

export const ACCENT_PRESETS = [
  { id: 'blue', name: '天蓝', swatch: '#4f8cff' },
  { id: 'teal', name: '青绿', swatch: '#2dd4bf' },
  { id: 'violet', name: '紫罗兰', swatch: '#a78bfa' },
  { id: 'amber', name: '琥珀', swatch: '#f59e0b' },
  { id: 'rose', name: '玫红', swatch: '#f472b6' },
]

const VALID_ACCENTS = new Set(ACCENT_PRESETS.map((p) => p.id))

function normalizeTheme(saved) {
  if (saved === 'light') return 'classic-light'
  if (saved === 'dark') return 'classic-dark'
  if (VALID_THEMES.has(saved)) return saved
  return 'classic-dark'
}

export function isLightTheme(themeId) {
  return normalizeTheme(themeId) === 'classic-light'
}

export function getTheme() {
  if (typeof window === 'undefined') return 'classic-dark'
  return normalizeTheme(window.localStorage.getItem(THEME_KEY))
}

export function getAccent() {
  if (typeof window === 'undefined') return 'blue'
  const saved = window.localStorage.getItem(ACCENT_KEY)
  return VALID_ACCENTS.has(saved) ? saved : 'blue'
}

export function applyTheme(theme) {
  if (typeof document === 'undefined') return
  const t = normalizeTheme(theme)
  document.documentElement.setAttribute('data-theme', t)
}

export function applyAccent(accent) {
  if (typeof document === 'undefined') return
  const id = VALID_ACCENTS.has(accent) ? accent : 'blue'
  document.documentElement.setAttribute('data-accent', id)
}

export function initTheme() {
  const theme = getTheme()
  const accent = getAccent()
  applyTheme(theme)
  applyAccent(accent)
  return { theme, accent }
}

export function setTheme(theme) {
  const t = normalizeTheme(theme)
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(THEME_KEY, t)
  }
  applyTheme(t)
  return t
}

export function setAccent(accent) {
  const id = VALID_ACCENTS.has(accent) ? accent : 'blue'
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(ACCENT_KEY, id)
  }
  applyAccent(id)
  return id
}

/** 登录页快捷切换：经典白 → 经典黑 → 深黑 */
export function cycleTheme(currentTheme) {
  const cur = normalizeTheme(currentTheme)
  const idx = THEME_PRESETS.findIndex((p) => p.id === cur)
  const next = THEME_PRESETS[(idx + 1) % THEME_PRESETS.length]
  return setTheme(next.id)
}
