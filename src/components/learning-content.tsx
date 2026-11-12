import type { LibraryState } from "../mygo";
import { useLayoutEffect, useRef } from "preact/hooks";

type Props = {
  library: LibraryState | null;
  error?: string;
  busy?: boolean;
  variant?: "study";
  onSettings?: () => void;
  onFocus?: (index: number) => void;
  onOrder?: (order: string) => void;
  onPage?: (page: number) => void;
  onMode?: (mode: string) => void;
  onAnswer?: (token: string, choice: number) => void;
  onNext?: (token: string) => void;
};

export function LearningContent(props: Props) {
  const { library, busy } = props;
  const study = props.variant === "study";
  const words = library?.words ?? [];
  const error = props.error || library?.error;
  const quiz = library?.quiz;
  const testing = !!quiz?.direction;
  const lastQuestion = !!quiz && quiz.index + 1 === quiz.total;
  const autoAdvancing = !!quiz?.answered && quiz.choice === quiz.correct && !lastQuestion;
  const quizStatus = !quiz ? "" : quiz.completed ? "本轮小测已完成" : quiz.answered
    ? quiz.choice === quiz.correct ? "答对了" : `答错了 · 正确答案：${quiz.options[quiz.correct]}`
    : quiz.total ? `题目与选项已乱序${quiz.skipped ? ` · 跳过 ${quiz.skipped} 词` : ""}` : `跳过 ${quiz.skipped} 词`;

  if (testing && quiz) return <>
    <div class={`vocabulary-page quiz-page${study ? " study-quiz-page" : ""}`}>
      {quiz.completed ? <div class="quiz-message" role="status"><strong>本轮答对 {quiz.score} / {quiz.total} 题</strong><span>{quiz.skipped ? `${quiz.skipped} 个词因干扰项缺失或重复未参与本轮` : "本词表已全部测完"}</span><div class="quiz-actions"><button type="button" disabled={busy} onClick={() => props.onMode?.(quiz.direction)}>再测一轮</button><button type="button" disabled={busy} onClick={() => props.onMode?.("browse")}>返回浏览</button></div></div>
      : quiz.total === 0 ? <div class="quiz-message" role="status"><strong>没有可测试的单词</strong><span>请在 Excel 中为每个词填写三个不同的中文干扰项，且不能与正确答案相同，再刷新词表。</span><span>原有单词仍可正常浏览。</span><div class="quiz-actions"><button type="button" disabled={busy} onClick={() => props.onMode?.("browse")}>返回浏览</button></div></div>
      : <><QuizHeading question={quiz.question} phonetic={quiz.phonetic} study={study} /><div class="quiz-options" role="group" aria-label="选择答案">{quiz.options.map((option, index) => {
        const correct = quiz.answered && index === quiz.correct;
        const wrong = quiz.answered && index === quiz.choice && index !== quiz.correct;
        return <button type="button" key={`${quiz.token}-${index}`} class={`quiz-choice${correct ? " correct" : ""}${wrong ? " wrong" : ""}`} disabled={busy || quiz.answered} aria-keyshortcuts={String(index + 1)} title={`按 ${index + 1} 选择`} aria-label={`${option}${correct ? "，正确答案" : wrong ? "，回答错误" : ""}`} onClick={() => props.onAnswer?.(quiz.token, index)}><span class="quiz-choice-text" tabIndex={quiz.answered ? 0 : -1}>{option}</span>{(correct || wrong) && <span class="quiz-mark">{correct ? "正确" : "选错"}</span>}</button>;
      })}</div></>}
    </div>
    <footer class="island-statusbar vocabulary-footer quiz-footer"><span class="island-status" role="status" title={error || quizStatus}>{error || quizStatus}</span>{quiz.total > 0 && !quiz.completed && <div class="word-pagination"><span>{quiz.index + 1} / {quiz.total}</span><button type="button" disabled={busy || !quiz.answered || autoAdvancing} onClick={() => props.onNext?.(quiz.token)}>{lastQuestion ? "完成" : autoAdvancing ? "自动下一题…" : "下一题"}</button></div>}</footer>
  </>;

  return <>
    <div class="vocabulary-page">
      {words.length > 0 ? <div class="vocabulary-table" role="table" aria-label="当前页单词">
        <div class="vocabulary-head" role="row"><span role="columnheader">英文单词</span><span role="columnheader">音标</span><span role="columnheader">中文释义</span></div>
        <div class="vocabulary-rows">{words.map((word, index) => <div class={`vocabulary-row${index === library?.focus ? " focused" : ""}`} role="row" key={`${library?.page}-${index}`}>
          <span role="cell"><button class="word-focus" type="button" onClick={() => props.onFocus?.(index)} aria-label={`在胶囊显示 ${word.english}`} aria-pressed={index === library?.focus}>{word.english}</button></span>
          <span class="word-phonetic" role="cell">{word.phonetic || "—"}</span>
          <span class="word-meaning" role="cell" tabIndex={0}>{word.meaning}</span>
        </div>)}</div>
      </div> : <div class="vocabulary-empty">{error ? <><span>{error}</span><button type="button" onClick={props.onSettings}>打开词表设置</button></> : <><span>{library ? "把 Excel 放进词表文件夹，然后点击刷新。" : "正在读取词表…"}</span><button type="button" onClick={props.onSettings}>选择词表文件夹</button></>}</div>}
    </div>
    <footer class="island-statusbar vocabulary-footer">
      <div class="browse-status"><span class={`island-status ${words.length ? "connected" : ""}`} aria-live="polite"><i />{error || (words.length ? `第 ${(library?.page ?? 0) * 5 + 1}–${(library?.page ?? 0) * 5 + words.length} 词` : "等待词表")}</span><div class="browse-order" role="group" aria-label="浏览顺序">{([['sequential', '顺序'], ['random', '乱序']] as const).map(([order, label]) => <button type="button" aria-pressed={(library?.browseOrder ?? "sequential") === order} disabled={busy || !library?.selected} onClick={() => props.onOrder?.(order)}>{label}</button>)}</div></div>
      <div class="word-pagination"><button type="button" title="上一页（←）" aria-keyshortcuts="ArrowLeft" disabled={busy || !library || library.page <= 0} onClick={() => props.onPage?.((library?.page ?? 0) - 1)}>上一页</button><span>{library?.pages ? `${library.page + 1} / ${library.pages}` : "0 / 0"}</span><button type="button" title="下一页（→）" aria-keyshortcuts="ArrowRight" disabled={busy || !library || library.page + 1 >= library.pages} onClick={() => props.onPage?.((library?.page ?? 0) + 1)}>下一页</button></div>
    </footer>
  </>;
}

