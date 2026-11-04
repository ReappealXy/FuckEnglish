import { expect, test } from "bun:test";
import { islandHover } from "../src/lib/island-hover";
import { islandTheme, islandThemes, islandThemeStyle } from "../src/lib/island-theme";

function harness() {
  let clock = 0, sequence = 0;
  const pending = new Map<number, { at: number; run: () => void }>();
  const changes: boolean[] = [];
  const hover = islandHover({ set: (run, delay) => { const id = ++sequence; pending.set(id, { at: clock + delay, run }); return id; }, clear: id => { pending.delete(id); } }, value => changes.push(value));
  const advance = (time: number) => {
    const to = clock + time;
    for (;;) {
      const next = [...pending.entries()].filter(([, item]) => item.at <= to).sort((a,b) => a[1].at-b[1].at)[0];
      if (!next) break;
      clock = next[1].at; pending.delete(next[0]); next[1].run();
    }
    clock = to;
  };
  return { hover, changes, advance };
}
test("leaving before the delay cancels expansion", () => {
  const h = harness(); h.hover.enter(true,160); h.advance(50); h.hover.leave(420); h.advance(500);
  expect(h.changes).toEqual([false]);
});
test("returning cancels a pending collapse", () => {
  const h = harness(); h.hover.enter(true,160); h.advance(160); h.hover.leave(420); h.advance(300); h.hover.enter(true,160); h.advance(500);
  expect(h.changes).toEqual([true,true]);
});
test("disabled hover and cancellation prevent queued opens", () => {
  const h = harness(); h.hover.enter(false,0); h.advance(1000); expect(h.changes).toEqual([]);
  h.hover.enter(true,160); h.hover.cancel(); h.advance(1000); expect(h.changes).toEqual([]);
});
test("all original themes and system light/dark remain available", () => {
  expect(islandThemes).toHaveLength(8);
  expect(islandTheme("dark").tokens.surface).toBe("#000000");
  expect(islandTheme("light").tokens.surface).toBe("#FAFAFC");
  expect(islandTheme("system",false).id).toBe("light");
  expect(islandTheme("system",true).id).toBe("dark");
  expect(islandTheme("missing").id).toBe("dark");
  for (const theme of islandThemes) expect(islandThemeStyle(theme.id)["--island-surface"]).toBe(theme.tokens.surface);
});
