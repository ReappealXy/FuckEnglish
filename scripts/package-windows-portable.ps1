param(
  [Parameter(Mandatory = $true)][string]$OutputDirectory,
  [string]$WordbookDirectory
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$stage = Join-Path $root "build\windows-amd64"
$config = Get-Content -LiteralPath (Join-Path $root "mygo.config.ts") -Raw
$versionMatch = [regex]::Match($config, 'version:\s*"([^"]+)"')
if (-not $versionMatch.Success) { throw "Cannot read the application version." }
$version = $versionMatch.Groups[1].Value
if (-not $WordbookDirectory) { $WordbookDirectory = Join-Path $root "words" }
$wordbooks = @(Get-ChildItem -LiteralPath $WordbookDirectory -File -Filter "*.xlsx" | Where-Object { -not $_.Name.StartsWith('~$') })
if ($wordbooks.Count -eq 0) { throw "No XLSX wordbooks were found." }

$destination = [System.IO.Path]::GetFullPath($OutputDirectory)
if (Test-Path -LiteralPath $destination) { throw "Output directory already exists: $destination" }
$folderName = "FuckEnglish $version 免安装版"
$portable = Join-Path $destination $folderName
$zipPath = Join-Path $destination "FuckEnglish-$version-Portable-windows-x64.zip"
$installerName = "FuckEnglish Setup $version amd64.exe"
$required = @("FuckEnglish.exe", $installerName, "BrimDeck-themes", "licenses", "icon.png", "NOTICE-BrimDeck.txt")
foreach ($item in $required) {
  if (-not (Test-Path -LiteralPath (Join-Path $stage $item))) { throw "Missing build output: $item" }
}

New-Item -ItemType Directory -Path $portable | Out-Null
foreach ($item in @("FuckEnglish.exe", "BrimDeck-themes", "licenses", "icon.png", "NOTICE-BrimDeck.txt")) {
  Copy-Item -LiteralPath (Join-Path $stage $item) -Destination $portable -Recurse
}
foreach ($item in @("README.md", "LICENSE")) {
  Copy-Item -LiteralPath (Join-Path $root $item) -Destination $portable
}
New-Item -ItemType Directory -Path (Join-Path $portable "words"), (Join-Path $portable "docs") | Out-Null
foreach ($book in $wordbooks) { Copy-Item -LiteralPath $book.FullName -Destination (Join-Path $portable "words") }
foreach ($item in @("CHANGES.md", "VERIFICATION.md")) {
  Copy-Item -LiteralPath (Join-Path $root "docs\$item") -Destination (Join-Path $portable "docs")
}

$instructions = @"
FuckEnglish $version · Windows x64 免安装版

请先退出旧版，再双击 FuckEnglish.exe，无需安装。保留整个文件夹。
本包 words 中包含 $($wordbooks.Count) 份词表。首次运行默认使用 words；已有设置会继续沿用原词表路径，不自动切换或覆盖。
如需使用本包词表，在设置的“词表文件夹”页选择本包 words 文件夹，点击“应用文件夹”。修改 Excel 后点击“刷新词表”。

顶部选择“浏览 / 英选中 / 中选英”。浏览底部可以切换顺序 / 乱序，每页 5 词，两种顺序分别记住页码。
小测使用当前整份词表，每题四选一，题目与选项随机排序。选择后立即显示对错，点击下一题，结束查看本轮答对数。
小测只读取每个词自己填写的三个中文或英文干扰项。缺失、重复或与正确答案相同的选项会使该词跳过对应方向小测，仍能浏览。
收起后保留当前题目；切换词表重新出题，刷新结束当前轮；退出软件后不保留本轮成绩。
设置的“单词模板”页可保存含十列表头的空白 Excel 模板。

本包和安装版使用同一设置目录 %APPDATA%\FuckEnglish。关闭主窗口会继续运行，完全退出请用托盘菜单。
安装版更新沿用旧安装位置和设置，请先退出旧版再运行安装器；安装器不覆盖用户词表。
更多说明和来源信息见 README.md。
"@
$instructions | Set-Content -LiteralPath (Join-Path $portable "使用说明.txt") -Encoding utf8
Copy-Item -LiteralPath (Join-Path $stage $installerName) -Destination $destination

function Write-Checksums([string]$Directory) {
  $files = Get-ChildItem -LiteralPath $Directory -File -Recurse | Where-Object { $_.Name -ne "SHA256SUMS.txt" } | Sort-Object FullName
  $lines = foreach ($file in $files) {
    $relative = $file.FullName.Substring($Directory.Length + 1).Replace('\', '/')
    "{0}  {1}" -f (Get-FileHash -LiteralPath $file.FullName -Algorithm SHA256).Hash.ToLowerInvariant(), $relative
  }
  $lines | Set-Content -LiteralPath (Join-Path $Directory "SHA256SUMS.txt") -Encoding utf8
}
Write-Checksums $portable
Compress-Archive -LiteralPath $portable -DestinationPath $zipPath -CompressionLevel Optimal
Write-Checksums $destination
Write-Output "Installer: $(Join-Path $destination $installerName)"
Write-Output "Portable directory: $portable"
Write-Output "Portable ZIP: $zipPath"
