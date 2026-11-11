import { useEffect, useRef, useState } from "preact/hooks";
import type { Settings, LibraryState } from "../mygo";
import { islandThemeStyle } from "../lib/island-theme";
import { LearningContent } from "./learning-content";
import { paginateBooks, selectedBookPage } from "../lib/book-pagination";
import { transparencyStyle } from "../lib/learning-appearance";

type Props = {
  settings: Partial<Settings> | null; library: LibraryState | null; expanded: boolean; error?: string; preview?: boolean;
  onEnter?: () => void; onLeave?: () => void; onExpand?: () => void; onCollapse?: () => void; onSettings?: () => void;
  onStudy?: () => void;
  onSelect?: (name: string) => void; onPage?: (page: number) => void; onFocus?: (index: number) => void; busy?: boolean;
  onMode?: (mode: string) => void; onOrder?: (order: string) => void;
  onAnswer?: (token: string, choice: number) => void; onNext?: (token: string) => void;
};

export function IslandSurface(props: Props) {
  const { settings, library, expanded, preview, busy } = props;
  const [bookPickerOpen, setBookPickerOpen] = useState(false);
  const [bookPage, setBookPage] = useState(0);
  const bookTrigger = useRef<HTMLButtonElement>(null);
  const books = library?.books ?? [];
  const bookList = paginateBooks(books, bookPage);
  const selectedBook = books.find(book => book.name === library?.selected);
  const pickerId = preview ? "preview-book-picker" : "island-book-picker";
  useEffect(() => { if (!expanded) setBookPickerOpen(false); }, [expanded]);
  useEffect(() => { setBookPickerOpen(false); }, [library?.folder, library?.selected]);
  useEffect(() => { setBookPage(page => paginateBooks(books, page).page); }, [books.length]);
  const [systemDark, setSystemDark] = useState(() => matchMedia("(prefers-color-scheme: dark)").matches);
  useEffect(() => { const media = matchMedia("(prefers-color-scheme: dark)"); const update = () => setSystemDark(media.matches); media.addEventListener("change", update); return () => media.removeEventListener("change", update); }, []);
  const shape = settings?.islandStyle === "notch" || settings?.islandStyle === "line" ? settings.islandStyle : "capsule";
  const words = library?.words ?? [];
  const current = words[library?.focus ?? 0] ?? words[0];
  const quiz = library?.quiz;
  const error = props.error || library?.error;
  return <div class={`island-stage ${preview ? "island-preview-stage" : ""}`} style={{ ...islandThemeStyle(settings?.islandTheme ?? "dark", systemDark), ...transparencyStyle(settings), "--island-duration": `${settings?.islandAnimationDuration ?? 360}ms` }}>
    <div class={`island-surface style-${shape} ${expanded ? "expanded" : "collapsed"}`} onPointerEnter={props.onEnter} onPointerLeave={props.onLeave}>
      <svg class="island-outline" width="0" height="0" aria-hidden="true"><defs><clipPath id={preview ? "preview-notch" : "word-notch"} clipPathUnits="objectBoundingBox"><path d="M0 0 H1 C.981 0 .957 0 .945 .039 C.933 .079 .933 .158 .933 .219 V.625 Q.933 1 .882 1 H.118 Q.067 1 .067 .625 V.219 C.067 .158 .067 .079 .055 .039 C.043 0 .019 0 0 0 Z" /></clipPath></defs></svg>
      <div class="island-background" aria-hidden="true">
        <div class="island-shadow"><div class="island-shadow-ink" style={!expanded && shape === "notch" ? { clipPath: `url(#${preview ? "preview-notch" : "word-notch"})` } : undefined} /></div>
        <div class="island-material" style={!expanded && shape === "notch" ? { clipPath: `url(#${preview ? "preview-notch" : "word-notch"})` } : undefined} />
      </div>
      <button type="button" class="island-compact" onClick={props.onExpand} disabled={preview || expanded} tabIndex={expanded ? -1 : 0} aria-label="展开单词胶囊" aria-expanded={expanded}>
        {shape !== "line" && <span class={current ? "compact-word" : "compact-empty"}>{current?.english ?? (library ? "选择一个词表" : "正在读取词表…")}</span>}
      </button>
      <section class="island-expanded-content" aria-label="单词词表" aria-hidden={!expanded} inert={!expanded} onKeyDown={event => {
        if (event.key === "Escape" && bookPickerOpen) { event.preventDefault(); event.stopPropagation(); setBookPickerOpen(false); bookTrigger.current?.focus(); }
      }}>
        <header class="island-toolbar">
          <button ref={bookTrigger} type="button" class="wordbook-select" data-book-picker aria-label="切换 Excel 词表" aria-haspopup="dialog" aria-expanded={bookPickerOpen} aria-controls={pickerId} disabled={busy || !books.length} onClick={() => {
            if (!bookPickerOpen) setBookPage(selectedBookPage(books, library?.selected ?? ""));
            setBookPickerOpen(!bookPickerOpen);
          }}><span>{selectedBook ? `${selectedBook.name.replace(/\.xlsx$/i, "")} · ${selectedBook.count} 词` : "暂无词表"}</span><Glyph path={bookPickerOpen ? "m6 15 6-6 6 6" : "m6 9 6 6 6-6"} /></button>
          <select class="capsule-mode" aria-label="学习模式" value={quiz?.direction || "browse"} disabled={busy || !library?.selected} onChange={e => { setBookPickerOpen(false); props.onMode?.(e.currentTarget.value); }}>
            <option value="browse">浏览</option><option value="english-chinese">英选中</option>
          </select>
          <span class="wordbook-count">{library?.total ?? 0} 个单词</span>
          <div class="island-tools">
            {!preview && <button type="button" class="island-study-button" aria-label="切换背词模式" title="切换背词模式" onClick={props.onStudy}>背词</button>}
            <button type="button" aria-label="打开设置" disabled={preview} onClick={props.onSettings}><Glyph path="M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8m0-6v3m0 14v3M2 12h3m14 0h3M5 5l2 2m10 10 2 2M5 19l2-2M17 7l2-2" /></button>
            <button type="button" aria-label="收起胶囊" disabled={preview} onClick={props.onCollapse}><Glyph path="m6 15 6-6 6 6" /></button>
          </div>
        </header>
        {bookPickerOpen ? <div id={pickerId} class="book-picker" role="dialog" aria-label="选择词表" data-book-picker>
          <div class="book-picker-list">{bookList.books.map(book => <button key={book.name} type="button" class={`book-picker-item${book.name === library?.selected ? " selected" : ""}`} aria-current={book.name === library?.selected ? "true" : undefined} disabled={busy || !!book.error} title={book.error || book.name} onClick={() => {
            setBookPickerOpen(false); bookTrigger.current?.focus(); props.onSelect?.(book.name);
          }}><span>{book.name.replace(/\.xlsx$/i, "")}</span><small>{book.error ? "格式错误" : `${book.count} 词${book.name === library?.selected ? " · 使用中" : ""}`}</small></button>)}</div>
          <nav class="book-picker-pagination" aria-label="胶囊词表分页"><span>共 {books.length} 份</span><button type="button" disabled={busy || bookList.page === 0} onClick={() => setBookPage(bookList.page - 1)}>上一页</button><span>{bookList.pages ? bookList.page + 1 : 0} / {bookList.pages}</span><button type="button" disabled={busy || bookList.page + 1 >= bookList.pages} onClick={() => setBookPage(bookList.page + 1)}>下一页</button></nav>
        </div> : <LearningContent library={library} error={error} busy={busy} onSettings={props.onSettings} onFocus={props.onFocus} onOrder={props.onOrder} onPage={props.onPage} onMode={props.onMode} onAnswer={props.onAnswer} onNext={props.onNext} />}
      </section>
    </div>
  </div>;
}

function Glyph({ path }: { path: string }) { return <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.65" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d={path} /></svg>; }
