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
    const msg = body?.msg || body?.message || body?.error || res.statusText
    throw new ApiError(msg, res.status)
  }
  return body
}

/** 带 JWT 拉取团队聊天文件，返回 blob URL（用毕请 revokeObjectURL）。 */
export async function fetchTeamChatFileBlob(teamId, messageId) {
  const token = getToken()
  const path = `/api/v1/teams/${encodeURIComponent(teamId)}/chat/files/${encodeURIComponent(messageId)}`
  const res = await fetch(path, {
    credentials: 'include',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) {
    throw new ApiError(res.statusText || 'download failed', res.status)
  }
  const blob = await res.blob()
  return URL.createObjectURL(blob)
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
  verifyPassword: (password) =>
    request('/users/me/verify-password', {
      method: 'POST',
      body: JSON.stringify({ password }),
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
  markTeamRead: (teamId) =>
    request(`/teams/${encodeURIComponent(teamId)}/read`, { method: 'POST', body: '{}' }),
  markAllTeamsRead: () => request('/teams/read-all', { method: 'POST', body: '{}' }),
  createTeam: (body) => request('/teams', { method: 'POST', body: JSON.stringify(body) }),
  getTeam: (id) => request(`/teams/${encodeURIComponent(id)}`),
  addTeamLedger: (teamId, ledgerId) =>
    request(`/teams/${encodeURIComponent(teamId)}/ledgers`, {
      method: 'POST',
      body: JSON.stringify({ ledgerId }),
    }),
  removeTeamLedger: (teamId, ledgerId) =>
    request(`/teams/${encodeURIComponent(teamId)}/ledgers/${encodeURIComponent(ledgerId)}`, {
      method: 'DELETE',
    }),
  listTeamMessages: (teamId, sinceId = 0, limit = 50) => {
    const q = new URLSearchParams()
    if (sinceId) q.set('sinceId', String(sinceId))
    if (limit) q.set('limit', String(limit))
    const qs = q.toString()
    return request(`/teams/${encodeURIComponent(teamId)}/messages${qs ? `?${qs}` : ''}`)
  },
  sendTeamMessage: (teamId, body) =>
    request(`/teams/${encodeURIComponent(teamId)}/messages`, {
      method: 'POST',
      body: JSON.stringify({ body }),
    }),
  uploadTeamChatFile: (teamId, file) => {
    const fd = new FormData()
    fd.append('file', file)
    return request(`/teams/${encodeURIComponent(teamId)}/messages/file`, { method: 'POST', body: fd })
  },
  teamChatFilePath: (teamId, messageId) =>
    `/api/v1/teams/${encodeURIComponent(teamId)}/chat/files/${encodeURIComponent(messageId)}`,
  refresh: () => request('/auth/refresh', { method: 'POST', body: '{}' }),
  logout: () => request('/auth/logout', { method: 'POST', body: '{}' }),
  health: () => request('/health'),
  chainStatus: () => request('/chain/status'),
  chainQueue: () => request('/chain/queue'),
  retryChainQueue: (id) =>
    request(`/chain/queue/${encodeURIComponent(id)}/retry`, { method: 'POST', body: '{}' }),
  chainBlocks: (page = 1, limit = 20) =>
    request(`/chain/blocks?page=${page}&limit=${limit}`),
  chainConsensus: () => request('/chain/consensus'),
  chainPeers: () => request('/chain/peers'),
  discoveryServices: () => request('/discovery/services'),
  listEntrySchemaTemplates: () => request('/entry-schema/templates'),
  listEntryTemplates: () => request('/entry-templates'),
  getEntryTemplate: (id) => request(`/entry-templates/${encodeURIComponent(id)}`),
  createEntryTemplate: (body) => request('/entry-templates', { method: 'POST', body: JSON.stringify(body) }),
  updateEntryTemplate: (id, body) =>
    request(`/entry-templates/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteEntryTemplate: (id) =>
    request(`/entry-templates/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  listLedgers: () => request('/ledgers'),
  getLedger: (id) => request(`/ledgers/${id}`),
  ragExport: (id) => request(`/ledgers/${encodeURIComponent(id)}/rag-export`),
  createLedger: (data) =>
    request('/ledgers', {
      method: 'POST',
      body: JSON.stringify({
        ...data,
        entrySchema: data.entrySchema || { templateId: 'default' },
      }),
    }),
  appendEntry: (id, entry) =>
    request(`/ledgers/${id}/entries`, {
      method: 'POST',
      body: JSON.stringify({
        entry: {
          signerId: entry.signerId,
          schemaId: entry.schemaId,
          data: entry.data,
        },
      }),
    }),
  proposeEntry: (id, entry) =>
    request(`/ledgers/${id}/entries/propose`, {
      method: 'POST',
      body: JSON.stringify({ entry }),
    }),
  listPending: (id) => request(`/ledgers/${id}/pending`),
  approvePending: (id, pendingId) =>
    request(`/ledgers/${id}/pending/${pendingId}/approve`, { method: 'POST', body: '{}' }),
  rejectPending: (id, pendingId) =>
    request(`/ledgers/${id}/pending/${pendingId}/reject`, { method: 'POST', body: '{}' }),
  inviteMember: (id, inviteeId, role = 'member') =>
    request(`/ledgers/${id}/members/invite`, {
      method: 'POST',
      body: JSON.stringify({ inviteeId, role }),
    }),
  listMyInvites: () => request('/ledgers/invites/mine'),
  listLedgerInvites: (id) => request(`/ledgers/${encodeURIComponent(id)}/invites`),
  acceptInvite: (id) =>
    request(`/ledgers/${id}/invites/accept`, { method: 'POST', body: '{}' }),
  syncLedger: (id, sinceSeq = 0) =>
    request(`/ledgers/${id}/sync?sinceSeq=${sinceSeq}`),
  setLedgerStorageLocation: (id, storageLocation) =>
    request(`/ledgers/${id}/storage-location`, {
      method: 'PATCH',
      body: JSON.stringify({ storageLocation }),
    }),
  updateLedger: (id, body) =>
    request(`/ledgers/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
  setLedgerApprovalPolicy: (id, approvalPolicy) =>
    request(`/ledgers/${id}/approval-policy`, {
      method: 'PATCH',
      body: JSON.stringify(approvalPolicy),
    }),
  enableLedgerEncryption: (id, encryption) =>
    request(`/ledgers/${id}/encryption/enable`, {
      method: 'POST',
      body: JSON.stringify(encryption),
    }),
  setLedgerPassphraseViewPolicy: (id, body) =>
    request(`/ledgers/${id}/encryption/passphrase-view-policy`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),
  registerLedgerPassphraseViewWrap: (id, body) =>
    request(`/ledgers/${id}/encryption/passphrase-view-wrap`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  archiveLedger: (id) => request(`/ledgers/${id}`, { method: 'DELETE' }),
  rotateGroupKeys: (id, wrappedKeys) =>
    request(`/ledgers/${id}/encryption/rotate`, {
      method: 'POST',
      body: JSON.stringify({ wrappedKeys }),
    }),
  putPublicKey: (publicKeyPem) =>
    request('/users/me/public-key', {
      method: 'PUT',
      body: JSON.stringify({ publicKeyPem }),
    }),
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
  downloadTemplate: (ledgerId) => {
    const q = ledgerId ? `?ledgerId=${encodeURIComponent(ledgerId)}` : ''
    return fetch(`${BASE}/import/template${q}`, {
      credentials: 'include',
      headers: { Authorization: `Bearer ${getToken() || ''}` },
    })
  },
  importPreview: (id, file) => {
    const fd = new FormData()
    fd.append('file', file)
    return request(`/ledgers/${id}/import/preview`, { method: 'POST', body: fd })
  },
  importCommit: (id, body) =>
    request(`/ledgers/${id}/import/commit`, { method: 'POST', body: JSON.stringify(body) }),
  ledgerBackup: (id, password) =>
    request(`/ledgers/${id}/backup`, { method: 'POST', body: JSON.stringify({ password }) }),
  restorePreview: (id, body) =>
    request(`/ledgers/${id}/restore/preview`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  restoreCommit: (id, body) =>
    request(`/ledgers/${id}/restore/commit`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  getAccountingChart: (id) => request(`/ledgers/${id}/accounting/chart`),
  putAccountingChart: (id, chart) =>
    request(`/ledgers/${id}/accounting/chart`, { method: 'PUT', body: JSON.stringify(chart) }),
  listAccountingJournals: (id) => request(`/ledgers/${id}/accounting/journals`),
  postAccountingJournal: (id, journal) =>
    request(`/ledgers/${id}/accounting/journals`, { method: 'POST', body: JSON.stringify(journal) }),
  listAccountingPeriods: (id) => request(`/ledgers/${id}/accounting/periods`),
  closeAccountingPeriod: (id, period) =>
    request(`/ledgers/${id}/accounting/periods/${encodeURIComponent(period)}/close`, {
      method: 'POST',
      body: '{}',
    }),
  reopenAccountingPeriod: (id, period) =>
    request(`/ledgers/${id}/accounting/periods/${encodeURIComponent(period)}/reopen`, {
      method: 'POST',
      body: '{}',
    }),
  getAccountingReports: (id, period = '') => {
    const q = period ? `?period=${encodeURIComponent(period)}` : ''
    return request(`/ledgers/${id}/accounting/reports${q}`)
  },
  listAccountingAttachments: (id, entrySeq = 0) => {
    const q = entrySeq ? `?entrySeq=${entrySeq}` : ''
    return request(`/ledgers/${id}/accounting/attachments${q}`)
  },
  uploadAccountingAttachment: (id, entrySeq, file) => {
    const fd = new FormData()
    fd.append('entrySeq', String(entrySeq))
    fd.append('file', file)
    return request(`/ledgers/${id}/accounting/attachments`, { method: 'POST', body: fd })
  },
  listBankStatements: (id) => request(`/ledgers/${id}/accounting/bank-statements`),
  importBankStatement: (id, file, accountCode = '1002') => {
    const fd = new FormData()
    fd.append('file', file)
    if (accountCode) fd.append('accountCode', accountCode)
    return request(`/ledgers/${id}/accounting/bank-statements/import`, { method: 'POST', body: fd })
  },
  matchBankLine: (id, stmtId, lineId, entrySeq) =>
    request(`/ledgers/${id}/accounting/bank-statements/${encodeURIComponent(stmtId)}/match`, {
      method: 'POST',
      body: JSON.stringify({ lineId, entrySeq }),
    }),
}
