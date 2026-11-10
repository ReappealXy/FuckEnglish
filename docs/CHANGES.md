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

## 6. Windows 开机启动

Summary: `feat: add Windows login startup setting`

Description: 在常规设置增加开机启动开关，调用 MyGo 系统登录启动接口并在设置保存时应用；由登录启动时静默进入托盘和单词胶囊，不显示主设置窗口。

大白话：勾选并保存后，登录 Windows 就会自动运行单词胶囊；关闭开关后会移除启动项。

## 7. 原位置覆盖更新

Summary: `feat: update existing installation in place`

Description: 安装器从旧版卸载记录读取自定义安装目录，更新时沿用该目录；稳定应用名称和登录启动标识，沿用现有设置文件及自定义词表路径。安装程序和免安装程序均移除 Preview 产品名称。

大白话：新版安装器找到旧版本装在哪儿，就在原位置更新，设置和 Excel 文件夹继续用原来的。

## 3. 完成文件夹和外观设置

Summary: `feat: configure wordbook folders and preserve island appearance`

Description: 增加词表文件夹浏览、手动输入、应用、打开和刷新；内置两份真实测试词表，默认使用程序旁 words。原八套主题、系统主题、形态、尺寸、悬浮延迟及动画设置保持可用，预览与真实胶囊共享内容组件。外观保存只修改外观字段，不覆盖词表与页码。保留完整释义并允许在单元格内滚动。

大白话：Excel 放哪里由你决定，原来的亮暗主题和开合设置照样能用。

## 4. 验证与 Windows 交付

Summary: `test: verify vocabulary workflows and package Windows app`

Description: 添加真实词表解析、Excel 格式错误、共享字符串、分页边界、目录切换和保存失败回滚测试；保留悬浮计时及原主题测试。执行类型检查、Go 测试、go vet 和原生桌面界面验证，输出 Windows EXE、可选安装包、README、来源说明和免安装交付目录。具体实测范围见 VERIFICATION.md。

大白话：检查词表能读、分页能用、设置能保存，再给你双击运行的程序。

## 8. 为个人词表补齐干扰项字段

Summary: `chore: extend personal wordbooks with distractor columns`

Description: 原地更新 D:\Study\English\list 中 list1.xlsx、list2.xlsx、list3.xlsx，在原四列后追加中文干扰项1～3、英文干扰项1～3，新增内容留空。修改前备份原文件，逐单元格核对原有数据，并保留其余 XLSX 部件、样式和工作表属性。三份个人词表位于仓库外，不包含在 Git 提交中。

大白话：三份表都加好了以后填错误选项的位置，原来的单词和释义没有动。

## 9. 更新下载模板与填写说明

Summary: `feat: add bilingual distractor columns to Excel template`

Description: 软件导出的空白 Excel 模板从四列扩展到十列，新增中英文各三个干扰项；同步设置页字段展示、填写说明和 README，扩展模板结构检查。旧四列词表及新增空列的词表继续正常浏览。记录后续测试规则：使用每行自行填写的干扰项，缺少对应方向三个干扰项的词不能参与该方向测试；测试与顺序 / 乱序切换尚未实现。

大白话：以后直接按新模板生成词表就行。中文错误选项和英文错误选项分开填，没填也不影响看单词。

## 10. 双向单词小测

Summary: `feat: add bilingual vocabulary quizzes`

Description: 胶囊顶部新增浏览、英选中、中选英切换。读取每行自备的中英文干扰项，整份当前词表出题，每题四选一，题目及选项默认乱序。缺失、重复或与正确答案相同的干扰项仅使该词跳过对应方向小测。答题后锁定本题并立即显示对错与正确答案，点击下一题继续，末题查看本轮得分，支持重新测试与返回浏览。收起保留当前题目，切换词表按当前方向重新开始。题目令牌阻止迟到请求或重复点击跨题计分。

大白话：现在可以选答案检查自己记没记住，选错会告诉你正确答案，测完看答对了多少个。错误选项只用你在 Excel 里准备的内容。

## 11. 顺序与乱序浏览

Summary: `feat: add persistent shuffled wordbook browsing`

Description: 浏览底栏新增顺序 / 乱序切换。每份词表使用稳定的随机排列，翻页不重复洗牌，重启及切换词表继续沿用原排列；分别保存两种浏览顺序的页码。保留旧设置、词表路径和原顺序页码。修正五词分页底栏仍按四词计算的编号，并补充解析、题目资格、计分、题目随机性、分页稳定性和保存失败回滚检查。

大白话：既能按 Excel 原顺序背，也能打乱背，来回翻页不会漏词或乱跳，两种模式各记各的位置。

