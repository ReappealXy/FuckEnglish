import { defineConfig } from "mygo-cli";
export default defineConfig({
  name: "FuckEnglish",
  identifier: "io.github.reappealxy.fuckenglish.preview011",
  version: "0.1.7",
  devUrl: "http://localhost:5173",
  devCommand: "bun run dev:web",
  buildCommand: "bun run build:web",
  frontendDist: "dist",
  bindings: "src/mygo.ts",
  out: "build",
  macos: { minimumSystemVersion: "13.0" },
});
