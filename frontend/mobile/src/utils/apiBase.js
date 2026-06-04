import { Capacitor } from '@capacitor/core'
import { Preferences } from '@capacitor/preferences'

const STORAGE_KEY = 'smart-ledger-api-base'

let runtimeBase = (import.meta.env.VITE_API_BASE || '').trim()

export function getApiBase() {
  if (runtimeBase) {
    return runtimeBase.replace(/\/+$/, '')
  }
  return '/api/v1'
}

export function getApiOrigin() {
  const base = getApiBase()
  if (base.startsWith('http')) {
    return base.replace(/\/api\/v1\/?$/, '')
  }
  return ''
}

export async function loadApiBaseFromStorage() {
  try {
    const { value } = await Preferences.get({ key: STORAGE_KEY })
    if (value?.trim()) runtimeBase = value.trim()
  } catch {
    const local = localStorage.getItem(STORAGE_KEY)
    if (local?.trim()) runtimeBase = local.trim()
  }
}

export async function saveApiBase(url) {
  const trimmed = (url || '').trim()
  runtimeBase = trimmed
  try {
    if (trimmed) {
      await Preferences.set({ key: STORAGE_KEY, value: trimmed })
      localStorage.setItem(STORAGE_KEY, trimmed)
    } else {
      await Preferences.remove({ key: STORAGE_KEY })
      localStorage.removeItem(STORAGE_KEY)
    }
  } catch {
    if (trimmed) localStorage.setItem(STORAGE_KEY, trimmed)
    else localStorage.removeItem(STORAGE_KEY)
  }
}

export function isNativeApp() {
  return Capacitor.isNativePlatform()
}

export function defaultServerHint() {
  if (Capacitor.getPlatform() === 'android') {
    return 'http://10.0.2.2:25175/api/v1'
  }
  return 'http://127.0.0.1:25175/api/v1'
}

/** 拼接绝对 URL（头像、文件等） */
export function resolveAssetUrl(path) {
  if (!path) return ''
  if (path.startsWith('http') || path.startsWith('data:') || path.startsWith('blob:')) return path
  const origin = getApiOrigin()
  if (origin) return `${origin}${path.startsWith('/') ? path : `/${path}`}`
  return path.startsWith('/') ? path : `/${path}`
}
