#!/usr/bin/env pwsh
#Requires -Version 7.0
# SZTU-Autologin 打包脚本

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
$ErrorActionPreference = "Stop"
$PSDefaultParameterValues['*:Encoding'] = 'utf8'

Write-Host "检查 uv 环境..." -ForegroundColor Cyan
if (-not (Get-Command uv -ErrorAction SilentlyContinue)) {
    Write-Host "错误: 未找到 uv，请先安装 https://docs.astral.sh/uv/" -ForegroundColor Red
    exit 1
}

Write-Host "安装/同步依赖..." -ForegroundColor Cyan
uv sync 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "依赖同步失败" -ForegroundColor Red
    exit 1
}

Write-Host "`n使用 PyInstaller 打包..." -ForegroundColor Cyan
uv run pyinstaller -F -w --name SZTU-Autologin main.py 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "打包失败，请检查输出信息" -ForegroundColor Red
    exit 1
}

$distPath = Join-Path $PSScriptRoot "dist"
$exePath = Join-Path $distPath "SZTU-Autologin.exe"

if (Test-Path $exePath) {
    Write-Host "`n打包完成！" -ForegroundColor Green
    Write-Host "可执行文件: $exePath" -ForegroundColor Cyan
} else {
    Write-Host "`n打包失败，未生成可执行文件" -ForegroundColor Red
    exit 1
}
