import { expect, test } from "bun:test";
import { browsePageTarget } from "../src/lib/island-pagination";

const state = { selected: "list.xlsx", page: 0, pages: 10, quiz: { direction: "" } };

test("arrow paging keeps the first and last page boundaries", () => {
  expect(browsePageTarget(state, -1)).toBeNull();
  expect(browsePageTarget(state, 1)).toBe(1);
  expect(browsePageTarget({ ...state, page: 4 }, -1)).toBe(3);
  expect(browsePageTarget({ ...state, page: 9 }, 1)).toBeNull();
});

test("paging ignores quizzes, unavailable wordbooks, and invalid directions", () => {
  for (const direction of ["english-chinese", "chinese-english"]) {
    expect(browsePageTarget({ ...state, quiz: { direction } }, 1)).toBeNull();
  }
  expect(browsePageTarget(null, 1)).toBeNull();
  expect(browsePageTarget({ ...state, selected: "" }, 1)).toBeNull();
  expect(browsePageTarget({ ...state, pages: 0 }, 1)).toBeNull();
  expect(browsePageTarget(state, 0)).toBeNull();
  expect(browsePageTarget(state, 2)).toBeNull();
});
