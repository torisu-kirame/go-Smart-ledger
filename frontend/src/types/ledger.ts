export type LedgerType = 'private' | 'multi'

export interface Member {
  id: string
  address: string
  role?: string
}

export interface Ledger {
  id: string
  type: string
  name: string
  creatorId: string
  members: Member[]
  latestSeq: number
  latestRoot: string
  blockHeight?: number
  anchorStatus: string
}

export interface LedgerEvent {
  seq: number
  type: string
  hash: string
  signerId?: string
  createdAt: string
}

export interface Health {
  status: string
  miniLedgerOnline?: boolean
  gateway?: string
}

export interface VerifyResult {
  valid: boolean
}

export interface AnchorResult {
  ledgerId: string
  seq: number
  root: string
  txHash?: string
  status: string
}

export interface CreateLedgerInput {
  type: LedgerType
  name: string
  creatorId: string
  members: Member[]
}

export interface AppendEntryInput {
  signerId: string
  date: string
  type: 'income' | 'expense'
  amount: string
  category?: string
  note?: string
}