function QuizHeading({ question, phonetic, study }: { question: string; phonetic?: string; study: boolean }) {
  const ref = useRef<HTMLDivElement>(null);
  useLayoutEffect(() => {
    const heading = ref.current;
    const page = heading?.parentElement;
    if (!heading || !page) return;
    let disposed = false;
    const fit = () => {
      if (disposed || !page.clientWidth || !page.clientHeight) return;
      const style = getComputedStyle(page);
      const height = page.clientHeight - parseFloat(style.paddingTop) - parseFloat(style.paddingBottom);
      // Keep room for all four answers. The study window uses fixed answer rows;
      // the capsule divides its remaining space evenly between the two rows.
      let budget = height * .45;
      const options = page.querySelector<HTMLElement>(".quiz-options");
      if (study && options) budget = Math.min(budget, height - options.offsetHeight - parseFloat(style.rowGap));
      budget = Math.max(1, Math.floor(budget));
      const scale = (value: number) => heading.style.setProperty("--quiz-heading-scale", String(value));
      const fits = () => heading.scrollHeight <= budget && heading.scrollWidth <= heading.clientWidth;
      scale(1);
      if (fits()) return;
      let low = 0, high = 1;
      for (let i = 0; i < 14; i++) {
        const mid = (low + high) / 2;
        scale(mid);
        if (fits()) low = mid; else high = mid;
      }
      scale(low);
    };
    fit();
    const observer = new ResizeObserver(fit);
    observer.observe(page);
    void document.fonts.ready.then(fit);
    return () => { disposed = true; observer.disconnect(); };
  }, [question, phonetic, study]);
  return <div ref={ref} class={`quiz-heading${study ? " study-quiz-heading" : ""}`}><div class="quiz-prompt" tabIndex={0} aria-label="当前题目">{question}</div>{phonetic && <div class="quiz-phonetic" aria-label="音标">{phonetic}</div>}</div>;
}
