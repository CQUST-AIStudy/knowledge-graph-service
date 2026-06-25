# 知识图谱服务本地启动脚本
# 使用 go run 直接运行，适合开发调试

$ErrorActionPreference = "Stop"
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$env:PYTHONIOENCODING = "utf-8"

Set-Location $PSScriptRoot

# 检查 go 命令
$goCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCmd) {
    Write-Host "错误：未找到 go 命令，请先安装 Go 1.25+" -ForegroundColor Red
    exit 1
}

# 检查 .env 文件
if (-not (Test-Path ".env")) {
    Write-Host "提示：未找到 .env 文件，可从 .env.example 复制" -ForegroundColor Yellow
    Write-Host "  Copy-Item .env.example .env" -ForegroundColor Gray
}

# 下载依赖
Write-Host "正在下载 Go 依赖..." -ForegroundColor Cyan
go mod download

# 启动服务
$addr = $env:KG_ADDR
if (-not $addr) { $addr = ":10171" }
Write-Host "知识图谱服务启动中: http://127.0.0.1$addr" -ForegroundColor Green

go run ./cmd/knowledge-graph-api

exit $LASTEXITCODE
