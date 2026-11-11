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
Copy-Item -LiteralPath (Join-Path $root "docs\images") -Destination (Join-Path $portable "docs") -Recurse
Copy-Item -LiteralPath (Join-Path $root "releases") -Destination $portable -Recurse

$instructions = @"
FuckEnglish $version · Windows x64 免安装版

请先退出旧版，再双击 FuckEnglish.exe，无需安装。保留整个文件夹。
本包 words 中包含 $($wordbooks.Count) 份词表。首次运行默认使用 words；已有设置会继续沿用原词表路径，不自动切换或覆盖。
如需使用本包词表，在设置的“词表文件夹”页选择本包 words 文件夹，点击“应用文件夹”。修改 Excel 后点击“刷新词表”。

顶部选择“浏览 / 英选中 / 中选英”。浏览底部可以切换顺序 / 乱序，每页 5 词，两种顺序分别记住页码。
词表按数字自然顺序排列，list9 后是 list10。设置页和胶囊词表面板每页 5 份；胶囊中点击词表名称打开面板，翻页后点击词表即可切换并关闭面板。
浏览时将鼠标停留在展开的胶囊内，可按键盘 ← 上一页、→ 下一页。移开或收起后停用，小测不使用这组翻页键。
操作词表面板或模式下拉框时不翻动单词；点击词表正文后可继续快捷翻页。
胶囊展开后点击“背词”，切换到可调大小、最大化的普通窗口；最小化、关闭或点击“回到胶囊”都会返回胶囊，进度保持同步。
胶囊和普通窗口小测均采用居中题目，英文题目下方显示词表音标；四个答案按两行两列排列，无编号，鼠标移入显示浅框，点击立即作答，无需确认。
小测使用当前整份词表，每题四选一，题目与选项随机排序。答对后短暂显示反馈并自动下一题；答错后点击下一题。最后一题无论对错，点击“完成”才查看本轮答对数。
小测只读取每个词自己填写的三个中文或英文干扰项。缺失、重复或与正确答案相同的选项会使该词跳过对应方向小测，仍能浏览。
收起后保留当前题目；切换词表重新出题，刷新结束当前轮；退出软件后不保留本轮成绩。
设置的“单词模板”页可保存含十列表头的空白 Excel 模板。
设置的“面板主题”页可分别调节背景透明度和文字透明度：0% 不透明，100% 完全透明。保存后胶囊与背词窗口同步，切换主题或重启后保留；设置窗口本身不透明，方便随时调回。

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
