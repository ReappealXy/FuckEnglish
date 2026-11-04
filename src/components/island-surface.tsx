import { useEffect, useState } from "preact/hooks";
import type { Settings, LibraryState } from "../mygo";
import { islandThemeStyle } from "../lib/island-theme";

type Props = {
  settings: Partial<Settings> | null; library: LibraryState | null; expanded: boolean; error?: string; preview?: boolean;
  onEnter?: () => void; onLeave?: () => void; onExpand?: () => void; onCollapse?: () => void; onSettings?: () => void;
  onSelect?: (name: string) => void; onPage?: (page: number) => void; onFocus?: (index: number) => void; busy?: boolean;
};

export function IslandSurface(props: Props) {
  const { settings, library, expanded, preview, busy } = props;
  const [systemDark, setSystemDark] = useState(() => matchMedia("(prefers-color-scheme: dark)").matches);
  useEffect(() => { const media = matchMedia("(prefers-color-scheme: dark)"); const update = () => setSystemDark(media.matches); media.addEventListener("change", update); return () => media.removeEventListener("change", update); }, []);
  const shape = settings?.islandStyle === "notch" || settings?.islandStyle === "line" ? settings.islandStyle : "capsule";
  const words = library?.words ?? [];
  const current = words[library?.focus ?? 0] ?? words[0];
  const error = props.error || library?.error;
  return <div class={`island-stage ${preview ? "island-preview-stage" : ""}`} style={{ ...islandThemeStyle(settings?.islandTheme ?? "dark", systemDark), "--island-duration": `${settings?.islandAnimationDuration ?? 360}ms` }}>
    <div class={`island-surface style-${shape} ${expanded ? "expanded" : "collapsed"}`} onPointerEnter={props.onEnter} onPointerLeave={props.onLeave}>
      <svg class="island-outline" width="0" height="0" aria-hidden="true"><defs><clipPath id={preview ? "preview-notch" : "word-notch"} clipPathUnits="objectBoundingBox"><path d="M0 0 H1 C.981 0 .957 0 .945 .039 C.933 .079 .933 .158 .933 .219 V.625 Q.933 1 .882 1 H.118 Q.067 1 .067 .625 V.219 C.067 .158 .067 .079 .055 .039 C.043 0 .019 0 0 0 Z" /></clipPath></defs></svg>
      <div class="island-shadow" aria-hidden="true"><div class="island-shadow-ink" style={!expanded && shape === "notch" ? { clipPath: `url(#${preview ? "preview-notch" : "word-notch"})` } : undefined} /></div>
      <div class="island-material" style={!expanded && shape === "notch" ? { clipPath: `url(#${preview ? "preview-notch" : "word-notch"})` } : undefined} />
      <button type="button" class="island-compact" onClick={props.onExpand} disabled={preview || expanded} tabIndex={expanded ? -1 : 0} aria-label="展开单词胶囊" aria-expanded={expanded}>
        {shape !== "line" && <span class={current ? "compact-word" : "compact-empty"}>{current?.english ?? (library ? "选择一个词表" : "正在读取词表…")}</span>}
      </button>
      <section class="island-expanded-content" aria-label="单词词表" aria-hidden={!expanded} inert={!expanded}>
        <header class="island-toolbar">
          <select class="wordbook-select" aria-label="切换 Excel 词表" value={library?.selected ?? ""} disabled={busy || !library?.books.some(b => !b.error)} onChange={e => props.onSelect?.(e.currentTarget.value)}>
            {!library?.selected && <option value="">暂无词表</option>}
            {library?.books.map(book => <option key={book.name} value={book.name} disabled={!!book.error}>{book.name.replace(/\.xlsx$/i, "")} · {book.count} 词{book.error ? "（格式错误）" : ""}</option>)}
          </select>
          <span class="wordbook-count">{library?.total ?? 0} 个单词</span>
          <div class="island-tools">
            <button type="button" aria-label="打开设置" disabled={preview} onClick={props.onSettings}><Glyph path="M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8m0-6v3m0 14v3M2 12h3m14 0h3M5 5l2 2m10 10 2 2M5 19l2-2M17 7l2-2" /></button>
            <button type="button" aria-label="收起胶囊" disabled={preview} onClick={props.onCollapse}><Glyph path="m6 15 6-6 6 6" /></button>
          </div>
        </header>
        <div class="vocabulary-page">
          {words.length > 0 ? <div class="vocabulary-table" role="table" aria-label="当前页单词">
            <div class="vocabulary-head" role="row"><span role="columnheader">英文单词</span><span role="columnheader">音标</span><span role="columnheader">中文释义</span></div>
            <div class="vocabulary-rows">{words.map((word, index) => <div class={`vocabulary-row${index === library?.focus ? " focused" : ""}`} role="row" key={`${library?.page}-${index}`}>
              <span role="cell"><button class="word-focus" type="button" onClick={() => props.onFocus?.(index)} aria-label={`在胶囊显示 ${word.english}`} aria-pressed={index === library?.focus}>{word.english}</button></span>
              <span class="word-phonetic" role="cell">{word.phonetic || "—"}</span>
              <span class="word-meaning" role="cell" tabIndex={0}>{word.meaning}</span>
            </div>)}</div>
          </div> : <div class="vocabulary-empty">{error ? <><span>{error}</span><button type="button" onClick={props.onSettings}>打开词表设置</button></> : <><span>{library ? "把 Excel 放进词表文件夹，然后点击刷新。" : "正在读取词表…"}</span><button type="button" onClick={props.onSettings}>选择词表文件夹</button></>}</div>}
        </div>
        <footer class="island-statusbar vocabulary-footer">
          <span class={`island-status ${words.length ? "connected" : ""}`} aria-live="polite"><i />{error ? "读取失败 · 请打开设置查看" : words.length ? `第 ${(library?.page ?? 0)*4+1}–${(library?.page ?? 0)*4+words.length} 词` : "等待词表"}</span>
          <div class="word-pagination"><button type="button" disabled={busy || !library || library.page <= 0} onClick={() => props.onPage?.((library?.page ?? 0)-1)}>上一页</button><span>{library?.pages ? `${library.page+1} / ${library.pages}` : "0 / 0"}</span><button type="button" disabled={busy || !library || library.page+1 >= library.pages} onClick={() => props.onPage?.((library?.page ?? 0)+1)}>下一页</button></div>
        </footer>
      </section>
    </div>
  </div>;
}

function Glyph({ path }: { path: string }) { return <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.65" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d={path} /></svg>; }
