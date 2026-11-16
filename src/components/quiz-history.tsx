import { useEffect, useRef, useState } from "preact/hooks";
import { events, Library, type LibraryState, type QuizHistoryState } from "../mygo";
import { buildTrend, historyByDay, monthDays, summarizeHistory, yearWeeks, type HistoryRecord, type TrendMode, type TrendPoint } from "../lib/quiz-history";
import "./quiz-history.css";

const scoreText = (value: number) => Number.isInteger(value) ? String(value) : value.toFixed(1);
const bookText = (value: string) => value.replace(/\.xlsx$/i, "");
const weekdays = ["一", "二", "三", "四", "五", "六", "日"];
const ALL_BOOKS = "*";

export function QuizHistory({ library }: { library: LibraryState | null }) {
  const today = new Date();
  const [history, setHistory] = useState<QuizHistoryState | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [book, setBook] = useState(ALL_BOOKS);
  const [year, setYear] = useState(today.getFullYear());
  const [month, setMonth] = useState(today.getMonth() + 1);
  const [mode, setMode] = useState<TrendMode>("day");
  const [wall, setWall] = useState<"month" | "year">("month");
  const [selection, setSelection] = useState<{ label: string; ids: string[]; prefix?: string } | null>(null);
  const [page, setPage] = useState(0);
  const revision = useRef(0);
  const currentFolder = useRef(library?.folder);
  currentFolder.current = library?.folder;
  const load = async () => {
    const request = ++revision.current;
    const folder = currentFolder.current;
    setLoading(true);
    try {
      const value = await Library.quizHistory();
      if (request !== revision.current || folder !== currentFolder.current) return;
      setHistory(value); setError("");
    } catch (reason) {
      if (request === revision.current && folder === currentFolder.current) { setHistory(null); setError(String(reason)); }
    } finally { if (request === revision.current) setLoading(false); }
  };
  useEffect(() => {
    setHistory(null); setBook(ALL_BOOKS); setSelection(null);
    void load();
    const off = events.historyChanged.on(folder => { if (folder === currentFolder.current) void load(); });
    return () => { revision.current++; off(); };
  }, [library?.folder]);
  const books = [...new Set([...(library?.books ?? []).map(item => item.name), ...(history?.records ?? []).map(item => item.book)])].sort((a, b) => a.localeCompare(b, "zh-CN", { numeric: true }));
  const activeBook = books.includes(book) ? book : ALL_BOOKS;
  const allBooks = activeBook === ALL_BOOKS;
  const scope = allBooks ? "all" : "single";
  const records = (history?.records ?? []).filter(record => allBooks || record.book === activeBook);
  const years = [...new Set([today.getFullYear(), year, ...records.map(record => Number(record.date.slice(0, 4)))])].sort((a, b) => b - a);
  const monthKey = `${year}-${String(month).padStart(2, "0")}`;
  const prefix = mode === "month" ? `${year}-` : `${monthKey}-`;
  const periodRecords = records.filter(record => record.date.startsWith(prefix));
  const summary = summarizeHistory(periodRecords, scope);
  const days = historyByDay(records);
  const points = buildTrend(records, mode, year, month, scope);
  const detailRecords = selection ? records.filter(record => selection.prefix ? record.date.startsWith(selection.prefix) : selection.ids.includes(record.id)) : periodRecords;
  const ordered = [...detailRecords].sort((a, b) => Date.parse(b.completedAt) - Date.parse(a.completedAt) || b.attempt - a.attempt);
  const pages = Math.ceil(ordered.length / 10);
  const safePage = Math.min(page, Math.max(0, pages - 1));
  const visibleDays = historyByDay(ordered.slice(safePage * 10, safePage * 10 + 10));
  useEffect(() => { setSelection(null); setPage(0); }, [activeBook, year, month, mode]);
  const choose = (label: string, rounds: HistoryRecord[], prefix?: string) => { setSelection({ label, ids: rounds.map(record => record.id), prefix }); setPage(0); };
  const dayButton = (date: string | null, compact: boolean) => {
    if (!date) return <span class="history-day-padding" aria-hidden="true" />;
    const rounds = days.get(date) ?? [];
    const daySummary = summarizeHistory(rounds, scope);
    const level = Math.min(4, rounds.length);
    const label = `${date}，完成 ${rounds.length} 轮${allBooks ? `，练习 ${daySummary.books} 份词表` : ""}${rounds.length ? `，${allBooks ? "综合" : ""}平均 ${scoreText(daySummary.average)} 分` : ""}`;
    return <button type="button" class={`history-day level-${level}${selection?.label === date ? " chosen" : ""}`} title={label} aria-label={label} aria-pressed={selection?.label === date} onClick={() => choose(date, rounds, date)}>{!compact && <><span>{Number(date.slice(8))}</span><small>{rounds.length ? `${rounds.length}轮${allBooks ? ` / ${daySummary.books}份` : ""}` : ""}</small></>}</button>;
  };
  return <div class="quiz-history">
    <p class="section-description">每完成一轮自动记一笔。同一天重复练习会分别保留，答对一个词得 2 分，跳过计 0 分。</p>
    <div class="history-filters">
      <label>词表<select aria-label="记录词表" value={activeBook} onChange={event => setBook(event.currentTarget.value)}><option value={ALL_BOOKS}>全部 List</option>{books.map(name => <option key={name} value={name}>{bookText(name)}</option>)}</select></label>
      <label>年份<select aria-label="记录年份" value={year} onChange={event => setYear(Number(event.currentTarget.value))}>{years.map(value => <option key={value} value={value}>{value} 年</option>)}</select></label>
      <label>月份<select aria-label="记录月份" value={month} onChange={event => setMonth(Number(event.currentTarget.value))}>{Array.from({ length: 12 }, (_, index) => <option key={index} value={index + 1}>{index + 1} 月</option>)}</select></label>
      <button type="button" disabled={loading} onClick={() => void load()}>{loading ? "读取中…" : "刷新记录"}</button>
    </div>
    {error && <p class="feedback error" role="alert">{error}</p>}
    {!error && !loading && !records.length && <p class="history-empty" role="status">{books.length ? `${allBooks ? "当前文件夹" : bookText(activeBook)}还没有测试记录。完成一轮英选中后，这里会自动出现成绩。` : "先添加一份词表，再完成一轮英选中测试。"}</p>}
    <section class="history-section" aria-label="成绩趋势">
      <div class="history-section-heading"><h4>成绩趋势</h4><Segments label="趋势粒度" value={mode} options={[["attempt", "按轮次"], ["day", "按日"], ["month", "按月"]]} onChange={value => setMode(value as TrendMode)} /></div>
      <div class="history-summary"><span>{mode === "month" ? `${year} 年` : `${year} 年 ${month} 月`}</span><span>完成 <strong>{summary.rounds}</strong> 轮</span><span>练习 <strong>{summary.days}</strong> 天</span>{allBooks && <span title="词表总数包含当前文件夹的词表和历史记录中的词表">练习词表 <strong>{summary.books} / {books.length}</strong> 份</span>}<span>{allBooks ? "综合平均" : "平均"} <strong>{summary.rounds ? scoreText(summary.average) : "—"}</strong> 分</span></div>
      <TrendChart points={points} mode={mode} allBooks={allBooks} onSelect={point => choose(point.label, point.records, mode === "attempt" ? undefined : point.key)} />
      <p class="history-note">{mode === "attempt" ? allBooks ? "每个点是一轮测试，按本月所有 List 的完成顺序排列。悬停查看词表、时间和成绩，点击查看明细。" : "每个点是一轮测试，按完成时间排列，点击可查看该轮成绩。" : allBooks ? mode === "day" ? "每天先计算各 List 当天的平均分，再对练过的 List 等权平均。点击查看当天各 List 的成绩，没练习的日期留空。" : "每月先计算各天的综合分，再对练习日等权平均；每天的各 List 也等权计算。没练习的月份留空。" : mode === "day" ? "每个点是当天各轮的平均分，点击可查看当天记录。没练习的日期留空。" : "每个点是一个月的平均分：先算每天平均，再对练习日等权平均。没练习的月份留空。"}</p>
    </section>
    <section class="history-section" aria-label="绿色打卡墙">
      <div class="history-section-heading"><h4>绿色打卡墙 <small>{wall === "month" ? `${year} 年 ${month} 月` : `${year} 年`}</small></h4><Segments label="打卡墙视图" value={wall} options={[["month", "月视图"], ["year", "年视图"]]} onChange={value => setWall(value as "month" | "year")} /></div>
      {wall === "month" ? <div class="history-month" role="group" aria-label={`${monthKey} 练习日历`}><div class="history-weekdays">{weekdays.map(day => <span key={day}>{day}</span>)}</div><div class="history-month-grid">{monthDays(year, month).map((date, index) => <div key={date ?? `empty-${index}`}>{dayButton(date, false)}</div>)}</div></div>
        : <div class="history-year-scroll"><div class="history-year" role="group" aria-label={`${year} 年练习日历`}><div class="history-year-weekdays">{["", "一", "", "三", "", "五", "", "日"].map((day, index) => <span key={index}>{day}</span>)}</div>{yearWeeks(year).map((week, index) => <div class="history-year-week" key={index}><span class="history-year-month">{week.some(date => date?.endsWith("-01")) ? `${Number(week.find(date => date?.endsWith("-01"))?.slice(5, 7))}月` : ""}</span>{week.map((date, day) => <div key={date ?? `empty-${day}`}>{dayButton(date, true)}</div>)}</div>)}</div></div>}
      <div class="history-legend"><span>{allBooks ? "每格一天 · 颜色累计全部 List 的轮数" : "每格一天 · 点击查看记录"}</span><span>少</span>{[0, 1, 2, 3, 4].map(level => <i key={level} class={`level-${level}`} title={level === 4 ? "4轮及以上" : `${level}轮`} />)}<span>多</span></div>
    </section>
    <section class="history-section" aria-label="历史明细">
      <div class="history-section-heading"><h4>历史明细 <small>{selection?.label ?? (mode === "month" ? `${year} 年` : `${year} 年 ${month} 月`)}</small></h4>{selection && <button type="button" onClick={() => { setSelection(null); setPage(0); }}>查看整个{mode === "month" ? "年份" : "月份"}</button>}</div>
      {ordered.length ? <div class="history-table-scroll"><table class="history-table"><thead><tr><th>完成时间</th>{allBooks && <th>词表</th>}<th>{allBooks ? "该词表当天次数" : "当天次数"}</th><th>{allBooks ? "该词表累计轮次" : "累计轮次"}</th><th>答对 / 答错 / 跳过</th><th>成绩</th></tr></thead><tbody>{[...visibleDays.entries()].reverse().flatMap(([date, rounds]) => [<tr class="history-date-row" key={date}><th colSpan={allBooks ? 6 : 5}>{date}</th></tr>, ...[...rounds].reverse().map(record => <tr key={record.id}><td>{record.completedAt.slice(11, 19)}</td>{allBooks && <td title={record.book}>{bookText(record.book)}</td>}<td>第 {record.dailyAttempt} 次</td><td>第 {record.attempt} 轮</td><td>{record.correct} / {record.wrong} / {record.skipped}</td><td><strong>{record.score}</strong> 分</td></tr>)])}</tbody></table></div> : <p class="history-empty">{loading ? "正在读取测试记录…" : "这段时间没有完成的测试。"}</p>}
      {pages > 0 && <nav class="book-list-pagination" aria-label="测试记录分页"><span>共 {ordered.length} 轮</span><button type="button" disabled={safePage === 0} onClick={() => setPage(safePage - 1)}>上一页</button><span>{safePage + 1} / {pages}</span><button type="button" disabled={safePage + 1 >= pages} onClick={() => setPage(safePage + 1)}>下一页</button></nav>}
    </section>
    <p class="history-note history-storage">保存位置：{history?.folder ?? `${library?.folder ?? "词表文件夹"}\\测试记录`}<br/>搬动词表时，请一起保留「测试记录」文件夹。</p>
  </div>;
}

