import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import '../styles/login.css'

export function Login() {
  const navigate = useNavigate()
  const { login, user } = useAuth()
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [captchaId, setCaptchaId] = useState('')
  const [captchaCode, setCaptchaCode] = useState('')
  const [captchaImg, setCaptchaImg] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const loadCaptcha = async () => {
    try {
      const c = await api.captcha()
      setCaptchaId(c.captchaId)
      setCaptchaImg(c.image.startsWith('data:') ? c.image : `data:image/png;base64,${c.image}`)
      setCaptchaCode('')
    } catch {
      setError('无法加载验证码')
    }
  }

  useEffect(() => {
    if (user) {
      navigate('/', { replace: true })
      return
    }
    loadCaptcha()
  }, [user, navigate])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const res = await api.login({
        username,
        password,
        captchaId,
        captchaCode,
      })
      login(res.accessToken, res.user, res.expiresIn)
      navigate('/', { replace: true })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '登录失败')
      await loadCaptcha()
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <h1>Smart Ledger</h1>
        <p className="login-sub">区块链自定义账本控制台</p>
        {error && <div className="alert alert-error">{error}</div>}
        <form onSubmit={handleSubmit}>
          <div className="form-row">
            <label>用户名</label>
            <input
              required
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          </div>
          <div className="form-row">
            <label>密码</label>
            <input
              required
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          <div className="form-row">
            <label>验证码</label>
            <div className="captcha-row">
              <input
                required
                placeholder="输入图中字符"
                value={captchaCode}
                onChange={(e) => setCaptchaCode(e.target.value)}
              />
              {captchaImg && (
                <img
                  src={captchaImg}
                  alt="captcha"
                  className="captcha-img"
                  onClick={loadCaptcha}
                  title="点击刷新"
                />
              )}
            </div>
          </div>
          <button type="submit" className="btn-primary login-btn" disabled={loading}>
            {loading ? '登录中…' : '登录'}
          </button>
        </form>
        <p className="login-hint">默认账号 admin / admin123 · 长期令牌存 HttpOnly Cookie</p>
      </div>
    </div>
  )
}
