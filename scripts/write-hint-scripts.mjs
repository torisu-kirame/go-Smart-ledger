import { writeFileSync } from 'node:fs'

const bom = '\uFEFF'

const up = `# UTF-8 hints after make up
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
Write-Host ''
Write-Host 'Web UI:     http://localhost:25173'
Write-Host 'Gateway:    http://localhost:28080/api/v1/health'
Write-Host 'MiniLedger: http://localhost:24441/dashboard'
Write-Host 'Login:      admin / admin123'
Write-Host ''
Write-Host 'OpenClaw:   http://localhost:18789  (token: .env.openclaw)'
Write-Host ''
Write-Host 'AI：设置 -> AI -> API Key + Token -> 测试连接（Gateway 填 http://127.0.0.1:18789）'
Write-Host '离线 Ollama：make offline-ai-up'
`

const offline = `# UTF-8 hints after make offline-ai-up
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
Write-Host ''
Write-Host 'Ollama:     http://localhost:11434/v1'
Write-Host 'OpenClaw:   http://localhost:18789  (token: .env.openclaw)'
Write-Host '设置 -> AI：选择 Ollama（本地离线），API 地址 http://127.0.0.1:11434/v1'
`

writeFileSync('scripts/print-up-hints.ps1', bom + up)
writeFileSync('scripts/print-offline-ai-hints.ps1', bom + offline)
writeFileSync(
  'scripts/print-dev-hint.ps1',
  bom +
    `$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
Write-Host '请确保已执行: make up'
`
)
console.log('ok')
