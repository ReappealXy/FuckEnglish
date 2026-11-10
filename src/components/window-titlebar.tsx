import { currentWindow, isMyGo } from "mygo-runtime";

export function WindowTitlebar({ subtitle }: { subtitle: string }) {
  // A hidden MyGo title bar already places native buttons over the page.
  // Reserve their area instead of drawing another set underneath them.
  return <header class="app-titlebar">
    <strong>FuckEnglish</strong><span>{subtitle}</span>
    {isMyGo() ? <div class="native-window-actions" aria-hidden="true" /> : <div class="window-actions">
      <button type="button" aria-label="最小化" onClick={() => void currentWindow.minimize()}>−</button>
      <button type="button" aria-label="最大化或恢复" onClick={() => void currentWindow.toggleMaximize()}>□</button>
      <button type="button" aria-label="关闭窗口并保留胶囊" onClick={() => void currentWindow.close()}>×</button>
    </div>}
  </header>;
}
