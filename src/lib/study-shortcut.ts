import { isQuizShortcutBlockedTarget } from "./quiz-shortcut";

type StudyShortcutEvent = Pick<KeyboardEvent, "code" | "defaultPrevented" | "repeat" | "isComposing" | "altKey" | "ctrlKey" | "metaKey" | "shiftKey">;

/** Q belongs to the learning surface only; menus and text entry keep their keys. */
export function studyToggleFromEvent(event: StudyShortcutEvent, target: Element | null = null, controlsOpen = false): boolean {
  return event.code === "KeyQ"
    && !event.defaultPrevented && !event.repeat && !event.isComposing
    && !event.altKey && !event.ctrlKey && !event.metaKey && !event.shiftKey
    && !controlsOpen && !isQuizShortcutBlockedTarget(target);
}
