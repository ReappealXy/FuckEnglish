import type { QuizState } from "../mygo";

/** The only quiz direction that accepts the four answer shortcuts. */
const ENGLISH_TO_CHINESE = "english-chinese";

/** Return the zero-based answer position for an unmodified number key. */
export function quizChoiceFromKey(key: string): number | null {
  if (key.length !== 1 || key < "1" || key > "4") return null;
  return key.charCodeAt(0) - "1".charCodeAt(0);
}

/**
 * Controls which focused elements keep number keys for themselves. This also
 * covers the custom wordbook picker, so opening a picker cannot answer a
 * question behind it.
 */
export function isQuizShortcutBlockedTarget(target: Element | null): boolean {
  // Selecting a mode restores focus to its trigger. Once that menu is closed,
  // number keys should immediately be available to answer the question.
  if (target?.closest('[data-learning-select]:not([data-learning-select-open="true"])')) return false;
  return !!target?.closest(
    'input, textarea, select, [contenteditable]:not([contenteditable="false"]), [role="combobox"], [role="listbox"], [role="menu"], [data-book-picker]',
  );
}

export function canUseQuizShortcuts(quiz: QuizState | null | undefined, target: Element | null = null): boolean {
  return !!quiz
    && quiz.direction === ENGLISH_TO_CHINESE
    && quiz.token.trim() !== ""
    && quiz.options.length === 4
    && !quiz.answered
    && !quiz.completed
    && !isQuizShortcutBlockedTarget(target);
}

/**
 * Convert a keyboard event into an answer position. Returning null means the
 * event belongs to another control or the current question cannot be answered.
 */
export function quizChoiceFromEvent(
  event: Pick<KeyboardEvent, "key" | "defaultPrevented" | "repeat" | "isComposing" | "altKey" | "ctrlKey" | "metaKey" | "shiftKey">,
  quiz: QuizState | null | undefined,
  target: Element | null = null,
): number | null {
  if (event.defaultPrevented || event.repeat || event.isComposing || event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return null;
  if (!canUseQuizShortcuts(quiz, target)) return null;
  return quizChoiceFromKey(event.key);
}
