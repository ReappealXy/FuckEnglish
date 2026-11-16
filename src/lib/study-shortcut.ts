export const defaultStudyShortcut = "Alt+Q";
export const studyShortcutModifiers = ["Ctrl", "Alt", "Shift", "Super"] as const;
export const studyShortcutKeys = [
  ..."ABCDEFGHIJKLMNOPQRSTUVWXYZ", ..."0123456789",
  ...Array.from({ length: 12 }, (_, index) => `F${index + 1}`),
];

export function studyShortcutParts(shortcut = defaultStudyShortcut) {
  const parts = shortcut.split("+");
  return { modifiers: studyShortcutModifiers.filter(modifier => parts.includes(modifier)), key: parts.at(-1) ?? "Q" };
}

export function formatStudyShortcut(shortcut = defaultStudyShortcut) {
  return shortcut.split("+").map(part => part === "Super" ? "Win" : part).join(" + ");
}

export function studyShortcutAria(shortcut = defaultStudyShortcut) {
  return shortcut.replaceAll("Ctrl", "Control").replaceAll("Super", "Meta");
}

export function hasStudyShortcutModifier(shortcut: string) {
  return studyShortcutParts(shortcut).modifiers.some(modifier => modifier !== "Shift");
}
