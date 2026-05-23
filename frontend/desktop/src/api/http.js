const BASE = '/api/v1'

let getToken = () => null
let onRefresh = async () => false

export function configureAuth(tokenGetter, refreshFn) {
  getToken = tokenGetter
  onRefresh = refreshFn
}

export class ApiError extends Error {
  constructor(message, status) {
    super(message)
    this.status = status
  }
}

async function request(path, init = {}, retry = true) {
  const headers = { ...(init.headers || {}) }
  if (!(init.body instanceof FormData)) {
    headers['Content-Type'] = headers['Content-Type'] || 'application/json'
  }
  const token = getToken()
  if (token) headers.Authorization = `Bearer ${token}`

  const res = await fetch(`${BASE}${path}`, {
    credentials: 'include',
    ...init,
    headers,
  })

  if (res.status === 401 && retry && onRefresh && !path.startsWith('/auth/')) {
    const ok = await onRefresh()
    if (ok) return request(path, init, false)
  }

  if (res.status === 204) return null

  const text = await res.text()
  let body = null
  if (text) {
    try {
      body = JSON.parse(text)
    } catch {
      body = { error: text }
    }
  }
  if (!res.ok) {
    const msg = body?.message || body?.error || res.statusText
    throw new ApiError(msg, res.status)
  }
  return body
}

export const api = {
  captcha: () => request('/auth/captcha'),
  login: (data) => request('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  register: (data) => request('/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  getProfile: () => request('/users/me'),
  updateProfile: (nickname) =>
    request('/users/me', { method: 'PATCH', body: JSON.stringify({ nickname }) }),
  deleteAccount: (username, password) =>
    request('/users/me/delete-account', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),
  uploadAvatar: (file) => {
    const fd = new FormData()
    fd.append('avatar', file)
    return request('/users/me/avatar', { method: 'POST', body: fd })
  },
  userAvatarUrl: (userId) => `/api/v1/users/${encodeURIComponent(userId)}/avatar`,
  searchUser: (userId) => request(`/users/search?userId=${encodeURIComponent(userId)}`),
  listFriends: () => request('/friends'),
  addFriend: (friendUserId) =>
    request('/friends', { method: 'POST', body: JSON.stringify({ friendUserId }) }),
  removeFriend: (friendId) => request(`/friends/${encodeURIComponent(friendId)}`, { method: 'DELETE' }),
  listTeams: () => request('/teams'),
  createTeam: (body) => request('/teams', { method: 'POST', body: JSON.stringify(body) }),
  getTeam: (id) => request(`/teams/${encodeURIComponent(id)}`),
  refresh: () => request('/auth/refresh', { method: 'POST', body: '{}' }),
  logout: () => request('/auth/logout', { method: 'POST', body: '{}' }),
  health: () => request('/health'),
  listLedgers: () => request('/ledgers'),
  getLedger: (id) => request(`/ledgers/${id}`),
  createLedger: (data) => request('/ledgers', { method: 'POST', body: JSON.stringify(data) }),
  appendEntry: (id, entry) =>
    request(`/ledgers/${id}/entries`, { method: 'POST', body: JSON.stringify({ entry }) }),
  listEvents: (id, from = 1, to = 0) => {
    const q = new URLSearchParams({ from: String(from) })
    if (to > 0) q.set('to', String(to))
    return request(`/ledgers/${id}/events?${q}`)
  },
  anchor: (id, seqFrom = 0, seqTo = 0) =>
    request(`/ledgers/${id}/anchor`, {
      method: 'POST',
      body: JSON.stringify({ seqFrom, seqTo }),
    }),
  verify: (id) => request(`/ledgers/${id}/verify`),
  downloadTemplate: () =>
    fetch(`${BASE}/import/template`, { credentials: 'include', headers: { Authorization: `Bearer ${getToken() || ''}` } }),
  importPreview: (id, file) => {
    const fd = new FormData()
    fd.append('file', file)
    return request(`/ledgers/${id}/import/preview`, { method: 'POST', body: fd })
  },
  importCommit: (id, body) =>
    request(`/ledgers/${id}/import/commit`, { method: 'POST', body: JSON.stringify(body) }),
  ledgerBackup: (id, password) =>
    request(`/ledgers/${id}/backup`, { method: 'POST', body: JSON.stringify({ password }) }),
  restorePreview: (id, ref, password) =>
    request(`/ledgers/${id}/restore/preview`, {
      method: 'POST',
      body: JSON.stringify({ ref, password }),
    }),
}
