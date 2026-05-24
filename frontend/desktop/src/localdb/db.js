/**
 * 浏览器端 SQLite（sql.js），库文件持久化在 IndexedDB（用户本机磁盘）。
 * 权威数据仍在服务端 MiniLedger；本地库用于离线查看与多人账本增量同步。
 */
import initSqlJs from 'sql.js'
import sqlWasmUrl from 'sql.js/dist/sql-wasm.wasm?url'

const IDB_NAME = 'smart-ledger-local'
const IDB_STORE = 'sqlite'
const DB_KEY = 'main'

let sqlModule = null
let db = null

function openIdb() {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(IDB_NAME, 1)
    req.onupgradeneeded = () => {
      req.result.createObjectStore(IDB_STORE)
    }
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => reject(req.error)
  })
}

async function loadDbBytes() {
  const idb = await openIdb()
  return new Promise((resolve, reject) => {
    const tx = idb.transaction(IDB_STORE, 'readonly')
    const req = tx.objectStore(IDB_STORE).get(DB_KEY)
    req.onsuccess = () => resolve(req.result || null)
    req.onerror = () => reject(req.error)
  })
}

async function saveDbBytes(bytes) {
  const idb = await openIdb()
  return new Promise((resolve, reject) => {
    const tx = idb.transaction(IDB_STORE, 'readwrite')
    tx.objectStore(IDB_STORE).put(bytes, DB_KEY)
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
}

function runSchema(database) {
  database.run(`
    CREATE TABLE IF NOT EXISTS ledgers (
      id TEXT PRIMARY KEY,
      name TEXT,
      type TEXT,
      latest_seq INTEGER DEFAULT 0,
      latest_root TEXT,
      anchor_status TEXT,
      raw_json TEXT NOT NULL,
      synced_at TEXT
    );
    CREATE TABLE IF NOT EXISTS events (
      ledger_id TEXT NOT NULL,
      seq INTEGER NOT NULL,
      type TEXT,
      hash TEXT,
      signer_id TEXT,
      payload TEXT,
      created_at TEXT,
      PRIMARY KEY (ledger_id, seq)
    );
    CREATE TABLE IF NOT EXISTS sync_cursors (
      ledger_id TEXT PRIMARY KEY,
      since_seq INTEGER DEFAULT 0
    );
  `)
}

export async function initLocalDb() {
  if (db) return db
  if (!sqlModule) {
    sqlModule = await initSqlJs({
      locateFile: () => sqlWasmUrl,
    })
  }
  const saved = await loadDbBytes()
  db = saved ? new sqlModule.Database(saved) : new sqlModule.Database()
  runSchema(db)
  if (!saved) await persistLocalDb()
  return db
}

export async function persistLocalDb() {
  if (!db) return
  const bytes = db.export()
  await saveDbBytes(bytes)
}

export async function getSyncCursor(ledgerId) {
  await initLocalDb()
  const stmt = db.prepare('SELECT since_seq FROM sync_cursors WHERE ledger_id = ?')
  stmt.bind([ledgerId])
  let seq = 0
  if (stmt.step()) seq = stmt.getAsObject().since_seq || 0
  stmt.free()
  return seq
}

async function setSyncCursor(ledgerId, sinceSeq) {
  db.run(
    `INSERT INTO sync_cursors (ledger_id, since_seq) VALUES (?, ?)
     ON CONFLICT(ledger_id) DO UPDATE SET since_seq = excluded.since_seq`,
    [ledgerId, sinceSeq]
  )
}

function upsertLedger(ledger) {
  if (!ledger?.id) return
  const now = new Date().toISOString()
  db.run(
    `INSERT INTO ledgers (id, name, type, latest_seq, latest_root, anchor_status, raw_json, synced_at)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?)
     ON CONFLICT(id) DO UPDATE SET
       name = excluded.name,
       type = excluded.type,
       latest_seq = excluded.latest_seq,
       latest_root = excluded.latest_root,
       anchor_status = excluded.anchor_status,
       raw_json = excluded.raw_json,
       synced_at = excluded.synced_at`,
    [
      ledger.id,
      ledger.name || '',
      ledger.type || '',
      ledger.latestSeq ?? 0,
      ledger.latestRoot || '',
      ledger.anchorStatus || '',
      JSON.stringify(ledger),
      now,
    ]
  )
}

function upsertEvents(ledgerId, events) {
  for (const ev of events || []) {
    const seq = ev.seq ?? ev.Seq
    if (seq == null) continue
    db.run(
      `INSERT INTO events (ledger_id, seq, type, hash, signer_id, payload, created_at)
       VALUES (?, ?, ?, ?, ?, ?, ?)
       ON CONFLICT(ledger_id, seq) DO UPDATE SET
         type = excluded.type,
         hash = excluded.hash,
         signer_id = excluded.signer_id,
         payload = excluded.payload,
         created_at = excluded.created_at`,
      [
        ledgerId,
        seq,
        ev.type || '',
        ev.hash || '',
        ev.signerId || ev.signer_id || '',
        typeof ev.payload === 'string' ? ev.payload : JSON.stringify(ev.payload ?? {}),
        ev.createdAt || ev.created_at || '',
      ]
    )
  }
}

/** Pull incremental events from server and merge into local SQLite. */
export async function syncLedgerToLocal(api, ledgerId) {
  await initLocalDb()
  const sinceSeq = await getSyncCursor(ledgerId)
  const res = await api.syncLedger(ledgerId, sinceSeq)
  const ledger = res.ledger
  const events = res.events || []
  upsertLedger(ledger)
  upsertEvents(ledgerId, events)
  const latest = ledger?.latestSeq ?? sinceSeq
  await setSyncCursor(ledgerId, latest)
  await persistLocalDb()
  return {
    ledger,
    events,
    sinceSeq,
    latestSeq: latest,
    newCount: events.length,
  }
}

export async function listLocalLedgers() {
  await initLocalDb()
  const rows = []
  const stmt = db.prepare(
    'SELECT id, name, type, latest_seq, anchor_status, synced_at FROM ledgers ORDER BY synced_at DESC'
  )
  while (stmt.step()) rows.push(stmt.getAsObject())
  stmt.free()
  return rows
}

export async function listLocalEvents(ledgerId, limit = 500) {
  await initLocalDb()
  const rows = []
  const stmt = db.prepare(
    `SELECT seq, type, hash, signer_id, payload, created_at FROM events
     WHERE ledger_id = ? ORDER BY seq DESC LIMIT ?`
  )
  stmt.bind([ledgerId, limit])
  while (stmt.step()) rows.push(stmt.getAsObject())
  stmt.free()
  return rows
}

export async function getLocalDbStats() {
  await initLocalDb()
  const ledgers = db.exec('SELECT COUNT(*) AS c FROM ledgers')[0]?.values[0][0] ?? 0
  const events = db.exec('SELECT COUNT(*) AS c FROM events')[0]?.values[0][0] ?? 0
  return { ledgers, events }
}
