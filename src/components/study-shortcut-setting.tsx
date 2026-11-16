import { defaultStudyShortcut, formatStudyShortcut, hasStudyShortcutModifier, studyShortcutKeys, studyShortcutModifiers, studyShortcutParts } from "../lib/study-shortcut";

export function StudyShortcutSetting({ value, disabled, onChange }: { value: string; disabled: boolean; onChange: (shortcut: string) => void }) {
  const { modifiers, key } = studyShortcutParts(value);
  const invalid = !hasStudyShortcutModifier(value);
  return <div class="study-shortcut-setting">
    <div class="study-shortcut-heading"><label for="study-shortcut-key">胶囊与背词窗口切换</label><strong>{formatStudyShortcut(value)}</strong></div>
    <div class="study-shortcut-controls" role="group" aria-label="全局切换快捷键" aria-describedby="study-shortcut-help">
      {studyShortcutModifiers.map(modifier => <label class="study-shortcut-modifier" key={modifier}><input type="checkbox" checked={modifiers.includes(modifier)} disabled={disabled} onChange={event => onChange([...studyShortcutModifiers.filter(item => item === modifier ? event.currentTarget.checked : modifiers.includes(item)), key].join("+"))}/>{modifier === "Super" ? "Win" : modifier}</label>)}
      <span aria-hidden="true">+</span>
      <select id="study-shortcut-key" aria-label="快捷键主键" disabled={disabled} value={key} onChange={event => onChange([...modifiers, event.currentTarget.value].join("+"))}>
        {studyShortcutKeys.map(item => <option key={item} value={item}>{item}</option>)}
      </select>
      <button type="button" disabled={disabled || value === defaultStudyShortcut} onClick={() => onChange(defaultStudyShortcut)}>恢复默认</button>
    </div>
    <p id="study-shortcut-help" class="deck-settings-note">全局生效，鼠标移开或使用其他软件时也能切换。至少勾选 Ctrl、Alt、Win 中的一项，修改后点击“保存设置”。默认 Alt + Q。</p>
    {invalid && <p class="error-text study-shortcut-error" role="alert">请至少勾选 Ctrl、Alt、Win 中的一项。</p>}
  </div>;
}
