import { useEffect, useLayoutEffect, useRef, useState } from "preact/hooks";
import { isMyGo } from "mygo-runtime";
import { events, Island, Library, Preferences, type LibraryState, type Settings } from "./mygo";
import { islandHover } from "./lib/island-hover";
import { IslandSurface } from "./components/island-surface";
import "./island.css";

export function IslandView() {
  const nativePointer = navigator.userAgent.includes("Windows");
  const [library, setLibrary] = useState<LibraryState | null>(null);
  const [settings, setSettings] = useState<Settings | null>(null);
  const [expanded, setExpanded] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const requestBusy = useRef(false);
  const state = useRef({ expanded: false, settings: null as Settings | null });
  const transition = useRef<(value: boolean) => void>(() => {});
  const hover = useRef(islandHover({ set: (fn, delay) => window.setTimeout(fn, delay), clear: (id) => window.clearTimeout(id) }, value => transition.current(value)));
  transition.current = value => {
    if (state.current.expanded === value) return;
    state.current.expanded = value; setExpanded(value);
    void Island.setExpanded(value).catch(e => setError(String(e)));
  };
  useEffect(() => {
    if (!isMyGo()) return;
    let active = true, gotLibrary = false, gotSettings = false;
    const apply = (s: Settings) => { hover.current.cancel(); state.current.settings = s; setSettings(s); };
    const off = events.library.on(v => { gotLibrary = true; setLibrary(v); });
    const offSettings = events.settings.on(v => { gotSettings = true; apply(v); });
    const offDismiss = events.islandDismiss.on(() => { hover.current.cancel(); transition.current(false); });
    const offWindow = events.islandWindow.on(v => { if (!v.expanded) { hover.current.cancel(); state.current.expanded = false; setExpanded(false); } });
    const offPointer = events.islandPointer.on(inside => {
      if (!nativePointer) return;
      if (inside) hover.current.enter(state.current.settings?.islandHover !== false, state.current.settings?.islandExpandDelay ?? 160);
      else hover.current.leave(state.current.settings?.islandCollapseDelay ?? 420);
    });
    Library.snapshot().then(v => { if (active && !gotLibrary) setLibrary(v); }).catch(e => setError(String(e)));
    Preferences.get().then(v => { if (active && !gotSettings) apply(v); }).catch(e => setError(String(e)));
    const keydown = (event: KeyboardEvent) => { if (event.key === "Escape") { hover.current.cancel(); transition.current(false); } };
    window.addEventListener("keydown", keydown);
    return () => { active = false; hover.current.cancel(); off(); offSettings(); offDismiss(); offWindow(); offPointer(); window.removeEventListener("keydown", keydown); };
  }, []);
  useLayoutEffect(() => {
    if (!isMyGo() || !nativePointer) return;
    const surface = document.querySelector<HTMLElement>(".island-surface"); if (!surface) return;
    const report = () => {
      const rect = surface.getBoundingClientRect();
      const material = getComputedStyle(surface.querySelector<HTMLElement>(".island-material")!);
      void Island.setHitRegion({ x: rect.x, y: rect.y, width: rect.width, height: rect.height,
        topRadius: parseFloat(material.borderTopLeftRadius) || 0, bottomRadius: parseFloat(material.borderBottomLeftRadius) || 0,
        compact: !expanded }).catch(e => setError(String(e)));
    };
    const observer = new ResizeObserver(report); observer.observe(surface); report();
    surface.addEventListener("transitionend", report);
    return () => { observer.disconnect(); surface.removeEventListener("transitionend", report); };
  }, [expanded, settings?.islandStyle, settings?.islandWidth, settings?.islandHeight]);
  const action = async (run: () => Promise<unknown>) => {
    if (requestBusy.current) return;
    requestBusy.current = true; setBusy(true);
    try { await run(); setError(""); } catch (e) { setError(String(e)); }
    finally { requestBusy.current = false; setBusy(false); }
  };
  return <IslandSurface settings={settings} library={library} expanded={expanded} error={error} busy={busy}
    onEnter={nativePointer ? undefined : () => hover.current.enter(state.current.settings?.islandHover !== false, state.current.settings?.islandExpandDelay ?? 160)}
    onLeave={nativePointer ? undefined : () => hover.current.leave(state.current.settings?.islandCollapseDelay ?? 420)}
    onExpand={() => { hover.current.cancel(); transition.current(true); }} onCollapse={() => { hover.current.cancel(); transition.current(false); }}
    onSettings={() => void action(() => Island.openSettings())} onSelect={name => void action(() => Library.select(name))}
    onPage={page => void action(() => Library.page(page))} onFocus={index => void action(() => Library.focus(index))} />;
}
