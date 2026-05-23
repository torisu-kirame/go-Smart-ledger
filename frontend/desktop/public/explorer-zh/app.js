const $ = (sel) => document.querySelector(sel)
const view = $('#view')
const titles = {
  '/': '概览',
  '/blocks': '区块',
  '/transactions': '交易',
  '/state': '状态',
  '/network': '节点与共识',
}

function route() {
  const hash = location.hash.slice(1) || '/'
  return hash.startsWith('/') ? hash : `/${hash}`
}

function setNav(routePath) {
  document.querySelectorAll('.nav-link').forEach((a) => {
    a.classList.toggle('active', a.dataset.route === routePath)
  })
  $('#page-title').textContent = titles[routePath] || '浏览器'
}

async function api(path) {
  const res = await fetch(path)
  if (!res.ok) throw new Error(`${path} ${res.status}`)
  return res.json()
}

function shortHash(h) {
  if (!h || h.length < 16) return h || '—'
  return `${h.slice(0, 10)}…${h.slice(-6)}`
}

function fmtTime(ts) {
  if (!ts) return '—'
  const n = Number(ts)
  if (Number.isNaN(n)) return String(ts)
  return new Date(n).toLocaleString('zh-CN')
}

async function refreshStatus() {
  try {
    const st = await api('/status')
    $('#chain-height').textContent = st.height ?? st.chainHeight ?? '—'
    $('#chain-uptime').textContent = st.uptime ?? '—'
    const dot = $('#status-dot')
    const txt = $('#status-text')
    dot.className = 'dot ok'
    txt.textContent = `在线 · ${st.role || '节点'}`
    return st
  } catch {
    $('#status-dot').className = 'dot err'
    $('#status-text').textContent = '离线'
    return null
  }
}

async function renderOverview() {
  const st = await refreshStatus()
  let blocks = []
  try {
    const data = await api('/blocks?page=1&limit=5')
    blocks = data.blocks || []
  } catch { /* ignore */ }

  view.innerHTML = `
    <div class="grid">
      <div class="card"><h3>区块高度</h3><div class="val">${st?.height ?? st?.chainHeight ?? '—'}</div></div>
      <div class="card"><h3>共识角色</h3><div class="val">${st?.role ?? '—'}</div></div>
      <div class="card"><h3>节点 ID</h3><div class="val mono" style="font-size:0.75rem">${shortHash(st?.nodeId)}</div></div>
    </div>
    <div class="panel">
      <h2>最近区块</h2>
      ${blocks.length ? tableBlocks(blocks) : '<p class="muted">暂无数据</p>'}
    </div>`
}

function tableBlocks(blocks) {
  return `<table><thead><tr><th>高度</th><th>哈希</th><th>交易数</th><th>时间</th></tr></thead><tbody>
    ${blocks.map((b) => `<tr>
      <td>${b.height}</td>
      <td class="mono">${shortHash(b.hash)}</td>
      <td>${b.transactions?.length ?? b.txCount ?? '—'}</td>
      <td>${fmtTime(b.timestamp)}</td>
    </tr>`).join('')}
  </tbody></table>`
}

async function renderBlocks() {
  await refreshStatus()
  const data = await api('/blocks?page=1&limit=20')
  const blocks = data.blocks || []
  view.innerHTML = `<div class="panel"><h2>区块列表</h2>
    ${blocks.length ? tableBlocks(blocks) : '<p class="muted">暂无区块</p>'}
  </div>`
}

async function renderTransactions() {
  await refreshStatus()
  let txs = []
  try {
    const data = await api('/tx/recent?limit=30')
    txs = data.transactions || data.items || (Array.isArray(data) ? data : [])
  } catch { /* ignore */ }

  view.innerHTML = `<div class="panel"><h2>最近交易</h2>
    ${txs.length ? `<table><thead><tr><th>哈希</th><th>类型</th><th>时间</th></tr></thead><tbody>
      ${txs.map((t) => `<tr>
        <td class="mono">${shortHash(t.hash)}</td>
        <td>${t.type || '—'}</td>
        <td>${fmtTime(t.timestamp)}</td>
      </tr>`).join('')}
    </tbody></table>` : '<p class="muted">暂无交易</p>'}
  </div>`
}

async function renderState() {
  await refreshStatus()
  let rows = []
  try {
    const data = await api('/state?page=1&limit=30')
    rows = data.entries || data.items || data.rows || []
  } catch { /* ignore */ }

  view.innerHTML = `<div class="panel"><h2>世界状态（前 30 条）</h2>
    ${rows.length ? `<table><thead><tr><th>键</th><th>值</th></tr></thead><tbody>
      ${rows.map((r) => `<tr>
        <td class="mono">${r.key}</td>
        <td class="mono">${typeof r.value === 'string' ? r.value.slice(0, 80) : JSON.stringify(r.value).slice(0, 80)}…</td>
      </tr>`).join('')}
    </tbody></table>` : '<p class="muted">暂无状态条目</p>'}
  </div>`
}

async function renderNetwork() {
  await refreshStatus()
  let consensus = null
  let peers = null
  try { consensus = await api('/consensus') } catch { /* */ }
  try { peers = await api('/peers') } catch { /* */ }

  view.innerHTML = `
    <div class="panel"><h2>共识</h2>
      <pre class="json">${consensus ? JSON.stringify(consensus, null, 2) : '无法读取'}</pre>
    </div>
    <div class="panel"><h2>节点</h2>
      <pre class="json">${peers ? JSON.stringify(peers, null, 2) : '无法读取'}</pre>
    </div>`
}

async function render() {
  const r = route()
  setNav(r)
  view.innerHTML = '<p class="muted">加载中…</p>'
  try {
    if (r === '/') await renderOverview()
    else if (r === '/blocks') await renderBlocks()
    else if (r === '/transactions') await renderTransactions()
    else if (r === '/state') await renderState()
    else if (r === '/network') await renderNetwork()
    else await renderOverview()
  } catch (e) {
    view.innerHTML = `<p class="err">加载失败：${e.message}</p>`
  }
}

window.addEventListener('hashchange', render)
render()
setInterval(refreshStatus, 15000)
