# UTF-8 hints after make up
param(
    [switch]$Legacy
)
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
Write-Host ''
Write-Host 'Web UI:     http://localhost:25173'
Write-Host 'Gateway:    http://localhost:28080/api/v1/health'
if ($Legacy) {
    Write-Host 'MiniLedger: http://localhost:24441/dashboard  (legacy profile)'
} else {
    Write-Host 'FISCO RPC:  http://127.0.0.1:20200  (host build_chain; see backend/infra/fisco/README.md)'
}
Write-Host 'Login:      admin / admin123'
Write-Host ''
Write-Host 'OpenClaw:   http://localhost:18789  (token: .env.openclaw)'
Write-Host ''
Write-Host 'AI：设置 -> AI -> API Key + Token -> 测试连接（Gateway 填 http://127.0.0.1:18789）'
Write-Host 'MiniLedger 回退：make legacy-up'
Write-Host '离线 Ollama：make offline-ai-up'