## 12. 补齐三份个人词表的小测内容

Summary: `docs: record completed vocabulary quiz data`

Description: 为 D:\Study\English\list 中三份个人词表各 50 词补齐中文与英文各三个干扰项，共 900 个新增单元格。逐词编写并复核多义词、近义词及英美拼写变体，避免把同义答案设为干扰项。保留原有单词、音标、释义、短语与工作簿属性，修改前完整备份。实际小测服务验证六轮均为 50 题、跳过 0 词并可答完得到 50/50。个人词表位于仓库外，本提交仅包含交付和验证记录。

大白话：三个 list 的错误选项都填好了，不再只有空列；刷新词表后两个方向都能各测 50 个词。

## 13. 真实词表验收与 Windows 打包

Summary: `build: verify quizzes and package Windows installer and portable app`

Description: 使用三份已补齐的真实词表与正式构建前端连接实际 Go 服务，在外部 Chrome 中逐题完成六轮共 300 题，每轮故意答错一题，验证最终 49/50。检查四个选项来源、即时反馈、乱序、浏览翻页、重启恢复、刷新与旧词表兼容。修正设置页版本为 0.1.3；重新编译原位置更新安装器，增加免安装打包脚本，附带三份完整词表、资源、许可、说明和 SHA256 校验文件，保留旧交付物。未提交、发布或执行安装器覆盖本机安装。

大白话：用你真实的词表完整测了两种题型，答错会扣分、测完分数正确；再把新版安装包和解压即用的免安装包整理好。

## 14. 修复乱序控件导致胶囊收起

Summary: `fix: keep browse order controls inside the island`

Description: 将浏览顺序的原生下拉框改为胶囊内的“顺序 / 乱序”按钮组，避免系统下拉菜单伸出胶囊命中区域后触发悬浮离开并自动收起。保留现有顺序切换和页码逻辑。

大白话：点顺序或乱序时不会再弹出胶囊外的菜单，鼠标移到选项上也不会把胶囊收起来。

## 15. 整理项目文档与界面截图

Summary: `docs: add screenshots and user-facing release notes`

Description: 为 README 增加文档导航和双向小测界面截图；新增面向使用者的 0.1.3 版本说明，记录浏览顺序切换、小测规则、Excel 干扰项填写和原位置更新方式。更新浏览截图以展示五词布局与胶囊内顺序按钮。截图使用通用示例词，不包含个人词表内容。

大白话：README 现在更容易找到说明，新增的版本说明讲清楚怎么浏览、怎么小测和怎么更新；截图展示的是示例词，不是个人词表。

## 16. 打包左右键翻页版本

Summary: `build: package arrow-key paging for Windows`

Description: 构建包含左右键翻页的 Windows x64 程序和原位置更新安装器，生成附带 list1、list2、list3 的免安装目录及 ZIP。补全便携包的截图与版本说明资源，更新快捷键使用说明，并校验构建产物、词表与压缩包完整性。保持 0.1.3 版本及现有应用标识，保留此前交付包，未运行安装器或发布。

大白话：这次安装版和免安装版都带上左右键翻页，解压后的说明、截图也能直接查看。

## 17. 新增普通窗口背词模式

Summary: `feat: add resizable study window and fix titlebar overlap`

Description: 胶囊展开栏新增“背词”按钮，打开沿用设置页样式的独立背词窗口，支持调整大小、最大化和恢复；最小化、关闭或点击“回到胶囊”均返回胶囊。两种模式共用词表、浏览顺序、页码和本轮小测状态，保留左右键浏览翻页。抽取共用学习内容与设置页配色，移除叠在系统标题栏按钮下的重复自绘按钮，修复最大化图标多余横线。补充使用说明与验证记录。

大白话：摸鱼时用胶囊，想认真背时点“背词”放大成普通窗口；窗口大小随你调，关掉或最小化就回胶囊，背到哪、测到哪都接着来。右上角按钮也不再叠在一起。

## 18. 调整背词窗口小测界面并打包

Summary: `fix: simplify study quiz layout and package Windows builds`

Description: 普通背词窗口小测改为居中大题目，英选中显示当前单词音标；四个答案保持两行两列，缩小区域并增加间距，默认隐藏边框、悬浮或键盘聚焦时显示浅框，去掉 ABCD 编号。点击立即判题并显示正确答案，无需确认，保留下一题和胶囊原有布局。验证双向小测、音标对应关系、亮暗主题和尺寸适配，生成 Windows 安装 EXE 与附带三份词表的便携 ZIP。

大白话：单词变大、音标补上，答案不再占满整个窗口，鼠标移上去才显示框；点哪个就选哪个。安装版和免安装版一起更新，方便直接测试。
