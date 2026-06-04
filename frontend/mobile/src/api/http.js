import { getApiBase, resolveAssetUrl } from '../utils/apiBase'

let getToken = () => null
let onRefresh = async () => false

export function configureAuth(tokenGetter, refreshFn) {
  getToken = tokenGetter
  onRefresh = refreshFn
}

export function authHeaders() {
  const headers = {}
  const token = getToken()
  if (token) headers.Authorization = `Bearer ${token}`
  return headers
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

  const res = await fetch(`${getApiBase()}${path}`, {
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
    const msg = body?.msg || body?.message || body?.error || res.statusText
    throw new ApiError(msg, res.status)
  }
  return body
}

export const api = {
  captcha: () => request('/auth/captcha'),
  login: (data) => request('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  register: (data) => request('/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  refresh: () => request('/auth/refresh', { method: 'POST', body: '{}' }),
  logout: () => request('/auth/logout', { method: 'POST', body: '{}' }),
  getProfile: () => request('/users/me'),
  updateProfile: (nickname) =>
    request('/users/me', { method: 'PATCH', body: JSON.stringify({ nickname }) }),
  userAvatarUrl: (userId) => resolveAssetUrl(`/api/v1/users/${encodeURIComponent(userId)}/avatar`),
  searchUser: (userId) => request(`/users/search?userId=${encodeURIComponent(userId)}`),
  listFriends: () => request('/friends'),
  addFriend: (friendUserId) =>
    request('/friends', { method: 'POST', body: JSON.stringify({ friendUserId }) }),
  removeFriend: (friendId) => request(`/friends/${encodeURIComponent(friendId)}`, { method: 'DELETE' }),
  listIncomingFriendRequests: () => request('/friends/requests/incoming'),
  listOutgoingFriendRequests: () => request('/friends/requests/outgoing'),
  acceptFriendRequest: (fromUserId) =>
    request(`/friends/requests/${encodeURIComponent(fromUserId)}/accept`, {
      method: 'POST',
      body: '{}',
    }),
  rejectFriendRequest: (fromUserId) =>
    request(`/friends/requests/${encodeURIComponent(fromUserId)}/reject`, {
      method: 'POST',
      body: '{}',
    }),
  cancelFriendRequest: (toUserId) =>
    request(`/friends/requests/${encodeURIComponent(toUserId)}`, { method: 'DELETE' }),
  listTeams: () => request('/teams'),
  getTeam: (id) => request(`/teams/${encodeURIComponent(id)}`),
  listEntryTemplates: () => request('/entry-templates'),
  listEntrySchemaTemplates: () => request('/entry-schema/templates'),
  listLedgers: () => request('/ledgers'),
  getLedger: (id) => request(`/ledgers/${id}`),
  createLedger: (data) => {
    const mode = data.bookkeepingMode || 'simple'
    const body = { ...data, bookkeepingMode: mode }
    if (mode === 'professional') {
      body.entrySchema = { templateId: 'professional', fields: [] }
      body.approvalPolicy = { enabled: false, threshold: 1 }
    } else {
      body.entrySchema = data.entrySchema || { templateId: 'default' }
    }
    return request('/ledgers', { method: 'POST', body: JSON.stringify(body) })
  },
  listLedgerEvents: (id, sinceSeq = 0) => {
    const q = sinceSeq ? `?sinceSeq=${sinceSeq}` : ''
    return request(`/ledgers/${id}/events${q}`)
  },
  appendEntry: (id, entry) =>
    request(`/ledgers/${id}/entries`, {
      method: 'POST',
      body: JSON.stringify({
        entry: {
          signerId: entry.signerId,
          tableId: entry.tableId,
          schemaId: entry.schemaId,
          data: entry.data,
        },
      }),
    }),
  listMyInvites: () => request('/ledgers/invites/mine'),
  acceptLedgerInvite: (id) =>
    request(`/ledgers/${id}/invites/accept`, { method: 'POST', body: '{}' }),
  chainStatus: () => request('/chain/status'),
}
