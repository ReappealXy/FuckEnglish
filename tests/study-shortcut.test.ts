import { expect, test } from "bun:test";
import { studyToggleFromEvent } from "../src/lib/study-shortcut";

const event = (patch: Partial<KeyboardEvent> = {}) => ({ code: "KeyQ", defaultPrevented: false, repeat: false, isComposing: false, altKey: false, ctrlKey: false, metaKey: false, shiftKey: false, ...patch });
const target = (...selectors: string[]) => ({ closest: (selector: string) => selectors.some(match => selector.split(", ").includes(match)) ? {} : null }) as unknown as Element;

test("the unmodified physical Q key switches learning views", () => {
  expect(studyToggleFromEvent(event())).toBe(true);
  expect(studyToggleFromEvent(event({ code: "KeyW" }))).toBe(false);
  expect(studyToggleFromEvent(event(), target())).toBe(true);
});

test("Q preserves text entry, open menus and handled keyboard actions", () => {
  for (const property of ["defaultPrevented", "repeat", "isComposing", "altKey", "ctrlKey", "metaKey", "shiftKey"] as const) expect(studyToggleFromEvent(event({ [property]: true }))).toBe(false);
  for (const selector of ["input", "textarea", "select", '[contenteditable]:not([contenteditable="false"])', '[role="combobox"]', '[role="listbox"]', '[role="menu"]', "[data-book-picker]"]) expect(studyToggleFromEvent(event(), target(selector))).toBe(false);
  expect(studyToggleFromEvent(event(), null, true)).toBe(false);
});

test("closed custom mode controls release Q once their menu closes", () => {
  expect(studyToggleFromEvent(event(), target('[role="combobox"]', '[data-learning-select]:not([data-learning-select-open="true"])'))).toBe(true);
});
