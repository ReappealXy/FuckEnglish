import { useEffect, useRef, useState } from "preact/hooks";
import { currentWindow, isMyGo } from "mygo-runtime";
import { events, Island, Library, Preferences, type LibraryState, type Settings } from "./mygo";
import { browsePageTarget, isPagingControl } from "./lib/island-pagination";
import { LearningContent } from "./components/learning-content";
import { WindowTitlebar } from "./components/window-titlebar";
import { transparencyStyle } from "./lib/learning-appearance";
import "./island.css";
import "./study.css";

export function StudyView() {
  const [library, setLibrary] = useState<LibraryState | null>(null);
  const [settings, setSettings] = useState<Settings | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [maximized, setMaximized] = useState(false);
  const requestBusy = useRef(false);

  useEffect(() => {
    if (!isMyGo()) return;
    let revision = 0;
    const update = () => {
      const request = ++revision;
      void currentWindow.isMaximized().then(value => {
        if (request === revision) setMaximized(value);
      }).catch(() => {});
    };
    update();
    window.addEventListener("resize", update);
    return () => { revision++; window.removeEventListener("resize", update); };
  }, []);

  useEffect(() => {
    let active = true;
    let gotLibrary = false;
    let gotSettings = false;
    const offSettings = events.settings.on(value => { gotSettings = true; setSettings(value); });
    const offLibrary = events.library.on(value => {
      gotLibrary = true;
      setLibrary(value);
    });
    Library.snapshot().then(value => {
      if (active && !gotLibrary) setLibrary(value);
    }).catch(reason => setError(String(reason)));
    Preferences.get().then(value => { if (active && !gotSettings) setSettings(value); }).catch(reason => setError(String(reason)));
    return () => {
      active = false;
      offLibrary();
      offSettings();
    };
  }, []);

  const action = async (run: () => Promise<unknown>) => {
    if (requestBusy.current) return;
    requestBusy.current = true;
    setBusy(true);
    try {
      await run();
      setError("");
    } catch (reason) {
      setError(String(reason));
    } finally {
      requestBusy.current = false;
      setBusy(false);
    }
  };

  const pageShortcut = (direction: number) => {
    if (requestBusy.current || isPagingControl(document.activeElement)) return false;
    const page = browsePageTarget(library, direction);
    if (page === null) return false;
    void action(() => Library.page(page));
    return true;
  };

  useEffect(() => {
    const keydown = (event: KeyboardEvent) => {
      if (event.defaultPrevented || event.repeat || event.isComposing || event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return;
      const direction = event.key === "ArrowLeft" ? -1 : event.key === "ArrowRight" ? 1 : 0;
      if (direction && pageShortcut(direction)) event.preventDefault();
    };
    window.addEventListener("keydown", keydown);
    return () => window.removeEventListener("keydown", keydown);
  });

  return <div class={`study-window deck-palette${maximized ? " maximized" : ""}`} style={transparencyStyle(settings)}>
    <WindowTitlebar subtitle="背单词" />
    <main class="study-main">
      <div class="study-toolbar">
        <div class="study-selectors">
          <select aria-label="切换 Excel 词表" value={library?.selected ?? ""} disabled={busy || !library?.books.some(book => !book.error)} onChange={event => void action(() => Library.select(event.currentTarget.value))}>
            {!library?.selected && <option value="">暂无词表</option>}
            {library?.books.map(book => <option key={book.name} value={book.name} disabled={!!book.error}>{book.name.replace(/\.xlsx$/i, "")} · {book.count} 词{book.error ? "（格式错误）" : ""}</option>)}
          </select>
          <select aria-label="学习模式" value={library?.quiz?.direction || "browse"} disabled={busy || !library?.selected} onChange={event => void action(() => event.currentTarget.value === "browse" ? Library.endQuiz() : Library.startQuiz(event.currentTarget.value))}>
            <option value="browse">浏览</option><option value="english-chinese">英选中</option>
          </select>
          <span class="study-word-count">{library?.total ?? 0} 个单词</span>
        </div>
        <div class="study-actions">
          <button type="button" onClick={() => void action(() => Island.openSettings())}>设置</button>
          <button type="button" class="study-return" onClick={() => void action(() => Island.returnToCapsule())}>回到胶囊</button>
        </div>
      </div>
      <section class="study-panel" aria-label="单词学习内容">
        <div class="study-content-stage island-stage">
          <LearningContent variant="study" library={library} error={error} busy={busy}
            onSettings={() => void action(() => Island.openSettings())}
            onFocus={index => void action(() => Library.focus(index))}
            onOrder={order => void action(() => Library.setBrowseOrder(order))}
            onPage={page => void action(() => Library.page(page))}
            onMode={mode => void action(() => mode === "browse" ? Library.endQuiz() : Library.startQuiz(mode))}
            onAnswer={(token, choice) => void action(() => Library.answerQuiz(token, choice))}
            onNext={token => void action(() => Library.nextQuiz(token))} />
        </div>
      </section>
    </main>
  </div>;
}
