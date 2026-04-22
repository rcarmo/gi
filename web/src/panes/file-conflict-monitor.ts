// @ts-nocheck
/**
 * file-conflict-monitor.ts — Reusable external-change detector for editor panes.
 *
 * Polls /workspace/stat for mtime changes and manages a conflict notification bar
 * with Reload / Save Copy / Overwrite / Dismiss actions.
 *
 * Usage:
 *   const monitor = createFileConflictMonitor({ path, getCurrentMtime, ... });
 *   monitor.start();          // begin polling
 *   monitor.stop();           // stop polling (e.g. on dispose)
 *   monitor.onSaved(newMtime); // call after a successful save
 */

import { getWorkspaceFileStat } from '../api.js';

export interface FileConflictMonitorOptions {
  /** Workspace-relative file path. */
  path: string;
  /** Returns the mtime the editor last loaded/saved. */
  getCurrentMtime: () => string | null;
  /** Element to insert the conflict bar before (typically the editor body). */
  anchorParent: HTMLElement;
  /** Element to insert the conflict bar before. */
  anchorBefore: HTMLElement;
  /** Called when user clicks Reload — should reload content from disk. */
  onReload: () => void | Promise<void>;
  /** Called when user clicks Save Copy — receives auto-generated path. */
  onSaveCopy: (copyPath: string) => void | Promise<void>;
  /** Called when user clicks Overwrite — should save current content. */
  onOverwrite: () => void | Promise<void>;
  /** Polling interval in ms. Default 3000. */
  pollMs?: number;
  /** ownerDocument for DOM creation. */
  ownerDocument?: Document;
}

export interface FileConflictMonitor {
  start(): void;
  stop(): void;
  /** Call after a successful save to update the known mtime and resume polling. */
  onSaved(newMtime: string | null): void;
  /** Remove the conflict bar if visible and stop everything. */
  dispose(): void;
}

function generateCopyPath(originalPath: string): string {
  const ext = originalPath.includes('.') ? originalPath.slice(originalPath.lastIndexOf('.')) : '';
  const base = originalPath.includes('.') ? originalPath.slice(0, originalPath.lastIndexOf('.')) : originalPath;
  const stamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
  return `${base}.${stamp}${ext}`;
}

export function createFileConflictMonitor(opts: FileConflictMonitorOptions): FileConflictMonitor {
  const {
    path,
    getCurrentMtime,
    anchorParent,
    anchorBefore,
    onReload,
    onSaveCopy,
    onOverwrite,
    pollMs = 3000,
    ownerDocument = document,
  } = opts;

  let timer: ReturnType<typeof setInterval> | null = null;
  let barEl: HTMLElement | null = null;
  let detected = false;
  let disposed = false;
  let saving = false;

  async function checkMtime() {
    if (disposed || saving || detected) return;
    const knownMtime = getCurrentMtime();
    if (!knownMtime) return;
    try {
      const stat = await getWorkspaceFileStat(path);
      if (disposed || saving || !stat?.mtime) return;
      if (stat.mtime !== knownMtime) {
        detected = true;
        stopPolling();
        showBar();
      }
    } catch (err) {
      // Network errors during mtime polling are expected (e.g. offline, reload).
      // Silently skip this cycle — the next interval will retry.
      if (typeof console !== 'undefined') console.debug('[file-conflict-monitor] mtime poll skipped:', err);
    }
  }

  function startPolling() {
    stopPolling();
    if (disposed) return;
    timer = setInterval(checkMtime, pollMs);
  }

  function stopPolling() {
    if (timer) { clearInterval(timer); timer = null; }
  }

  function showBar() {
    if (barEl || disposed) return;
    const bar = ownerDocument.createElement('div');
    bar.className = 'editor-conflict-bar';
    bar.innerHTML = `
      <span class="editor-conflict-text">File changed on disk</span>
      <div class="editor-conflict-actions">
        <button class="editor-conflict-btn" data-action="reload" title="Discard and reload from disk">Reload</button>
        <button class="editor-conflict-btn" data-action="save-copy" title="Save current content with a new name">Save copy</button>
        <button class="editor-conflict-btn" data-action="overwrite" title="Overwrite the disk version">Overwrite</button>
        <button class="editor-conflict-btn editor-conflict-dismiss" data-action="dismiss" title="Dismiss">\u00d7</button>
      </div>
    `;
    bar.addEventListener('click', (e) => {
      const btn = (e.target as HTMLElement).closest('[data-action]');
      if (!btn) return;
      const action = btn.getAttribute('data-action');
      if (action === 'reload') { dismissBar(); onReload(); }
      else if (action === 'save-copy') { const cp = generateCopyPath(path); onSaveCopy(cp); }
      else if (action === 'overwrite') { dismissBar(); onOverwrite(); }
      else if (action === 'dismiss') { dismissBar(); }
    });
    barEl = bar;
    anchorParent.insertBefore(bar, anchorBefore);
  }

  function dismissBar() {
    if (barEl) { barEl.remove(); barEl = null; }
    detected = false;
    startPolling();
  }

  return {
    start() { startPolling(); },
    stop() { stopPolling(); },
    onSaved(newMtime: string | null) {
      detected = false;
      saving = false;
      startPolling();
    },
    dispose() {
      disposed = true;
      stopPolling();
      if (barEl) { barEl.remove(); barEl = null; }
    },
  };
}
