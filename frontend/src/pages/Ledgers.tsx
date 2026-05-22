import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import type { CreateLedgerInput, Ledger, LedgerType, Member } from '../types/ledger'

function emptyMember(): Member {
  return { id: '', address: '' }
}

export function Ledgers() {
  const [list, setList] = useState<Ledger[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const [form, setForm] = useState<CreateLedgerInput>({
    type: 'private',
    name: '',
    creatorId: 'user1',
    members: [emptyMember()],
  })

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setList(await api.listLedgers())
    } catch (e) {
      setError(e instanceof ApiError ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const setType = (type: LedgerType) => {
    setForm((f) => ({
      ...f,
      type,
      members:
        type === 'private'
          ? [f.members[0] ?? emptyMember()]
          : f.members.length >= 2
            ? f.members
            : [f.members[0] ?? emptyMember(), emptyMember()],
    }))
  }

  const updateMember = (i: number, field: keyof Member, value: string) => {
    setForm((f) => {
      const members = [...f.members]
      members[i] = { ...members[i], [field]: value }
      return { ...f, members }
    })
  }

  const addMember = () => setForm((f) => ({ ...f, members: [...f.members, emptyMember()] }))

  const removeMember = (i: number) => {
    setForm((f) => ({ ...f, members: f.members.filter((_, idx) => idx !== i) }))
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError('')
    try {
      const members = form.members.filter((m) => m.id.trim())
      if (form.type === 'multi' && members.length < 2) {
        throw new Error('多人账本至少需要 2 名成员')
      }
      await api.createLedger({ ...form, members })
      setShowCreate(false)
      setForm({ type: 'private', name: '', creatorId: form.creatorId, members: [emptyMember()] })
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <>
      <header className="topbar">
        <h2>账本管理</h2>
      </header>
      <div className="content">
        {error && <div className="alert alert-error">{error}</div>}

        <div className="panel">
          <div className="panel-header">
            <h3>全部账本 {loading ? '' : `(${list.length})`}</h3>
            <button type="button" className="btn-primary" onClick={() => setShowCreate(true)}>
              创建账本
            </button>
          </div>

          {loading ? (
            <p className="empty">加载中…</p>
          ) : list.length === 0 ? (
            <p className="empty">暂无账本，点击「创建账本」开始</p>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>名称</th>
                    <th>ID</th>
                    <th>类型</th>
                    <th>成员</th>
                    <th>事件序</th>
                    <th>状态</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {list.map((l) => (
                    <tr key={l.id}>
                      <td>{l.name}</td>
                      <td className="mono">{l.id.slice(0, 12)}…</td>
                      <td>
                        <span className={`badge badge-${l.type}`}>
                          {l.type === 'multi' ? '多人' : '私人'}
                        </span>
                      </td>
                      <td>{l.members.length}</td>
                      <td className="mono">{l.latestSeq}</td>
                      <td>
                        <span className={`badge badge-${l.anchorStatus === 'synced' ? 'ok' : 'pending'}`}>
                          {l.anchorStatus}
                        </span>
                      </td>
                      <td>
                        <Link to={`/ledgers/${l.id}`}>进入 →</Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {showCreate && (
        <div className="modal-backdrop" onClick={() => setShowCreate(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>创建账本</h3>
            <form onSubmit={handleCreate}>
              <div className="form-row">
                <label>账本类型</label>
                <select
                  value={form.type}
                  onChange={(e) => setType(e.target.value as LedgerType)}
                >
                  <option value="private">私人账本（1 人）</option>
                  <option value="multi">多人账本（≥2 人）</option>
                </select>
              </div>
              <div className="form-row">
                <label>账本名称</label>
                <input
                  required
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  placeholder="例如：家庭开支、项目组"
                />
              </div>
              <div className="form-row">
                <label>创建者 ID</label>
                <input
                  required
                  value={form.creatorId}
                  onChange={(e) => setForm((f) => ({ ...f, creatorId: e.target.value }))}
                />
              </div>
              <div className="form-row">
                <label>成员 {form.type === 'multi' && '（至少 2 人）'}</label>
                {form.members.map((m, i) => (
                  <div key={i} className="member-row">
                    <input
                      placeholder="成员 ID"
                      required
                      value={m.id}
                      onChange={(e) => updateMember(i, 'id', e.target.value)}
                    />
                    <input
                      placeholder="地址 / 公钥"
                      value={m.address}
                      onChange={(e) => updateMember(i, 'address', e.target.value)}
                    />
                    {form.type === 'multi' && form.members.length > 2 && (
                      <button type="button" className="btn-ghost" onClick={() => removeMember(i)}>
                        删
                      </button>
                    )}
                  </div>
                ))}
                {form.type === 'multi' && (
                  <button type="button" className="btn-ghost" onClick={addMember}>
                    + 添加成员
                  </button>
                )}
              </div>
              <div className="modal-actions">
                <button type="button" className="btn-ghost" onClick={() => setShowCreate(false)}>
                  取消
                </button>
                <button type="submit" className="btn-primary" disabled={submitting}>
                  {submitting ? '创建中…' : '创建'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </>
  )
}
