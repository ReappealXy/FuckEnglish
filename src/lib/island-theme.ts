import palettes from "./island-themes.json";

export const islandThemes = palettes;
const legacy: Record<string, string> = { ocean: "night-owl", violet: "dracula", mint: "one-dark-pro", amber: "cyberpunk-edgerunners" };

export function islandTheme(id: string, systemDark = true) {
  const key = id === "system" ? systemDark ? "dark" : "light" : legacy[id] ?? id;
  return palettes.find((theme) => theme.id === key) ?? palettes[0]!;
}

export function islandThemeStyle(id: string, systemDark = true): Record<string, string> {
  const theme = islandTheme(id, systemDark);
  return Object.fromEntries(Object.entries(theme.tokens).map(([key, value]) => [`--island-${key}`, value]));
}
