param(
    [string] $Backend = "go"
)

$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
Write-Host ''
if ($Backend -eq 'java') {
    Write-Host 'Backend:    Java (java-backend/)'
} else {
    Write-Host 'Backend:    Go (backend/)'
}
Write-Host 'Web UI:     http://localhost:25173'
Write-Host 'Gateway:    http://localhost:28080/api/v1/health'
Write-Host 'MiniLedger: http://localhost:24441/dashboard'
Write-Host 'Login:      admin / admin123'
Write-Host ''
if ($Backend -eq 'java') {
    Write-Host 'Note: AI/LangChain 请使用 make up-go'
} else {
    Write-Host 'AI：设置 -> AI -> 配置 LLM -> 测试连接'
    Write-Host '离线 Ollama：make offline-ai-up'
}
Write-Host ''
