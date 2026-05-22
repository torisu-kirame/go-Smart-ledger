import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import { StatusDot } from '../components/StatusDot'
import type { Health, Ledger } from '../types/ledger'

export function Dashboard() {
  const [health, setHealth] = useState<Health | null>(null)
  const [ledgers, setLedgers] = useState<Ledger[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      setLoading(true)
      setError('')
      try {
        const [h, list] = await Promise.all([api.health(), api.listLedgers().catch(() => [])])
        if (!cancelled) {
          setHealth(h)
          setLedgers(list)
        }
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof ApiError ? e.message : '无法连接网关，请确认后端已启动')
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  const multi = ledgers.filter((l) => l.type === 'multi').length
  const priv = ledgers.filter((l) => l.type === 'private').length

  return (
    <>
      <header className="topbar">
        <h2>系统概览</h2>
      </header>
      <div className="content">
        {error && <div className="alert alert-error">{error}</div>}

        <div className="grid-3">
          <div className="card">
            <h3>API 网关</h3>
            <div className="value">
              {loading ? '…' : <StatusDot ok={health?.status === 'ok'} label={health?.status ?? '未知'} />}
            </div>
            <p className="sub">http://localhost:8080</p>
          </div>
          <div className="card">
            <h3>MiniLedger 链</h3>
            <div className="value">
              {loading ? (
                '…'
              ) : (
                <StatusDot
                  ok={!!health?.miniLedgerOnline}
                  label={health?.miniLedgerOnline ? '在线' : '离线'}
                />
              )}
            </div>
            <p className="sub">许可链节点 · Raft 共识</p>
          </div>
          <div className="card">
            <h3>账本总数</h3>
            <div className="value">{loading ? '…' : ledgers.length}</div>
            <p className="sub">
              私人 {priv} · 多人 {multi}
            </p>
          </div>
        </div>

        <div className="panel">
          <div className="panel-header">
            <h3>快捷操作</h3>
          </div>
          <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
            <Link to="/ledgers">
              <button type="button" className="btn-primary">
                管理账本
              </button>
            </Link>
            <a href="http://localhost:4441/dashboard" target="_blank" rel="noreferrer">
              <button type="button" className="btn-ghost">
                区块浏览器
              </button>
            </a>
          </div>
        </div>

        <div className="panel">
          <div className="panel-header">
            <h3>最近账本</h3>
            <Link to="/ledgers">查看全部 →</Link>
          </div>
          {ledgers.length === 0 ? (
            <p className="empty">暂无账本，前往账本管理创建</p>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>名称</th>
                    <th>类型</th>
                    <th>序号</th>
                    <th>锚定</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {ledgers.slice(0, 5).map((l) => (
                    <tr key={l.id}>
                      <td>{l.name}</td>
                      <td>
                        <span className={`badge badge-${l.type}`}>{l.type === 'multi' ? '多人' : '私人'}</span>
                      </td>
                      <td className="mono">{l.latestSeq}</td>
                      <td>
                        <span className={`badge badge-${l.anchorStatus === 'synced' ? 'ok' : 'pending'}`}>
                          {l.anchorStatus}
                        </span>
                      </td>
                      <td>
                        <Link to={`/ledgers/${l.id}`}>详情</Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </>
  )
}
