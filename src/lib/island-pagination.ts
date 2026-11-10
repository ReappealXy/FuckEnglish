type BrowseState = { selected: string; page: number; pages: number; quiz?: { direction: string } };

export function browsePageTarget(library: BrowseState | null, direction: number): number | null {
  if (!library?.selected || library.quiz?.direction || (direction !== -1 && direction !== 1)) return null;
  const page = library.page + direction;
  return page >= 0 && page < library.pages ? page : null;
}

export function isPagingControl(element: Element | null): boolean {
  return !!element?.closest('select, input, textarea, [contenteditable]:not([contenteditable="false"]), [role="combobox"]');
}
