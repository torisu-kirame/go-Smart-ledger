import type {
  AnchorResult,
  AppendEntryInput,
  CreateLedgerInput,
  Health,
  Ledger,
  LedgerEvent,
  VerifyResult,
} from '../types/ledger'
import type { CaptchaResp, LoginResp, RefreshResp } from '../types/auth'

const BASE = '/api/v1'

let getToken: (() => string | null) | null = null
let onUnauthorized: (() => Promise<boolean>) | null = null

export function configureAuth(
  tokenGetter: () => string | null,
  refreshFn: () => Promise<boolean>,
) {
  getToken = tokenGetter
  onUnauthorized = refreshFn
}

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function request<T>(path: string, init?: RequestInit, retry = true): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init?.headers as Record<string, string>),
  }
  const token = getToken?.()
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  const res = await fetch(`${BASE}${path}`, {
    credentials: 'include',
    ...init,
    headers,
  })

  if (res.status === 401 && retry && onUnauthorized && !path.startsWith('/auth/')) {
    const ok = await onUnauthorized()
    if (ok) return request<T>(path, init, false)
  }

  const text = await res.text()
  let body: unknown = null
  if (text) {
    try {
      body = JSON.parse(text)
    } catch {
      body = { error: text }
    }
  }
  if (!res.ok) {
    const msg =
      typeof body === 'object' && body && 'message' in body
        ? String((body as { message: string }).message)
        : typeof body === 'object' && body && 'error' in body
          ? String((body as { error: string }).error)
          : res.statusText
    throw new ApiError(msg, res.status)
  }
  return body as T
}

export const api = {
  captcha: () => request<CaptchaResp>('/auth/captcha'),

  login: (body: {
    username: string
    password: string
    captchaId: string
    captchaCode: string
  }) =>
    request<LoginResp>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  refresh: () =>
    request<RefreshResp>('/auth/refresh', {
      method: 'POST',
      body: '{}',
    }),

  logout: () =>
    request<void>('/auth/logout', {
      method: 'POST',
      body: '{}',
    }),

  health: () => request<Health>('/health'),

  listLedgers: () => request<Ledger[]>('/ledgers'),

  getLedger: (id: string) => request<Ledger>(`/ledgers/${id}`),

  createLedger: (data: CreateLedgerInput) =>
    request<Ledger>('/ledgers', { method: 'POST', body: JSON.stringify(data) }),

  appendEntry: (id: string, entry: AppendEntryInput) =>
    request<LedgerEvent>(`/ledgers/${id}/entries`, {
      method: 'POST',
      body: JSON.stringify({ entry }),
    }),

  listEvents: (id: string, from = 1, to = 0) => {
    const q = new URLSearchParams({ from: String(from) })
    if (to > 0) q.set('to', String(to))
    return request<LedgerEvent[]>(`/ledgers/${id}/events?${q}`)
  },

  anchor: (id: string, seqFrom = 0, seqTo = 0) =>
    request<AnchorResult>(`/ledgers/${id}/anchor`, {
      method: 'POST',
      body: JSON.stringify({ seqFrom, seqTo }),
    }),

  verify: (id: string) => request<VerifyResult>(`/ledgers/${id}/verify`),
}
