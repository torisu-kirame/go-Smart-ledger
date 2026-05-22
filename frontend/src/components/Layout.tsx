import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import '../styles/layout.css'

const nav = [
  { to: '/', label: '概览', end: true },
  { to: '/ledgers', label: '账本管理' },
]

export function Layout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <h1>Smart Ledger</h1>
          <p>区块链自定义账本控制台</p>
        </div>
        <ul className="sidebar-nav">
          {nav.map((item) => (
            <li key={item.to}>
              <NavLink to={item.to} end={item.end}>
                {item.label}
              </NavLink>
            </li>
          ))}
        </ul>
        <div className="sidebar-footer">
          <div className="sidebar-user">{user?.username ?? '—'}</div>
          <button type="button" className="btn-ghost sidebar-logout" onClick={handleLogout}>
            退出登录
          </button>
          <span>MiniLedger 浏览器</span>
          <a href="http://localhost:24441/dashboard" target="_blank" rel="noreferrer">
            打开链节点面板 →
          </a>
        </div>
      </aside>
      <div className="main">
        <Outlet />
      </div>
    </div>
  )
}
