import { render } from "preact";
import { isMyGo, runtime } from "mygo-runtime";
import { App } from "./app";
import { IslandView } from "./island";

const island = new URLSearchParams(location.search).get("view") === "island";
if (island) document.documentElement.classList.add("island-window");
if (isMyGo()) document.documentElement.dataset.platform = runtime().platform;
render(island ? <IslandView /> : <App />, document.getElementById("app")!);
