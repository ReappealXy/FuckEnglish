import { expect, test } from "bun:test";
import { buildTrend, dailyAverage, dailyBookAverage, historyByDay, monthDays, summarizeHistory, yearWeeks, type HistoryRecord } from "../src/lib/quiz-history";

function record(date: string, score: number, attempt: number, overrides: Partial<HistoryRecord> = {}): HistoryRecord {
  return {
    id: `round-${attempt}`, book: "list1.xlsx", completedAt: `${date}T09:00:00+08:00`, date,
    correct: score / 2, wrong: 50 - score / 2, skipped: 0, total: 50, score, attempt, dailyAttempt: 1, ...overrides,
  };
}

test("same-day rounds are retained, ordered, and averaged without mutating records", () => {
  const records = [
    record("2026-11-16", 92, 3, { dailyAttempt: 3, completedAt: "2026-11-16T20:12:45+08:00" }),
    record("2026-11-16", 60, 1, { completedAt: "2026-11-16T09:10:23+08:00" }),
    record("2026-11-16", 76, 2, { dailyAttempt: 2, completedAt: "2026-11-16T10:35:08+08:00" }),
  ];
  const original = records.map(item => item.id);
  const grouped = historyByDay(records);
  expect([...grouped.keys()]).toEqual(["2026-11-16"]);
  expect(grouped.get("2026-11-16")?.map(item => item.score)).toEqual([60, 76, 92]);
  expect(dailyAverage(grouped.get("2026-11-16")!)).toBe(76);
  expect(summarizeHistory(records)).toEqual({ rounds: 3, days: 1, books: 1, average: 76 });
  expect(records.map(item => item.id)).toEqual(original);
  expect(buildTrend(records, "attempt", 2026, 11).map(point => point.key)).toEqual(["round-1", "round-2", "round-3"]);
});

test("monthly means weight practiced days equally and never treat missing days as zero", () => {
  const records = [record("2026-11-01", 100, 1), record("2026-11-01", 100, 2), record("2026-11-02", 40, 3), record("2026-12-01", 0, 4)];
  const month = buildTrend(records, "month", 2026, 11);
  expect(month).toHaveLength(12);
  expect(month[10]?.value).toBe(70);
  expect(month[11]?.value).toBe(0);
  expect(month[0]?.value).toBeNull();
  const days = buildTrend(records, "day", 2026, 11);
  expect(days).toHaveLength(30);
  expect(days[0]?.value).toBe(100);
  expect(days[1]?.value).toBe(40);
  expect(days[2]?.value).toBeNull();
  expect(days[2]?.records).toEqual([]);
});

test("stored local date controls grouping even when UTC belongs to another date", () => {
  const rounds = [record("2026-11-16", 80, 1, { completedAt: "2026-11-16T00:10:00+08:00" })];
  expect([...historyByDay(rounds).keys()]).toEqual(["2026-11-16"]);
  expect(buildTrend(rounds, "day", 2026, 11)[15]?.value).toBe(80);
  expect(buildTrend(rounds, "day", 2026, 11)[14]?.value).toBeNull();
});

test("same-second completions remain separate and year/month filters exclude other periods", () => {
  const rounds = [record("2026-11-16", 80, 2), record("2026-11-16", 60, 1), record("2025-11-16", 30, 3), record("2026-12-01", 90, 4)];
  const points = buildTrend(rounds, "attempt", 2026, 11);
  expect(points.map(point => point.key)).toEqual(["round-1", "round-2"]);
  expect(points.map(point => point.value)).toEqual([60, 80]);
  expect(buildTrend(rounds, "month", 2025, 11)[10]?.value).toBe(30);
});

test("month calendars start on Monday and include all days exactly once", () => {
  const november = monthDays(2026, 11); // November 1, 2026 is Sunday.
  expect(november.slice(0, 7)).toEqual([null, null, null, null, null, null, "2026-11-01"]);
  expect(november).toHaveLength(42);
  expect(november.filter(Boolean)).toHaveLength(30);
  const february = monthDays(2024, 2);
  expect(february.filter(Boolean)).toHaveLength(29);
  expect(february).toContain("2024-02-29");
  expect(monthDays(2026, 2).filter(Boolean)).toHaveLength(28);
  expect(monthDays(2027, 2)).toHaveLength(28); // Four complete Monday-first weeks.
});

test("year heatmaps preserve leap days, year edges, and seven-day columns", () => {
  for (const year of [2024, 2026, 2028]) {
    const weeks = yearWeeks(year);
    const dates = weeks.flat().filter((date): date is string => date !== null);
    expect(weeks.every(week => week.length === 7)).toBe(true);
    expect(dates[0]).toBe(`${year}-01-01`);
    expect(dates.at(-1)).toBe(`${year}-12-31`);
    expect(new Set(dates).size).toBe(dates.length);
    expect(dates).toHaveLength(year % 4 === 0 ? 366 : 365);
  }
  expect(yearWeeks(2024)[0]?.[0]).toBe("2024-01-01");
  expect(yearWeeks(2026)[0]?.slice(0, 4)).toEqual([null, null, null, "2026-01-01"]);
});

