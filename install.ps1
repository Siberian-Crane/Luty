# install.ps1 - 上传到 GitHub 仓库
$repo = "Siberian-Crane/luty"
$installPath = "$env:LOCALAPPDATA\Luty"

Write-Host "正在下载 luty..." -ForegroundColor Cyan

# 获取最新版本
$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$asset = $release.assets | Where-Object { $_.name -match "windows-amd64" }

# 下载
New-Item -ItemType Directory -Force -Path $installPath | Out-Null
$url = $asset.browser_download_url
$output = "$installPath\luty.exe"
Invoke-WebRequest -Uri $url -OutFile $output

# 添加到 PATH
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($currentPath -notlike "*$installPath*") {
    [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$installPath", "User")
    $env:PATH = "$env:PATH;$installPath"
}

Write-Host "✅ 安装成功！" -ForegroundColor Green
Write-Host "请重新打开终端，然后运行: luty --help"