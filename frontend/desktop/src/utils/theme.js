const THEME_KEY = 'smart-ledger-theme'
const ACCENT_KEY = 'smart-ledger-accent'

export const ACCENT_PRESETS = [
  { id: 'blue', name: '天蓝', swatch: '#4f8cff' },
  { id: 'teal', name: '青绿', swatch: '#2dd4bf' },
  { id: 'violet', name: '紫罗兰', swatch: '#a78bfa' },
  { id: 'amber', name: '琥珀', swatch: '#f59e0b' },
  { id: 'rose', name: '玫红', swatch: '#f472b6' },
]

const VALID_ACCENTS = new Set(ACCENT_PRESETS.map((p) => p.id))

function preferredTheme() {
  if (typeof window === 'undefined') return 'dark'
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark'
}

export function getTheme() {
  if (typeof window === 'undefined') return 'dark'
  const saved = window.localStorage.getItem(THEME_KEY)
  if (saved === 'light' || saved === 'dark') return saved
  return preferredTheme()
}

export function getAccent() {
  if (typeof window === 'undefined') return 'blue'
  const saved = window.localStorage.getItem(ACCENT_KEY)
  return VALID_ACCENTS.has(saved) ? saved : 'blue'
}

export function applyTheme(theme) {
  if (typeof document === 'undefined') return
  const t = theme === 'light' ? 'light' : 'dark'
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
  const t = theme === 'light' ? 'light' : 'dark'
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

export function toggleTheme(currentTheme) {
  return setTheme(currentTheme === 'light' ? 'dark' : 'light')
}
