type TransparencySettings = { backgroundTransparency?: number; textTransparency?: number };

export function transparencyStyle(settings: TransparencySettings | null): Record<string, string> {
  const opacity = (value = 0) => String(1 - Math.max(0, Math.min(100, Number.isFinite(value) ? value : 0)) / 100);
  return {
    "--learning-background-opacity": opacity(settings?.backgroundTransparency),
    "--learning-text-opacity": opacity(settings?.textTransparency),
  };
}
