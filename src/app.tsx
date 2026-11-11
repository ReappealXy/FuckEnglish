import { useEffect, useRef, useState } from "preact/hooks";
import { WindowTitlebar } from "./components/window-titlebar";
import { events, Library, Preferences, type LibraryState, type Settings } from "./mygo";
import { IslandSurface } from "./components/island-surface";
import { islandTheme, islandThemes } from "./lib/island-theme";
import { paginateBooks, selectedBookPage } from "./lib/book-pagination";
import "./island.css";
import "./pages/island-settings.css";

const tabs = ["词表文件夹", "单词模板", "常规", "收起样式", "展开外观", "面板主题", "悬浮与动画"];

export function App() {
  const [library, setLibrary] = useState<LibraryState | null>(null);
  const [saved, setSaved] = useState<Settings | null>(null);
  const [form, setForm] = useState<Settings | null>(null);
  const [folder, setFolder] = useState("");
  const [tab, setTab] = useState(tabs[0]);
  const [expanded, setExpanded] = useState(true);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [failed, setFailed] = useState(false);
  const [bookPage, setBookPage] = useState(0);
  const savedRef = useRef<Settings | null>(null);
  const books = library?.books ?? [];
  const bookList = paginateBooks(books, bookPage);
  useEffect(() => { setBookPage(selectedBookPage(books, library?.selected ?? "")); }, [library?.folder, library?.selected]);
  useEffect(() => { setBookPage(page => paginateBooks(books, page).page); }, [books.length]);
  useEffect(() => {
    let active = true, gotLibrary = false, gotSettings = false;
    const offLibrary = events.library.on(s => { gotLibrary = true; setLibrary(s); });
    const offSettings = events.settings.on(s => { gotSettings = true; const previous = savedRef.current; setSaved(s); setForm(old => old && previous ? {...s, ...Object.fromEntries(Object.keys(old).filter(key => old[key as keyof Settings] !== previous[key as keyof Settings]).map(key => [key, old[key as keyof Settings]]))} : s); savedRef.current = s; });
    const offNavigate = events.islandNavigate.on(() => setTab("词表文件夹"));
    Library.snapshot().then(s => { if (active && !gotLibrary) { setLibrary(s); setFolder(s.folder); } }).catch(e => { setMessage(String(e)); setFailed(true); });
    Preferences.get().then(s => { if (active && !gotSettings) { setSaved(s); setForm(s); savedRef.current = s; } }).catch(e => { setMessage(String(e)); setFailed(true); });
    return () => { active = false; offLibrary(); offSettings(); offNavigate(); };
  }, []);
  const set = (patch: Partial<Settings>) => setForm(old => old ? {...old, ...patch} : old);
  const run = async (action: () => Promise<unknown>, success = "") => {
    setBusy(true); setMessage(""); setFailed(false);
    try { await action(); setMessage(success); } catch (e) { setMessage(String(e)); setFailed(true); }
    finally { setBusy(false); }
  };
  const dirty = JSON.stringify(form) !== JSON.stringify(saved);
  const save = () => run(async () => { if (!form) return; const s = await Preferences.save(form); savedRef.current = s; setSaved(s); setForm(s); }, "设置已保存");
  return <div class="app-shell">
    <WindowTitlebar subtitle="摸鱼背单词" />
    <main class="app-main"><div class="deck-settings deck-palette">
      <nav class="deck-settings-nav" aria-label="软件设置"><h2>单词胶囊</h2>{tabs.map(label => <button type="button" aria-current={tab === label ? "page" : undefined} class={tab === label ? "selected" : ""} onClick={() => { setTab(label); setExpanded(label !== "收起样式"); }}>{label}</button>)}<p class="app-version">FuckEnglish 0.1.3<br/>本地词表 · 随手翻阅</p></nav>
      <div class="deck-settings-content">
        <div class="deck-preview-heading"><span>实时词表预览</span><div><button type="button" class={!expanded ? "selected" : ""} onClick={() => setExpanded(false)}>收起</button><button type="button" class={expanded ? "selected" : ""} onClick={() => setExpanded(true)}>展开</button></div></div>
        <div class="deck-preview" style={{height: `${expanded ? (form?.islandHeight ?? 200)+16 : 70}px`}}><div style={{width:"100%",maxWidth:`${(form?.islandWidth ?? 520)+24}px`,height:`${(form?.islandHeight ?? 200)+16}px`,margin:"0 auto"}}><IslandSurface preview settings={form} library={library} expanded={expanded} busy={busy} onSelect={name => void run(() => Library.select(name))} onPage={page => void run(() => Library.page(page))} onFocus={index => void run(() => Library.focus(index))} onMode={mode => void run(() => mode === "browse" ? Library.endQuiz() : Library.startQuiz(mode))} onOrder={order => void run(() => Library.setBrowseOrder(order))} onAnswer={(token, choice) => void run(() => Library.answerQuiz(token, choice))} onNext={token => void run(() => Library.nextQuiz(token))} /></div></div>
        <section class="deck-settings-section"><h3>{tab}</h3>
          {tab === "词表文件夹" && <>
            <p class="section-description">将 .xlsx 词表放入同一个文件夹，在胶囊顶部选择词表。添加或修改文件后点击刷新。</p>
            <label class="folder-label" for="word-folder">Excel 文件夹</label><div class="folder-controls"><input id="word-folder" value={folder} onInput={e => setFolder(e.currentTarget.value)} spellcheck={false}/><button disabled={busy} onClick={() => void run(async () => { const path = await Library.pickFolder(); if (path) setFolder(path); })}>浏览…</button><button class="primary" disabled={busy || !folder.trim() || folder === library?.folder} onClick={() => void run(async () => { const s=await Library.setFolder(folder.trim()); setLibrary(s); setFolder(s.folder); }, "词表文件夹已切换")}>应用文件夹</button></div>
            <div class="library-toolbar"><span>{books.length} 份 Excel · 每页 5 份词表</span><div><button disabled={busy} onClick={() => void run(() => Library.openFolder())}>打开文件夹</button><button disabled={busy} onClick={() => void run(() => Library.refresh(), "词表已刷新")}>刷新词表</button></div></div>
            {library?.error && <p class="feedback error" role="alert">{library.error}</p>}
            <div class="book-list">{bookList.books.map(book => <button key={book.name} class={`book-item${book.name === library?.selected ? " selected" : ""}`} disabled={busy || !!book.error} onClick={() => void run(() => Library.select(book.name))}><span><strong>{book.name}</strong>{book.error && <small>{book.error}</small>}</span><span>{book.error ? "无法读取" : `${book.count} 词${book.name === library?.selected ? " · 使用中" : ""}`}</span></button>)}{library && books.length === 0 && <p class="empty-books">这里还没有 .xlsx 词表。把文件放入上方文件夹，再点击“刷新词表”。</p>}</div>
            {bookList.pages > 0 && <nav class="book-list-pagination" aria-label="设置词表分页"><span>共 {books.length} 份</span><button type="button" disabled={busy || bookList.page === 0} onClick={() => setBookPage(bookList.page - 1)}>上一页</button><span>{bookList.page + 1} / {bookList.pages}</span><button type="button" disabled={busy || bookList.page + 1 >= bookList.pages} onClick={() => setBookPage(bookList.page + 1)}>下一页</button></nav>}
            <p class="deck-settings-note">读取第一个工作表，表头使用：英文单词、音标、英文单词释义。第四列短语暂不展示。较长释义可在单元格内滚动阅读；点击英文可将该词显示在收起胶囊中。切换词表会记住各自页码。</p>
          </>}
          {tab === "单词模板" && <div class="word-template"><p>下载十列 Excel 模板，按列填写后，保存到词表文件夹并刷新词表即可浏览。</p><div class="word-template-columns">{["英文单词", "音标", "英文单词释义", "常用短语", "中文干扰项1", "中文干扰项2", "中文干扰项3", "英文干扰项1", "英文干扰项2", "英文干扰项3"].map(label => <span key={label}>{label}</span>)}</div><p>中文干扰项用于“看英文选中文”，英文干扰项用于“看中文选英文”。每组填写三个不同的错误选项，不要填写正确答案。</p><p>缺少该方向三个干扰项，或选项重复的单词仍可浏览，但不能参与该方向小测。在胶囊顶部选择“英选中”或“中选英”开始测试。</p><button class="word-template-save" disabled={busy} onClick={() => void (async () => { setBusy(true); setMessage(""); setFailed(false); try { const path = await Library.saveTemplate(); if (path) setMessage(`模板已保存：${path}`); } catch (e) { setMessage(String(e)); setFailed(true); } finally { setBusy(false); } })()}>保存 Excel 模板…</button></div>}
          {form && tab === "常规" && <><Row label="显示单词胶囊"><Toggle checked={form.islandVisible} set={v => set({islandVisible:v})} label="显示单词胶囊"/></Row><Row label="鼠标悬浮展开"><Toggle checked={form.islandHover} set={v => set({islandHover:v})} label="鼠标悬浮展开"/></Row><Row label="开机启动"><Toggle checked={form.openAtLogin} set={v => set({openAtLogin:v})} label="开机启动"/></Row><p class="deck-settings-note">登录 Windows 后自动在托盘启动胶囊，不打开设置窗口。关闭主窗口后胶囊继续运行；完全退出请使用托盘菜单。</p></>}
          {form && tab === "收起样式" && <Row label="形态"><div class="deck-shape-choices">{([['capsule','胶囊'],['notch','刘海'],['line','指示条']] as const).map(([value,label]) => <button class={form.islandStyle === value ? "selected" : ""} aria-pressed={form.islandStyle === value} onClick={() => set({islandStyle:value})}><span class={`deck-shape deck-shape-${value}`}/>{label}</button>)}</div></Row>}
          {form && tab === "展开外观" && <><Row label="面板宽度"><Slider value={form.islandWidth} min={440} max={1200} step={20} unit="px" set={v => set({islandWidth:v})}/></Row><Row label="面板高度"><Slider value={form.islandHeight} min={200} max={600} step={20} unit="px" set={v => set({islandHeight:v})}/></Row><p class="deck-settings-note">沿用原版默认 520 × 200，收起胶囊保持 196 × 32。每页固定 5 词。</p></>}
          {form && tab === "面板主题" && <div class="deck-theme-list"><button class={`deck-theme${form.islandTheme === 'system' ? ' selected' : ''}`} aria-pressed={form.islandTheme === 'system'} onClick={() => set({islandTheme:'system'})}><span class="deck-system-theme">◐</span><span>跟随系统</span></button>{islandThemes.map(theme => <button class={`deck-theme${islandTheme(form.islandTheme).id === theme.id && form.islandTheme !== 'system' ? ' selected' : ''}`} aria-pressed={form.islandTheme === theme.id} onClick={() => set({islandTheme:theme.id})}><span class="deck-theme-sample" style={{background:theme.tokens.surface,color:theme.tokens.text}}><i style={{background:theme.tokens.down}}/><i style={{background:theme.tokens.up}}/><b>Aa</b></span><span>{theme.name}</span></button>)}</div>}
          {form && tab === "悬浮与动画" && <><Row label="悬浮展开延迟"><Slider value={form.islandExpandDelay} min={0} max={1200} step={20} unit="ms" set={v=>set({islandExpandDelay:v})}/></Row><Row label="离开收起延迟"><Slider value={form.islandCollapseDelay} min={120} max={2000} step={20} unit="ms" set={v=>set({islandCollapseDelay:v})}/></Row><Row label="动画时长"><Slider value={form.islandAnimationDuration} min={100} max={1000} step={20} unit="ms" set={v=>set({islandAnimationDuration:v})}/></Row><p class="deck-settings-note">保留原版开合动画，并遵循系统减少动态效果设置。</p></>}
        </section>
        <footer class="deck-settings-actions"><span role={failed ? "alert" : "status"} class={failed ? "error-text" : ""}>{message || (dirty ? "设置有未保存修改" : "设置已保存")}</span><button disabled={busy || !dirty} onClick={() => setForm(saved)}>还原</button><button class="deck-save" disabled={busy || !dirty} onClick={() => void save()}>{busy ? "处理中…" : "保存设置"}</button></footer>
      </div>
    </div></main>
  </div>;
}

function Row({label,children}:{label:string;children:preact.ComponentChildren}) {return <div class="deck-setting-row"><div>{label}</div><div class="deck-setting-control">{children}</div></div>;}
function Toggle({checked,set,label}:{checked:boolean;set:(v:boolean)=>void;label:string}){return <label class="deck-toggle"><input type="checkbox" aria-label={label} checked={checked} onChange={e=>set(e.currentTarget.checked)}/><span/></label>;}
function Slider({value,min,max,step,unit,set}:{value:number;min:number;max:number;step:number;unit:string;set:(v:number)=>void}){return <label class="deck-slider"><input type="range" aria-label={`数值 ${value} ${unit}`} min={min} max={max} step={step} value={value} onInput={e=>set(Number(e.currentTarget.value))}/><output>{value} {unit}</output></label>;}
