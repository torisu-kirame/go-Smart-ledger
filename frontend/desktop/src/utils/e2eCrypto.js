/**
 * F19: client-side group key wrap/unwrap (AES-GCM via Web Crypto).
 * Group key never sent to server in plaintext.
 */

const STORAGE_PREFIX = 'sl-group-key:'
const PASSPHRASE_PREFIX = 'sl-e2e-passphrase:'

function enc() {
  return new TextEncoder()
}

function toB64(buf) {
  const bytes = new Uint8Array(buf)
  let s = ''
  bytes.forEach((b) => { s += String.fromCharCode(b) })
  return btoa(s)
}

function fromB64(s) {
  const bin = atob(s)
  const bytes = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i)
  return bytes
}

export function generateGroupKey() {
  const raw = crypto.getRandomValues(new Uint8Array(32))
  return toB64(raw)
}

async function deriveWrapKey(passphrase, ledgerId, userId) {
  const material = enc().encode(`${passphrase}|${ledgerId}|${userId}`)
  const base = await crypto.subtle.importKey('raw', material, 'PBKDF2', false, ['deriveKey'])
  return crypto.subtle.deriveKey(
    { name: 'PBKDF2', salt: enc().encode('smart-ledger-e2e'), iterations: 120000, hash: 'SHA-256' },
    base,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt']
  )
}

export async function wrapGroupKey(groupKeyB64, passphrase, ledgerId, userId) {
  const key = await deriveWrapKey(passphrase, ledgerId, userId)
  const iv = crypto.getRandomValues(new Uint8Array(12))
  const ct = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    key,
    fromB64(groupKeyB64)
  )
  const packed = new Uint8Array(iv.length + ct.byteLength)
  packed.set(iv, 0)
  packed.set(new Uint8Array(ct), iv.length)
  return toB64(packed.buffer)
}

export async function unwrapGroupKey(wrappedB64, passphrase, ledgerId, userId) {
  const key = await deriveWrapKey(passphrase, ledgerId, userId)
  const packed = fromB64(wrappedB64)
  const iv = packed.slice(0, 12)
  const data = packed.slice(12)
  const raw = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, data)
  return toB64(raw)
}

export async function encryptEntryData(groupKeyB64, dataObj) {
  const key = await crypto.subtle.importKey(
    'raw',
    fromB64(groupKeyB64),
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt']
  )
  const iv = crypto.getRandomValues(new Uint8Array(12))
  const plain = enc().encode(JSON.stringify(dataObj))
  const ct = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, plain)
  const packed = new Uint8Array(iv.length + ct.byteLength)
  packed.set(iv, 0)
  packed.set(new Uint8Array(ct), iv.length)
  return { __encrypted: true, payload: toB64(packed.buffer) }
}

export async function decryptEntryData(groupKeyB64, dataObj) {
  if (!dataObj?.__encrypted || !dataObj.payload) return dataObj
  const key = await crypto.subtle.importKey(
    'raw',
    fromB64(groupKeyB64),
    { name: 'AES-GCM', length: 256 },
    false,
    ['decrypt']
  )
  const packed = fromB64(dataObj.payload)
  const iv = packed.slice(0, 12)
  const data = packed.slice(12)
  const plain = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, data)
  return JSON.parse(new TextDecoder().decode(plain))
}

export function saveLocalGroupKey(ledgerId, groupKeyB64) {
  localStorage.setItem(STORAGE_PREFIX + ledgerId, groupKeyB64)
}

export function loadLocalGroupKey(ledgerId) {
  return localStorage.getItem(STORAGE_PREFIX + ledgerId) || ''
}

export function saveLocalPassphrase(ledgerId, passphrase) {
  if (passphrase) {
    localStorage.setItem(PASSPHRASE_PREFIX + ledgerId, passphrase)
  }
}

export function loadLocalPassphrase(ledgerId) {
  return localStorage.getItem(PASSPHRASE_PREFIX + ledgerId) || ''
}

async function deriveLoginViewKey(loginPassword, userId) {
  const material = enc().encode(`${loginPassword}|${userId}`)
  const base = await crypto.subtle.importKey('raw', material, 'PBKDF2', false, ['deriveKey'])
  return crypto.subtle.deriveKey(
    {
      name: 'PBKDF2',
      salt: enc().encode('smart-ledger-passphrase-view'),
      iterations: 120000,
      hash: 'SHA-256',
    },
    base,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt']
  )
}

/** Wrap ledger passphrase with login-password-derived key (for server-side member retrieval). */
export async function wrapPassphraseForLoginView(passphrase, loginPassword, userId) {
  const key = await deriveLoginViewKey(loginPassword, userId)
  const iv = crypto.getRandomValues(new Uint8Array(12))
  const ct = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, enc().encode(passphrase))
  const packed = new Uint8Array(iv.length + ct.byteLength)
  packed.set(iv, 0)
  packed.set(new Uint8Array(ct), iv.length)
  return toB64(packed.buffer)
}

export async function unwrapPassphraseForLoginView(wrappedB64, loginPassword, userId) {
  const key = await deriveLoginViewKey(loginPassword, userId)
  const packed = fromB64(wrappedB64)
  const iv = packed.slice(0, 12)
  const data = packed.slice(12)
  const plain = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, data)
  return new TextDecoder().decode(plain)
}

export async function buildEncryptionForCreate(members, creatorId, passphrase, ledgerIdPlaceholder = 'new') {
  const groupKey = generateGroupKey()
  const wrappedKeys = {}
  for (const m of members) {
    const uid = m.id || m
    wrappedKeys[uid] = await wrapGroupKey(groupKey, passphrase, ledgerIdPlaceholder, uid)
  }
  if (!wrappedKeys[creatorId]) {
    wrappedKeys[creatorId] = await wrapGroupKey(groupKey, passphrase, ledgerIdPlaceholder, creatorId)
  }
  return {
    enabled: true,
    algo: 'aes-gcm-v1',
    wrappedKeys,
    _groupKey: groupKey,
  }
}
