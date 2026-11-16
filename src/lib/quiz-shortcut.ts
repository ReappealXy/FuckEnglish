import type { QuizState } from "../mygo";

/** The only quiz direction that accepts quiz shortcuts. */
const ENGLISH_TO_CHINESE = "english-chinese";

type QuizShortcutEvent = Pick<KeyboardEvent, "defaultPrevented" | "repeat" | "isComposing" | "altKey" | "ctrlKey" | "metaKey" | "shiftKey">;

function isQuizShortcutBlockedEvent(event: QuizShortcutEvent): boolean {
  return event.defaultPrevented || event.repeat || event.isComposing || event.altKey || event.ctrlKey || event.metaKey || event.shiftKey;
}

/** Return the zero-based answer position for an unmodified number key. */
export function quizChoiceFromKey(key: string): number | null {
  if (key.length !== 1 || key < "1" || key > "4") return null;
  return key.charCodeAt(0) - "1".charCodeAt(0);
}

/**
 * Controls which focused elements keep quiz keys for themselves. This also
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

/** Match the enabled next/finish button, including the automatic-advance wait. */
export function canUseQuizNextShortcut(quiz: QuizState | null | undefined, target: Element | null = null): boolean {
  return !!quiz
    && quiz.direction === ENGLISH_TO_CHINESE
    && quiz.token.trim() !== ""
    && quiz.total > 0
    && quiz.answered
    && !quiz.completed
    && (quiz.choice !== quiz.correct || quiz.index + 1 === quiz.total)
    && !isQuizShortcutBlockedTarget(target);
}

/**
 * Convert a keyboard event into an answer position. Returning null means the
 * event belongs to another control or the current question cannot be answered.
 */
export function quizChoiceFromEvent(
  event: QuizShortcutEvent & Pick<KeyboardEvent, "key">,
  quiz: QuizState | null | undefined,
  target: Element | null = null,
): number | null {
  if (isQuizShortcutBlockedEvent(event)) return null;
  if (!canUseQuizShortcuts(quiz, target)) return null;
  return quizChoiceFromKey(event.key);
}

/** Use the physical key before 1, regardless of the active input language. */
export function quizNextFromEvent(
  event: QuizShortcutEvent & Pick<KeyboardEvent, "code">,
  quiz: QuizState | null | undefined,
  target: Element | null = null,
): boolean {
  return !isQuizShortcutBlockedEvent(event)
    && event.code === "Backquote"
    && canUseQuizNextShortcut(quiz, target);
}
