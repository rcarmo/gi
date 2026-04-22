import { useEffect, useState } from "../vendor/preact-htm.js";

function readSilenceOverride(key: string, fallback: number): number {
  try {
    if (typeof window === "undefined") return fallback;
    const overrides = (window as any).__PICLAW_SILENCE || {};
    const directKey = `__PICLAW_SILENCE_${key.toUpperCase()}_MS`;
    const raw = overrides[key] ?? (window as any)[directKey];
    const value = Number(raw);
    return Number.isFinite(value) ? value : fallback;
  } catch {
    return fallback;
  }
}

export const SILENCE_WARNING_MS = readSilenceOverride("warning", 30_000);
export const SILENCE_FINALIZE_MS = readSilenceOverride("finalize", 120_000);
export const SILENCE_REFRESH_MS = readSilenceOverride("refresh", 30_000);
export const LAST_ACTIVITY_TTL_MS = 30_000;

export function buildAgentsMap(data: { agents?: Array<{ id: string }> } | null | undefined): Record<string, unknown> {
  const map: Record<string, unknown> = {};
  (data?.agents || []).forEach((agent) => {
    map[agent.id] = agent;
  });
  return map;
}

/** Detect iOS devices for layout adjustments. */
export function isIOSDevice(): boolean {
  if (/iPad|iPhone/.test(navigator.userAgent)) {
    return true;
  }
  return navigator.platform === "MacIntel" && navigator.maxTouchPoints > 1;
}

/** Hook to force re-render for updating timestamps. */
export function useTimestampRefresh(intervalMs = 30_000): void {
  const [, setTick] = useState(0);

  useEffect(() => {
    const timer = setInterval(() => setTick((t) => t + 1), intervalMs);
    return () => clearInterval(timer);
  }, [intervalMs]);
}

export function estimatePreviewLines(text: unknown, maxCharsPerLine = 160): number {
  const value = String(text || "").replace(/\r\n/g, "\n");
  if (!value) return 0;
  return value
    .split("\n")
    .reduce((acc, line) => acc + Math.max(1, Math.ceil(line.length / maxCharsPerLine)), 0);
}
