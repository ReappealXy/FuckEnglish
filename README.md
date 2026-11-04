# FuckEnglish

一个常驻桌面顶部的胶囊式背单词工具，让你在工作间隙随手记几个单词。

沿用已有灵动岛的尺寸、主题和开合交互。鼠标悬浮展开词表，查看英文、音标和中文释义，用上一页、下一页翻阅。

![收起的单词胶囊](docs/images/capsule.png)

![展开后的分页词表](docs/images/wordbook.png)

## 功能

- 顶部居中的常驻胶囊，收起默认 196 × 32，展开默认 520 × 200（逻辑像素）。
- 保留深色、浅色、跟随系统及八套原有主题，支持胶囊、刘海和指示条。
- 读取本地 `.xlsx` 词表，每页 4 个单词，显示英文、音标和中文释义。
- 胶囊顶部切换不同 Excel，自动记住各词表的页码。
- 点击英文，将该词显示在收起的胶囊中；翻页后默认显示本页第一个词。
- 自定义词表文件夹，支持浏览、手动填写、打开文件夹和刷新。
- 保留悬浮展开、移开延迟收起、点击展开、Esc 收起及动画设置。
- 关闭主窗口后继续在托盘运行，完全退出使用托盘菜单。

程序不依赖 Clash / VPN，不采集网络流量，不含云端同步、账号或自动更新。词表读取和设置保存均在本地完成。系统 WebView2 运行时的网络行为由系统管理。

## Windows 使用

1. 打开完整的免安装文件夹，双击 `FuckEnglish.exe`，无需先运行安装包。
2. 默认词表位于程序旁的 `words` 文件夹，内含两份各 50 词的示例 Excel。首次启动也可从内嵌示例生成词表；默认位置不可写时回退到用户数据目录。
3. 在设置的“词表文件夹”页选择其他目录，然后点击“应用文件夹”。添加、删除或编辑 Excel 后点击“刷新词表”。
4. 鼠标移到屏幕顶部胶囊上展开，使用顶部下拉框切换词表，底部按钮翻页。较长的中文释义可在对应单元格内滚动阅读。
5. 通过齿轮按钮打开设置，修改主题或交互后点击“保存外观”。修改外观不会重置词表选择与页码。

Windows 需要 Microsoft Edge WebView2 Runtime，Windows 11 通常已预装。当前已构建 Windows x64；其他平台尚未验证。

## Excel 格式

读取每个 `.xlsx` 文件的**第一个工作表**。推荐表头：

| 英文单词 | 音标 | 英文单词释义 |
| --- | --- | --- |
| pioneer | [ˌpaɪəˈnɪə(r)] | n. 开拓者；先驱者 |

支持列顺序调整，也接受“单词 / 英文”“中文释义 / 中文翻译 / 释义 / 翻译”等常见表头。英文和释义不能为空，缺失的音标显示为 `—`。空行会跳过，Excel 临时锁文件 `~$*.xlsx` 会忽略。第四列短语暂不展示，词表原文不会被软件改写。

暂不支持旧式 `.xls`、加密 Excel、在线表格和自动补全音标；单份词表最多 50000 词，解压后的 Excel 内容最多 64 MB。格式错误会在设置页标明，不会阻止其他有效词表使用。

示例词表由项目提供，释义按原文展示，尚未进行专业校订。

## 设置保存

Windows 通常位于 `%APPDATA%\FuckEnglish\settings.json`，保存外观、词表目录、当前词表及各词表页码。该软件使用独立应用标识和数据目录，卸载原流量软件不影响使用。卸载本程序时请先自行备份词表和设置。

## 开发与构建

需要 Go 1.27.1+、Bun 和 Windows WebView2 Runtime。

```sh
bun install --frozen-lockfile
bun run generate
bun run dev
```

检查及 Windows 构建：

```sh
go test ./...
go vet ./...
bun run typecheck
bun run test:web
bun run build
```

构建产物位于 `build/windows-amd64/`。前端为 Preact / TypeScript，桌面宿主为 Go / MyGo，Excel 读取使用 Go 标准库，不需要安装 Excel 或 Python。

## 来源与许可

本项目由已有的 Clash Traffic 本地灵动岛改造版拆分而来，保留其 Go / MyGo 宿主和 Preact 胶囊实现，词表功能独立实现。

- 原项目：[lufficc/clash-traffic](https://github.com/lufficc/clash-traffic)。当前本地快照和上游 GitHub license 接口未发现项目级许可证；公开分发继承部分前，应进一步核实作者授权。不得将本项目 MIT 声明理解为替上游新增授权。
- 灵动岛外观、交互和主题来源：[MoFeng2223/BrimDeck](https://github.com/MoFeng2223/BrimDeck)，遵循 Apache-2.0。许可证、主题来源和作者信息保留于 `licenses/` 及 `resources/`。
- 本项目自行新增代码采用根目录 `LICENSE` 的 MIT 许可；第三方代码和资源遵循各自许可。

改造说明、每阶段 Git 提交文案见 [docs/CHANGES.md](docs/CHANGES.md)。