test("empty histories stay empty and calendar bounds are explicit", () => {
  expect(summarizeHistory([])).toEqual({ rounds: 0, days: 0, books: 0, average: 0 });
  expect(dailyAverage([])).toBe(0);
  expect(buildTrend([], "attempt", 2026, 11)).toEqual([]);
  expect(buildTrend([], "month", 2026, 11).every(point => point.value === null)).toBe(true);
  expect(() => monthDays(2026, 0)).toThrow(RangeError);
  expect(() => yearWeeks(0)).toThrow(RangeError);
});

test("all-List daily scores weight Lists equally despite repeated familiar rounds", () => {
  const rounds = [
    ...Array.from({ length: 5 }, (_, index) => record("2026-11-16", 90, index + 1, { dailyAttempt: index + 1 })),
    record("2026-11-16", 60, 1, { id: "list2-round-1", book: "list2.xlsx" }),
  ];
  expect(dailyBookAverage(rounds)).toBe(75);
  expect(dailyAverage(rounds)).toBe(85);
  expect(summarizeHistory(rounds, "all")).toEqual({ rounds: 6, days: 1, books: 2, average: 75 });
  const days = buildTrend(rounds, "day", 2026, 11, "all");
  expect(days[15]?.value).toBe(75);
  expect(days[15]?.records).toHaveLength(6);
  expect(days[14]?.value).toBeNull();
});

test("all-List monthly scores average practiced daily composite scores rather than rounds or per-List monthly scores", () => {
  const rounds = [
    ...Array.from({ length: 5 }, (_, index) => record("2026-11-16", 90, index + 1)),
    record("2026-11-16", 60, 1, { id: "list2-round-1", book: "list2.xlsx" }),
    record("2026-11-17", 30, 6),
    record("2026-12-01", 0, 2, { id: "list2-round-2", book: "list2.xlsx" }),
    record("2025-11-16", 100, 7),
  ];
  const months = buildTrend(rounds, "month", 2026, 11, "all");
  expect(months[10]?.value).toBe(52.5); // (75 + 30) / 2
  expect(months[11]?.value).toBe(0);
  expect(months[0]?.value).toBeNull();
  expect(months[10]?.records).toHaveLength(7);
  expect(buildTrend(rounds, "month", 2025, 11, "all")[10]?.value).toBe(100);
  expect(summarizeHistory(rounds.filter(round => round.date.startsWith("2026-11")), "all").average).toBe(52.5);
});

test("all-List attempts retain same-second records with List labels and sort by actual completion time", () => {
  const rounds = [
    record("2026-11-16", 60, 1, { id: "list2-round-1", book: "list2.xlsx", completedAt: "2026-11-16T09:00:00+08:00" }),
    record("2026-11-16", 90, 1, { completedAt: "2026-11-16T09:00:00+08:00" }),
    record("2026-11-16", 80, 2, { completedAt: "2026-11-16T08:30:00+07:00", dailyAttempt: 2 }),
    record("2026-11-16", 70, 3, { completedAt: "2026-11-16T09:15:00+08:00", dailyAttempt: 3 }),
  ];
  const points = buildTrend(rounds, "attempt", 2026, 11, "all");
  expect(points).toHaveLength(4);
  expect(new Set(points.map(point => point.key)).size).toBe(4);
  expect(points.slice(0, 2).map(point => point.label).sort()).toEqual(["list1 11-16 第1次", "list2 11-16 第1次"]);
  expect(points.map(point => point.value).slice(2)).toEqual([70, 80]);
  expect(rounds[0]?.id).toBe("list2-round-1");
});

test("all-List aggregation excludes unpracticed Lists and retains zero scores and single-List behavior", () => {
  const rounds = [record("2026-11-16", 100, 1), record("2026-11-16", 0, 2), record("2026-11-17", 0, 3)];
  expect(dailyBookAverage(rounds.slice(0, 2))).toBe(50);
  expect(summarizeHistory(rounds, "all")).toEqual({ rounds: 3, days: 2, books: 1, average: 25 });
  expect(summarizeHistory(rounds, "all")).toEqual(summarizeHistory(rounds));
  expect(buildTrend(rounds, "day", 2026, 11, "all")).toEqual(buildTrend(rounds, "day", 2026, 11));
  expect(buildTrend(rounds, "month", 2026, 11, "all")).toEqual(buildTrend(rounds, "month", 2026, 11));
  expect(summarizeHistory([], "all")).toEqual({ rounds: 0, days: 0, books: 0, average: 0 });
  expect(dailyBookAverage([])).toBe(0);
  expect(buildTrend([], "day", 2026, 11, "all").every(point => point.value === null)).toBe(true);
});
