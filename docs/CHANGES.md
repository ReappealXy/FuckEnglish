# Git 提交文案

以下为分阶段可复制的文案，未自动执行提交。

## 1. 拆分独立桌面胶囊

Summary: `refactor: extract standalone vocabulary capsule host`

Description: 从现有本地灵动岛迁入透明宿主、固定窗口、命中检测、顶部定位、阴影分层与开合时序。保留主题和设置页视觉体系，改用 FuckEnglish 独立名称、应用标识、数据目录及图标。移除流量采集、Clash 连接、网络统计和无关服务依赖，保留托盘驻留及主窗口关闭隐藏行为。

大白话：保留原胶囊的壳和操作，软件独立运行，不需要 VPN 软件。

## 2. 读取本地 Excel 并分页

Summary: `feat: add local Excel wordbooks and paginated vocabulary`

Description: 使用 Go 标准库读取 XLSX 首个工作表，支持共享字符串、内联文字、富文本、稀疏列和常见表头。展示英文、音标、中文释义，每页 5 词，支持末页、词表切换、各词表页码持久化及点击单词切换紧凑胶囊内容。错误文件单独标注，处理空文件夹和文件被删除等情况，暂不展示短语。

大白话：一份 Excel 就是一本词表，翻页看词，切回来还能接着上次的位置。

## 5. 每页五词与 Excel 模板

Summary: `feat: show five words per page and add Excel template`

Description: 每页展示五个单词，设置页新增 Excel 模板保存入口，模板包含英文、音标、中文释义和常用短语四列；安装包使用独立版本标识，旧版安装与免安装交付保留。

大白话：一页多看一个词，还能从设置里保存一份标准单词表格。

## 3. 完成文件夹和外观设置

Summary: `feat: configure wordbook folders and preserve island appearance`

Description: 增加词表文件夹浏览、手动输入、应用、打开和刷新；内置两份真实测试词表，默认使用程序旁 words。原八套主题、系统主题、形态、尺寸、悬浮延迟及动画设置保持可用，预览与真实胶囊共享内容组件。外观保存只修改外观字段，不覆盖词表与页码。保留完整释义并允许在单元格内滚动。

大白话：Excel 放哪里由你决定，原来的亮暗主题和开合设置照样能用。

## 4. 验证与 Windows 交付

Summary: `test: verify vocabulary workflows and package Windows app`

Description: 添加真实词表解析、Excel 格式错误、共享字符串、分页边界、目录切换和保存失败回滚测试；保留悬浮计时及原主题测试。执行类型检查、Go 测试、go vet 和原生桌面界面验证，输出 Windows EXE、可选安装包、README、来源说明和免安装交付目录。具体实测范围见 VERIFICATION.md。

大白话：检查词表能读、分页能用、设置能保存，再给你双击运行的程序。
