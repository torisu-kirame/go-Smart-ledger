export const SETTINGS_SECTIONS = ['appearance', 'language', 'ai', 'account', 'security']

const LEGACY_HASH = {
  personal: 'account',
  theme: 'appearance',
}

export function normalizeSettingsHash(hash) {
  const raw = (hash || '').replace(/^#/, '')
  if (!raw) return 'appearance'
  const id = LEGACY_HASH[raw] || raw
  return SETTINGS_SECTIONS.includes(id) ? id : 'appearance'
}
