export function parseStatusStartedAt(status: Record<string, unknown> | null | undefined): number | null {
  if (!status || typeof status !== "object") return null;
  const raw = status.started_at ?? status.startedAt;
  if (typeof raw !== "string" || !raw) return null;
  const value = Date.parse(raw);
  return Number.isFinite(value) ? value : null;
}

export function parseStatusRetryAt(status: Record<string, unknown> | null | undefined): number | null {
  if (!status || typeof status !== "object") return null;
  const raw = status.retry_at ?? status.retryAt;
  if (typeof raw !== "string" || !raw) return null;
  const value = Date.parse(raw);
  return Number.isFinite(value) ? value : null;
}

export function parseStatusLastEventAt(status: Record<string, unknown> | null | undefined): number | null {
  if (!status || typeof status !== "object") return null;
  const raw = status.last_event_at ?? status.lastEventAt ?? status.started_at ?? status.startedAt;
  if (typeof raw !== "string" || !raw) return null;
  const value = Date.parse(raw);
  return Number.isFinite(value) ? value : null;
}

export function isCompactionStatus(status: Record<string, unknown> | null | undefined): boolean {
  if (!status || typeof status !== "object") return false;
  const intentKey = status.intent_key ?? status.intentKey;
  return status.type === "intent" && intentKey === "compaction";
}

export function resolveStatusPanelTitle(status: Record<string, unknown> | null | undefined): string {
  if (!status || typeof status !== "object") return "";
  const rawTitle = status.title;
  if (typeof rawTitle === "string" && rawTitle.trim()) return rawTitle.trim();

  const statusText = status.status;
  if (typeof statusText === "string" && statusText.trim()) return statusText.trim();

  return isCompactionStatus(status) ? "Compacting context" : "Working...";
}

export function formatElapsedDuration(elapsedMs: number): string {
  const totalSeconds = Math.max(0, Math.floor(elapsedMs / 1000));
  const seconds = totalSeconds % 60;
  const minutes = Math.floor(totalSeconds / 60) % 60;
  const hours = Math.floor(totalSeconds / 3600);
  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
  }
  return `${minutes}:${String(seconds).padStart(2, "0")}`;
}

export function getStatusElapsedLabel(
  status: Record<string, unknown> | null | undefined,
  nowMs = Date.now()
): string | null {
  const startedAtMs = parseStatusStartedAt(status);
  if (startedAtMs === null) return null;
  return formatElapsedDuration(Math.max(0, nowMs - startedAtMs));
}

export function getStatusRetryCountdownLabel(
  status: Record<string, unknown> | null | undefined,
  nowMs = Date.now()
): string | null {
  const retryAtMs = parseStatusRetryAt(status);
  if (retryAtMs === null) return null;
  const remainingMs = retryAtMs - nowMs;
  if (remainingMs <= 0) return "retrying now";
  return `retry in ${formatElapsedDuration(remainingMs)}`;
}
