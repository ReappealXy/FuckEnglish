import { render } from "preact";
import { isMyGo, runtime } from "mygo-runtime";
import { App } from "./app";
import { IslandView } from "./island";
import { StudyView } from "./study";

const view = new URLSearchParams(location.search).get("view");
const island = view === "island";
const study = view === "study";
if (island) document.documentElement.classList.add("island-window");
if (isMyGo()) document.documentElement.dataset.platform = runtime().platform;
render(island ? <IslandView /> : study ? <StudyView /> : <App />, document.getElementById("app")!);
