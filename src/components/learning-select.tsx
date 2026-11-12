import { useEffect, useLayoutEffect, useRef, useState } from "preact/hooks";
import "./learning-select.css";

export type LearningSelectOption = { value: string; label: string; disabled?: boolean };

type Props = {
  id: string;
  label: string;
  value: string;
  options: LearningSelectOption[];
  disabled?: boolean;
  className?: string;
  onChange: (value: string) => void;
  onOpenChange?: (open: boolean) => void;
};

export function LearningSelect(props: Props) {
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(-1);
  const [placement, setPlacement] = useState({ above: false, maxHeight: 178 });
  const root = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const list = useRef<HTMLDivElement>(null);
  const opened = useRef(false);
  const notify = useRef(props.onOpenChange);
  notify.current = props.onOpenChange;
  const selected = props.options.findIndex(option => option.value === props.value);
  const available = props.options.map((option, index) => option.disabled ? -1 : index).filter(index => index >= 0);
  const disabled = props.disabled || !available.length;
  const menuId = `${props.id}-options`;

  const changeOpen = (value: boolean, restoreFocus = false) => {
    if (opened.current !== value) {
      opened.current = value;
      setOpen(value);
      notify.current?.(value);
    }
    if (restoreFocus) trigger.current?.focus();
  };

  const show = (last = false) => {
    if (disabled) return;
    setActive(available.includes(selected) ? selected : (last ? available.at(-1) : available[0]) ?? -1);
    changeOpen(true);
  };

  const choose = (index: number) => {
    const option = props.options[index];
    if (!option || option.disabled || disabled) return;
    changeOpen(false, true);
    if (option.value !== props.value) props.onChange(option.value);
  };

  useEffect(() => {
    if (disabled) changeOpen(false);
  }, [disabled]);

  useEffect(() => () => { if (opened.current) notify.current?.(false); }, []);

  useEffect(() => {
    if (!open) return;
    const outside = (event: Event) => {
      if (event.target instanceof Node && !root.current?.contains(event.target)) changeOpen(false);
    };
    document.addEventListener("pointerdown", outside, true);
    document.addEventListener("focusin", outside);
    return () => {
      document.removeEventListener("pointerdown", outside, true);
      document.removeEventListener("focusin", outside);
    };
  }, [open]);

  useLayoutEffect(() => {
    if (!open) return;
    const measure = () => {
      const control = trigger.current;
      if (!control) return;
      const rect = control.getBoundingClientRect();
      const boundary = root.current?.closest(".island-surface, .study-window")?.getBoundingClientRect();
      const row = parseFloat(getComputedStyle(root.current!).getPropertyValue("--learning-select-row-height")) || 34;
      const desired = Math.min(props.options.length, 5) * row + 8;
      const below = Math.max(0, (boundary?.bottom ?? window.innerHeight) - rect.bottom - 6);
      const above = Math.max(0, rect.top - (boundary?.top ?? 0) - 6);
      const upward = below < desired && above > below;
      setPlacement({ above: upward, maxHeight: Math.min(desired, upward ? above : below) });
    };
    measure();
    list.current?.focus({ preventScroll: true });
    window.addEventListener("resize", measure);
    return () => window.removeEventListener("resize", measure);
  }, [open, props.options.length]);

  useLayoutEffect(() => {
    if (open) list.current?.querySelector<HTMLElement>(`[data-option-index="${active}"]`)?.scrollIntoView({ block: "nearest" });
  }, [open, active]);

  useEffect(() => {
    if (open && !available.includes(active)) setActive(available.includes(selected) ? selected : available[0] ?? -1);
  }, [open, props.options, active, selected]);

  return <div ref={root} class={`learning-select${props.className ? ` ${props.className}` : ""}${open ? " is-open" : ""}`} data-learning-select data-learning-select-open={open ? "true" : undefined} onKeyDown={event => {
    if (event.defaultPrevented || event.isComposing || event.altKey || event.ctrlKey || event.metaKey) return;
    if (event.key === "Tab") { changeOpen(false); return; }
    if (event.key === "Escape" && open) { event.preventDefault(); event.stopPropagation(); changeOpen(false, true); return; }
    if (!["ArrowUp", "ArrowDown", "Home", "End", "Enter", " "].includes(event.key) || disabled) return;
    event.preventDefault();
    event.stopPropagation();
    if (event.repeat && (event.key === "Enter" || event.key === " ")) return;
    if (!open) {
      show(event.key === "ArrowUp" || event.key === "End");
      if (event.key === "Home") setActive(available[0] ?? -1);
      else if (event.key === "End") setActive(available.at(-1) ?? -1);
      return;
    }
    if (event.key === "Enter" || event.key === " ") { if (!event.repeat) choose(active); return; }
    if (event.key === "Home") setActive(available[0] ?? -1);
    else if (event.key === "End") setActive(available.at(-1) ?? -1);
    else {
      const position = available.indexOf(active);
      const next = Math.max(0, Math.min(available.length - 1, position + (event.key === "ArrowDown" ? 1 : -1)));
      setActive(available[next] ?? -1);
    }
  }}>
    <button ref={trigger} id={props.id} class="learning-select-trigger" type="button" role="combobox" aria-label={props.label} aria-haspopup="listbox" aria-expanded={open} aria-controls={menuId} disabled={disabled} onClick={() => open ? changeOpen(false) : show()}>
      <span>{props.options[selected]?.label ?? props.options[0]?.label ?? "暂无选项"}</span>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d={open ? "m6 15 6-6 6 6" : "m6 9 6 6 6-6"} /></svg>
    </button>
    {open && <div class={`learning-select-menu${placement.above ? " opens-above" : ""}`} style={{ maxHeight: `${placement.maxHeight}px` }}>
      <div ref={list} id={menuId} class="learning-select-list" role="listbox" tabIndex={-1} aria-label={props.label} aria-activedescendant={active >= 0 ? `${props.id}-option-${active}` : undefined}>
        {props.options.map((option, index) => <div id={`${props.id}-option-${index}`} key={option.value} role="option" aria-selected={option.value === props.value} aria-disabled={option.disabled || undefined} data-option-index={index} class={`learning-select-option${index === active ? " active" : ""}`} onPointerMove={() => { if (!option.disabled) setActive(index); }} onMouseDown={event => event.preventDefault()} onClick={() => choose(index)}>{option.label}</div>)}
      </div>
    </div>}
  </div>;
}
