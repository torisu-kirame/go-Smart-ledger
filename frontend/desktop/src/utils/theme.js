const THEME_KEY = 'smart-ledger-theme'
const ACCENT_KEY = 'smart-ledger-accent'
export const THEME_CHANGE = 'smart-ledger-theme-change'

/** 界面主题（非系统昼夜跟随） */
export const THEME_PRESETS = [
  { id: 'classic-light', i18nKey: 'classicLight' },
  { id: 'classic-dark', i18nKey: 'deepBlue' },
  { id: 'deep-dark', i18nKey: 'deepDark' },
]

/**
 * 与 global.css 中各 data-theme 的 CSS 变量保持一致（改主题色时请同步两处）。
 */
export const THEME_PALETTE = {
  'classic-light': {
    bg: '#dce3ef',
    bgElevated: '#e8edf5',
    bgCard: '#f2f5fa',
    border: '#b8c5d9',
    text: '#1a2438',
    textMuted: '#5a6b85',
    accent: '#3d7eff',
    accentDim: '#2d66d2',
  },
  'classic-dark': {
    bg: '#0a1424',
    bgElevated: '#0f1c32',
    bgCard: '#152845',
    border: '#243d5e',
    text: '#e4ecf8',
    textMuted: '#8fa3c4',
    accent: '#4f8cff',
    accentDim: '#2e6ddf',
  },
  'deep-dark': {
    bg: '#050508',
    bgElevated: '#0a0a0e',
    bgCard: '#121218',
    border: '#222228',
    text: '#ececf1',
    textMuted: '#8b8b96',
    accent: '#5b8def',
    accentDim: '#3d6fd6',
  },
}

/** @deprecated 使用 THEME_PALETTE */
export const THEME_SWATCHES = THEME_PALETTE

export function getThemePalette(themeId) {
  const id = normalizeTheme(themeId)
  return THEME_PALETTE[id] || THEME_PALETTE['classic-dark']
}

/** 强调色在各主题下的 accent / accent-dim（与 global.css data-accent 一致） */
export const ACCENT_PALETTE = {
  teal: {
    dark: { accent: '#2dd4bf', accentDim: '#14b8a6' },
    light: { accent: '#0d9488', accentDim: '#0f766e' },
  },
  violet: {
    dark: { accent: '#a78bfa', accentDim: '#8b5cf6' },
    light: { accent: '#7c3aed', accentDim: '#6d28d9' },
  },
  amber: {
    dark: { accent: '#f59e0b', accentDim: '#d97706' },
    light: { accent: '#d97706', accentDim: '#b45309' },
  },
  rose: {
    dark: { accent: '#f472b6', accentDim: '#ec4899' },
    light: { accent: '#db2777', accentDim: '#be185d' },
  },
}

const VALID_THEMES = new Set(THEME_PRESETS.map((p) => p.id))

export const ACCENT_PRESETS = [
  { id: 'blue', name: '天蓝', swatch: '#4f8cff' },
  { id: 'teal', name: '青绿', swatch: '#2dd4bf' },
  { id: 'violet', name: '紫罗兰', swatch: '#a78bfa' },
  { id: 'amber', name: '琥珀', swatch: '#f59e0b' },
  { id: 'rose', name: '玫红', swatch: '#f472b6' },
]

const VALID_ACCENTS = new Set(ACCENT_PRESETS.map((p) => p.id))

/** 预览卡上下两条模拟条用的强调色（随所选主题 + 主题色） */
export function getPreviewAccentColors(themeId, accentId) {
  const theme = normalizeTheme(themeId)
  const accent = VALID_ACCENTS.has(accentId) ? accentId : 'blue'
  if (accent === 'blue') {
    const p = getThemePalette(theme)
    return { accent: p.accent, accentDim: p.accentDim }
  }
  const row = ACCENT_PALETTE[accent]
  const variant = theme === 'classic-light' ? row.light : row.dark
  return { accent: variant.accent, accentDim: variant.accentDim }
}

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

function notifyThemeChange(theme) {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent(THEME_CHANGE, { detail: theme }))
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
  notifyThemeChange(t)
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

/** 登录页快捷切换：经典白 → 深蓝 → 深黑 */
export function cycleTheme(currentTheme) {
  const cur = normalizeTheme(currentTheme)
  const idx = THEME_PRESETS.findIndex((p) => p.id === cur)
  const next = THEME_PRESETS[(idx + 1) % THEME_PRESETS.length]
  return setTheme(next.id)
}