function Segments({ label, value, options, onChange }: { label: string; value: string; options: string[][]; onChange: (value: string) => void }) {
  return <div class="history-segments" role="group" aria-label={label}>{options.map(([key, text]) => <button key={key} type="button" aria-pressed={value === key} onClick={() => key && onChange(key)}>{text}</button>)}</div>;
}

function TrendChart({ points, mode, allBooks, onSelect }: { points: TrendPoint[]; mode: TrendMode; allBooks: boolean; onSelect: (point: TrendPoint) => void }) {
  const width = Math.max(600, mode === "attempt" ? points.length * 30 + 80 : 600);
  const left = 44, right = width - 20, top = 22, bottom = 210;
  const maximum = Math.max(100, ...points.map(point => Math.ceil((point.value ?? 0) / 20) * 20));
  const x = (index: number) => points.length < 2 ? (left + right) / 2 : left + index / (points.length - 1) * (right - left);
  const y = (value: number) => bottom - value / maximum * (bottom - top);
  let path = "", connected = false;
  points.forEach((point, index) => { if (point.value === null) connected = false; else { path += `${connected ? " L" : " M"}${x(index)},${y(point.value)}`; connected = true; } });
  const hasData = points.some(point => point.value !== null);
  const tickStep = Math.max(1, Math.ceil(points.length / 8));
  return <div class="history-chart-scroll"><svg class="history-chart" style={{ minWidth: `${width}px` }} viewBox={`0 0 ${width} 252`} role="group" aria-label="成绩折线图">
    <title>成绩趋势；点击数据点查看明细</title>
    {Array.from({ length: 6 }, (_, index) => {
      const value = maximum / 5 * index;
      return <g key={index}><line class="history-grid-line" x1={left} x2={right} y1={y(value)} y2={y(value)} /><text x={left - 10} y={y(value) + 4} text-anchor="end">{scoreText(value)}</text></g>;
    })}
    <path class="history-trend-line" d={path} />
    {points.map((point, index) => {
      const showTick = (index % tickStep === 0 && points.length - 1 - index >= tickStep) || index === points.length - 1;
      const tick = mode === "attempt" ? allBooks ? `第${index + 1}次` : `第${point.records[0]?.attempt ?? index + 1}轮` : point.label;
      const description = point.value === null ? "" : `${point.label}，${scoreText(point.value)} 分，${point.records.length}轮${mode === "attempt" ? `，${point.records[0]?.completedAt.slice(0, 19).replace("T", " ")}` : ""}`;
      return <g key={point.key}>
        {showTick && <text x={x(index)} y={236} text-anchor="middle">{tick}</text>}
        {point.value !== null && <g class="history-trend-point" role="button" tabIndex={0} aria-label={description} onClick={() => onSelect(point)} onKeyDown={event => { if (event.key === "Enter" || event.key === " ") { event.preventDefault(); onSelect(point); } }}>
          <title>{description}</title><circle class="history-point-hit" cx={x(index)} cy={y(point.value)} r={9} /><circle cx={x(index)} cy={y(point.value)} r={4} />
        </g>}
      </g>;
    })}
    {!hasData && <text x={width / 2} y={118} text-anchor="middle">这段时间没有测试成绩</text>}
  </svg></div>;
}
