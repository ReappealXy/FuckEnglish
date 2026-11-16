/** Saved completion dates are local calendar dates, independent of the viewer's timezone. */
export interface HistoryRecord {
  id: string;
  book: string;
  completedAt: string;
  date: string;
  correct: number;
  wrong: number;
  skipped: number;
  total: number;
  score: number;
  attempt: number;
  dailyAttempt: number;
}

export type TrendMode = "attempt" | "day" | "month";

export interface TrendPoint {
  key: string;
  label: string;
  value: number | null;
  records: HistoryRecord[];
}

function sortedRecords(records: readonly HistoryRecord[]): HistoryRecord[] {
  return [...records].sort((a, b) => a.date.localeCompare(b.date)
    || Date.parse(a.completedAt) - Date.parse(b.completedAt)
    || a.attempt - b.attempt);
}

export function historyByDay(records: readonly HistoryRecord[]): Map<string, HistoryRecord[]> {
  const days = new Map<string, HistoryRecord[]>();
  for (const record of sortedRecords(records)) {
    const existing = days.get(record.date);
    if (existing) existing.push(record);
    else days.set(record.date, [record]);
  }
  return days;
}

/** Arithmetic mean of the supplied rounds; normally called for a single day. */
export function dailyAverage(records: readonly HistoryRecord[]): number {
  return records.length ? records.reduce((sum, record) => sum + record.score, 0) / records.length : 0;
}

/** Every practiced day has the same weight, even when one day has many more rounds. */
export function summarizeHistory(records: readonly HistoryRecord[]): { rounds: number; days: number; average: number } {
  const days = historyByDay(records);
  const average = days.size
    ? [...days.values()].reduce((sum, rounds) => sum + dailyAverage(rounds), 0) / days.size
    : 0;
  return { rounds: records.length, days: days.size, average };
}

// UTC is used only for calendar arithmetic, never to reinterpret a saved completion date.
function calendarDate(year: number, month: number, day: number): Date {
  const value = new Date(0);
  value.setUTCFullYear(year, month - 1, day);
  value.setUTCHours(0, 0, 0, 0);
  return value;
}

function dateKey(value: Date): string {
  return `${String(value.getUTCFullYear()).padStart(4, "0")}-${String(value.getUTCMonth() + 1).padStart(2, "0")}-${String(value.getUTCDate()).padStart(2, "0")}`;
}

function validateCalendar(year: number, month = 1): void {
  if (!Number.isInteger(year) || year < 1 || year > 9999 || !Number.isInteger(month) || month < 1 || month > 12) {
    throw new RangeError("请选择有效的年份和月份");
  }
}

/** Month grid, Monday first. Null cells pad both ends to complete weeks. Month is 1–12. */
export function monthDays(year: number, month: number): (string | null)[] {
  validateCalendar(year, month);
  const first = calendarDate(year, month, 1);
  const count = calendarDate(year, month + 1, 0).getUTCDate();
  const offset = (first.getUTCDay() + 6) % 7;
  const cells: (string | null)[] = Array.from({ length: offset }, () => null);
  for (let day = 1; day <= count; day++) cells.push(dateKey(calendarDate(year, month, day)));
  while (cells.length % 7) cells.push(null);
  return cells;
}

/** Year heatmap columns; every column contains Monday through Sunday, with year-external cells null. */
export function yearWeeks(year: number): (string | null)[][] {
  validateCalendar(year);
  const cursor = calendarDate(year, 1, 1);
  cursor.setUTCDate(cursor.getUTCDate() - (cursor.getUTCDay() + 6) % 7);
  const last = calendarDate(year, 12, 31);
  const weeks: (string | null)[][] = [];
  while (cursor <= last) {
    const week: (string | null)[] = [];
    for (let day = 0; day < 7; day++) {
      week.push(cursor.getUTCFullYear() === year ? dateKey(cursor) : null);
      cursor.setUTCDate(cursor.getUTCDate() + 1);
    }
    weeks.push(week);
  }
  return weeks;
}

/** Build one month of rounds/days, or one year of monthly averages. Missing practice stays null. */
export function buildTrend(records: readonly HistoryRecord[], mode: TrendMode, year: number, month: number): TrendPoint[] {
  validateCalendar(year, month);
  const yearPrefix = `${String(year).padStart(4, "0")}-`;
  const monthPrefix = `${yearPrefix}${String(month).padStart(2, "0")}-`;
  if (mode === "attempt") {
    return sortedRecords(records.filter(record => record.date.startsWith(monthPrefix))).map(record => ({
      key: record.id,
      label: `${record.date.slice(5)} 第${record.dailyAttempt}次`,
      value: record.score,
      records: [record],
    }));
  }
  if (mode === "day") {
    const days = historyByDay(records);
    return monthDays(year, month).filter((date): date is string => date !== null).map(date => {
      const rounds = days.get(date) ?? [];
      return { key: date, label: `${Number(date.slice(8))}日`, value: rounds.length ? dailyAverage(rounds) : null, records: rounds };
    });
  }
  return Array.from({ length: 12 }, (_, index) => {
    const key = `${yearPrefix}${String(index + 1).padStart(2, "0")}`;
    const rounds = sortedRecords(records.filter(record => record.date.startsWith(`${key}-`)));
    return { key, label: `${index + 1}月`, value: rounds.length ? summarizeHistory(rounds).average : null, records: rounds };
  });
}
