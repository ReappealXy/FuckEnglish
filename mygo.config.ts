import { defineConfig } from "mygo-cli";
export default defineConfig({
  name: "FuckEnglish Preview 0.1.2",
  identifier: "io.github.reappealxy.fuckenglish.preview012",
  version: "0.1.2",
  devUrl: "http://localhost:5173",
  devCommand: "bun run dev:web",
  buildCommand: "bun run build:web",
  frontendDist: "dist",
  bindings: "src/mygo.ts",
  out: "build",
  macos: { minimumSystemVersion: "13.0" },
});
