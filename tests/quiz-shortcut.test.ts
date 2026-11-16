import { expect, test } from "bun:test";
import type { QuizState } from "../src/mygo";
import { canUseQuizNextShortcut, canUseQuizShortcuts, isQuizShortcutBlockedTarget, quizChoiceFromEvent, quizChoiceFromKey, quizNextFromEvent } from "../src/lib/quiz-shortcut";

const quiz = (overrides: Partial<QuizState> = {}): QuizState => ({
  direction: "english-chinese", token: "question-1", question: "analyse", phonetic: "[ˈænəlaɪz]",
  options: ["分析", "总结", "检查", "研究"], index: 0, total: 1, skipped: 0,
  answered: false, choice: -1, correct: -1, score: 0, completed: false, ...overrides,
});

const event = (key: string, overrides: Partial<KeyboardEvent> = {}) => ({
  key, code: "", defaultPrevented: false, repeat: false, isComposing: false,
  altKey: false, ctrlKey: false, metaKey: false, shiftKey: false, ...overrides,
});

// Match only selectors that this focused element or its ancestors represent.
// An unconditional closest() mock would also pretend that every control lives
// inside a closed custom select and bypass the input/menu protections.
const focusedElement = (...matchingSelectors: string[]) => ({
  closest: (selector: string) => matchingSelectors.some(match => selector.split(", ").includes(match)) ? {} : null,
}) as unknown as Element;

test("number keys map to the two-by-two answer positions", () => {
  expect(["1", "2", "3", "4"].map(quizChoiceFromKey)).toEqual([0, 1, 2, 3]);
  expect(quizChoiceFromKey("0")).toBeNull();
  expect(quizChoiceFromKey("5")).toBeNull();
  expect(quizChoiceFromKey("Num1")).toBeNull();
});

test("answer shortcuts accept only an unanswered English-to-Chinese question", () => {
  expect(quizChoiceFromEvent(event("1"), quiz())).toBe(0);
  expect(quizChoiceFromEvent(event("4"), quiz())).toBe(3);
  expect(quizChoiceFromEvent(event("1"), quiz({ direction: "chinese-english" }))).toBeNull();
  expect(quizChoiceFromEvent(event("1"), quiz({ token: "" }))).toBeNull();
  expect(quizChoiceFromEvent(event("1"), quiz({ options: ["a"] }))).toBeNull();
  expect(quizChoiceFromEvent(event("1"), quiz({ answered: true }))).toBeNull();
  expect(quizChoiceFromEvent(event("1"), quiz({ completed: true }))).toBeNull();
  expect(canUseQuizShortcuts(null)).toBe(false);
});

test("modifiers, repeats, composition, and handled events are ignored", () => {
  for (const property of ["defaultPrevented", "repeat", "isComposing", "altKey", "ctrlKey", "metaKey", "shiftKey"] as const) {
    expect(quizChoiceFromEvent(event("2", { [property]: true }), quiz())).toBeNull();
  }
});

test("focused editable and menu controls keep number keys", () => {
  for (const selector of ["input", "textarea", "select", '[contenteditable]:not([contenteditable="false"])', '[role="combobox"]', '[role="listbox"]', '[role="menu"]', "[data-book-picker]"]) {
    const blocked = focusedElement(selector);
    expect(isQuizShortcutBlockedTarget(blocked)).toBe(true);
    expect(quizChoiceFromEvent(event("3"), quiz(), blocked)).toBeNull();
  }
  const allowed = focusedElement();
  expect(isQuizShortcutBlockedTarget(allowed)).toBe(false);
  expect(quizChoiceFromEvent(event("3"), quiz(), allowed)).toBe(2);
});

test("closed custom selects release number keys while open menus keep them", () => {
  const closedTrigger = focusedElement('[role="combobox"]', '[data-learning-select]:not([data-learning-select-open="true"])');
  const openTrigger = focusedElement('[role="combobox"]');
  const openMenu = focusedElement('[role="listbox"]');
  expect(isQuizShortcutBlockedTarget(closedTrigger)).toBe(false);
  expect(quizChoiceFromEvent(event("1"), quiz(), closedTrigger)).toBe(0);
  expect(quizChoiceFromEvent(event("1"), quiz(), openTrigger)).toBeNull();
  expect(quizChoiceFromEvent(event("1"), quiz(), openMenu)).toBeNull();
});

test("the physical key before 1 advances wrong answers and finishes the last answer", () => {
  const wrong = quiz({ answered: true, choice: 1, correct: 0, total: 3 });
  for (const key of ["`", "·", "~", "Dead"]) {
    expect(quizNextFromEvent(event(key, { code: "Backquote" }), wrong)).toBe(true);
  }
  expect(quizNextFromEvent(event("`", { code: "Quote" }), wrong)).toBe(false);
  expect(quizNextFromEvent(event("`"), wrong)).toBe(false);
  expect(canUseQuizNextShortcut(quiz({ answered: true, choice: 0, correct: 0 }))).toBe(true);
  expect(canUseQuizNextShortcut(quiz({ answered: true, choice: 1, correct: 0 }))).toBe(true);
});

test("next shortcuts cannot skip unanswered questions or interrupt automatic advance", () => {
  const next = event("`", { code: "Backquote" });
  expect(quizNextFromEvent(next, quiz())).toBe(false);
  expect(quizNextFromEvent(next, quiz({ answered: true, choice: 0, correct: 0, total: 3 }))).toBe(false);
  for (const overrides of [
    { direction: "chinese-english" }, { token: "" }, { total: 0 }, { completed: true },
  ]) {
    expect(quizNextFromEvent(next, quiz({ answered: true, choice: 1, correct: 0, ...overrides }))).toBe(false);
  }
  expect(canUseQuizNextShortcut(null)).toBe(false);
});

test("next shortcuts ignore repeats, composition, modifiers, inputs, and open menus", () => {
  const answered = quiz({ answered: true, choice: 1, correct: 0 });
  for (const property of ["defaultPrevented", "repeat", "isComposing", "altKey", "ctrlKey", "metaKey", "shiftKey"] as const) {
    expect(quizNextFromEvent(event("`", { code: "Backquote", [property]: true }), answered)).toBe(false);
  }
  for (const selector of ["input", "textarea", "select", '[contenteditable]:not([contenteditable="false"])', '[role="combobox"]', '[role="listbox"]', '[role="menu"]', "[data-book-picker]"]) {
    expect(quizNextFromEvent(event("`", { code: "Backquote" }), answered, focusedElement(selector))).toBe(false);
  }
  const closedTrigger = focusedElement('[role="combobox"]', '[data-learning-select]:not([data-learning-select-open="true"])');
  expect(quizNextFromEvent(event("`", { code: "Backquote" }), answered, closedTrigger)).toBe(true);
});
