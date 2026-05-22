import type {
  AnchorResult,
  AppendEntryInput,
  CreateLedgerInput,
  Health,
  Ledger,
  LedgerEvent,
  VerifyResult,
} from '../types/ledger'

const BASE = '/api/v1'

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
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
