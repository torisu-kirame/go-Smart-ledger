import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import type { AppendEntryInput, Ledger, LedgerEvent } from '../types/ledger'

const today = () => new Date().toISOString().slice(0, 10)

export function LedgerDetail() {
  const { id } = useParams<{ id: string }>()
  const [ledger, setLedger] = useState<Ledger | null>(null)
  const [events, setEvents] = useState<LedgerEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [busy, setBusy] = useState(false)
  const [verifyOk, setVerifyOk] = useState<boolean | null>(null)

  const [entry, setEntry] = useState<AppendEntryInput>({
    signerId: '',
    date: today(),
    type: 'expense',
    amount: '',
    category: '',
    note: '',
  })

  const load = useCallback(async () => {
    if (!id) return
    setLoading(true)
    setError('')
    try {
      const [l, ev] = await Promise.all([api.getLedger(id), api.listEvents(id)])
      setLedger(l)
      setEvents(ev)
      setEntry((e) =>
        e.signerId ? e : l.members[0] ? { ...e, signerId: l.members[0].id } : e,
      )
    } catch (e) {
      setError(e instanceof ApiError ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    load()
  }, [load])

  const handleEntry = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!id) return
    setBusy(true)
    setError('')
    setMessage('')
    try {
      await api.appendEntry(id, entry)
      setMessage('记账成功')
      setEntry((x) => ({ ...x, amount: '', note: '', category: '' }))
      await load()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '记账失败')
    } finally {
      setBusy(false)
    }
  }

  const handleAnchor = async () => {
    if (!id) return
    setBusy(true)
    setError('')
    setMessage('')
    try {
      const res = await api.anchor(id)
      setMessage(`封账锚定成功 · seq ${res.seq} · ${res.status}`)
      await load()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '锚定失败')
    } finally {
      setBusy(false)
    }
  }

  const handleVerify = async () => {
    if (!id) return
    setBusy(true)
    setError('')
    try {
      const res = await api.verify(id)
      setVerifyOk(res.valid)
      setMessage(res.valid ? 'Merkle 完整性校验通过' : '校验未通过，请检查链上数据')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '校验失败')
    } finally {
      setBusy(false)
    }
  }

  if (!id) return null

  const typeLabel = ledger?.type === 'multi' ? '多人' : '私人'

  return (
    <>
      <header className="topbar">
        <h2>
          <Link to="/ledgers" style={{ color: 'var(--text-muted)', fontWeight: 500, marginRight: '0.5rem' }}>
            ←
          </Link>
          {loading ? '加载中…' : ledger?.name ?? id}
        </h2>
      </header>
      <div className="content">
        {error && <div className="alert alert-error">{error}</div>}
        {message && <div className="alert alert-success">{message}</div>}

        {ledger && (
          <>
            <div className="grid-3">
              <div className="card">
                <h3>类型</h3>
                <div className="value">
                  <span className={`badge badge-${ledger.type}`}>{typeLabel}</span>
                </div>
              </div>
              <div className="card">
                <h3>最新序号</h3>
                <div className="value mono">{ledger.latestSeq}</div>
              </div>
              <div className="card">
                <h3>锚定状态</h3>
                <div className="value">
                  <span className={`badge badge-${ledger.anchorStatus === 'synced' ? 'ok' : 'pending'}`}>
                    {ledger.anchorStatus}
                  </span>
                </div>
              </div>
            </div>

            <div className="panel">
              <h3 style={{ margin: '0 0 0.75rem', fontSize: '0.875rem', color: 'var(--text-muted)' }}>
                账本 ID
              </h3>
              <p className="mono" style={{ margin: '0 0 1rem', wordBreak: 'break-all' }}>
                {ledger.id}
              </p>
              <p className="mono" style={{ margin: 0, fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                Merkle Root: {ledger.latestRoot}
              </p>
              {verifyOk !== null && (
                <p style={{ marginTop: '0.75rem' }}>
                  完整性: {verifyOk ? '✓ 通过' : '✗ 失败'}
                </p>
              )}
            </div>

            <div className="panel">
              <div className="panel-header">
                <h3>链上操作</h3>
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <button type="button" className="btn-ghost" disabled={busy} onClick={handleVerify}>
                    校验完整性
                  </button>
                  <button type="button" className="btn-primary" disabled={busy} onClick={handleAnchor}>
                    封账并锚定
                  </button>
                </div>
              </div>
            </div>

            <div className="panel">
              <div className="panel-header">
                <h3>记一笔</h3>
              </div>
              <form onSubmit={handleEntry}>
                <div className="form-grid">
                  <div className="form-row">
                    <label>记账人 ID</label>
                    <select
                      value={entry.signerId}
                      onChange={(e) => setEntry((x) => ({ ...x, signerId: e.target.value }))}
                    >
                      {ledger.members.map((m) => (
                        <option key={m.id} value={m.id}>
                          {m.id}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="form-row">
                    <label>日期</label>
                    <input
                      type="date"
                      required
                      value={entry.date}
                      onChange={(e) => setEntry((x) => ({ ...x, date: e.target.value }))}
                    />
                  </div>
                  <div className="form-row">
                    <label>类型</label>
                    <select
                      value={entry.type}
                      onChange={(e) =>
                        setEntry((x) => ({ ...x, type: e.target.value as 'income' | 'expense' }))
                      }
                    >
                      <option value="expense">支出</option>
                      <option value="income">收入</option>
                    </select>
                  </div>
                  <div className="form-row">
                    <label>金额</label>
                    <input
                      required
                      placeholder="128.50"
                      value={entry.amount}
                      onChange={(e) => setEntry((x) => ({ ...x, amount: e.target.value }))}
                    />
                  </div>
                  <div className="form-row">
                    <label>分类</label>
                    <input
                      value={entry.category ?? ''}
                      onChange={(e) => setEntry((x) => ({ ...x, category: e.target.value }))}
                    />
                  </div>
                  <div className="form-row">
                    <label>备注</label>
                    <input
                      value={entry.note ?? ''}
                      onChange={(e) => setEntry((x) => ({ ...x, note: e.target.value }))}
                    />
                  </div>
                </div>
                <button type="submit" className="btn-primary" disabled={busy}>
                  提交到链
                </button>
              </form>
            </div>

            <div className="panel">
              <div className="panel-header">
                <h3>事件流水 ({events.length})</h3>
                <button type="button" className="btn-ghost" onClick={load} disabled={loading}>
                  刷新
                </button>
              </div>
              {events.length === 0 ? (
                <p className="empty">暂无事件</p>
              ) : (
                <div className="table-wrap">
                  <table>
                    <thead>
                      <tr>
                        <th>Seq</th>
                        <th>类型</th>
                        <th>签名者</th>
                        <th>哈希</th>
                        <th>时间</th>
                      </tr>
                    </thead>
                    <tbody>
                      {[...events].reverse().map((ev) => (
                        <tr key={ev.seq}>
                          <td className="mono">{ev.seq}</td>
                          <td>{ev.type}</td>
                          <td>{ev.signerId ?? '—'}</td>
                          <td className="mono">{ev.hash.slice(0, 16)}…</td>
                          <td>{new Date(ev.createdAt).toLocaleString('zh-CN')}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </>
  )
}
