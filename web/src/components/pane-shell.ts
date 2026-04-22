// @ts-nocheck
import { html, useEffect, useState } from '../vendor/preact-htm-entry.ts';

export function PaneShell({ filePath, fileContent }) {
  const [editorError, setEditorError] = useState(null);
  const hostRef = { current: null };
  useEffect(() => {
    let view = null;
    let cancelled = false;
    async function mount() {
      if (!filePath || !fileContent) return;
      try {
        const cm = await import('#editor-vendor/codemirror');
        if (cancelled || !document.getElementById('gi-pane-editor')) return;
        const state = cm.EditorState.create({ doc: fileContent, extensions: [cm.lineNumbers(), cm.highlightActiveLine(), cm.syntaxHighlighting(cm.defaultHighlightStyle, { fallback: true }), ...cm.minimalSetup] });
        view = new cm.EditorView({ state, parent: document.getElementById('gi-pane-editor') });
        console.debug('[pane] codemirror mounted', { filePath });
      } catch (err) {
        console.error('[pane] codemirror load failed', err);
        setEditorError(String(err?.message || err));
      }
    }
    mount();
    return () => { cancelled = true; try { view?.destroy?.(); } catch {} };
  }, [filePath, fileContent]);
  return html`<section class="editor-pane-container gi-pane-shell"><div class="panel-title">Pane</div>${filePath ? html`<div class="chat-window-header-badge">${filePath}</div>` : html`<div class="empty-state">No file open.</div>`}<div id="gi-pane-editor" class="gi-pane-editor"></div>${editorError ? html`<pre class="workspace-error">${editorError}</pre>` : null}</section>`;
}
