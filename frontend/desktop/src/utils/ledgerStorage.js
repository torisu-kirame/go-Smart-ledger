export const STORAGE_LOCATIONS = [
  { id: 'ipfs', label: 'IPFS（默认）', hint: '分布式存储，推荐' },
  { id: 'cloud', label: '云端', hint: '官方服务器 MiniLedger' },
  { id: 'local', label: '本地', hint: '服务端本地磁盘备份' },
]

export const DEFAULT_STORAGE_LOCATION = 'ipfs'

export function normalizeStorageLocation(value) {
  const id = String(value || '').trim()
  return STORAGE_LOCATIONS.some((o) => o.id === id) ? id : DEFAULT_STORAGE_LOCATION
}

export function storageLocationLabel(id) {
  return STORAGE_LOCATIONS.find((o) => o.id === normalizeStorageLocation(id))?.label || id
}
