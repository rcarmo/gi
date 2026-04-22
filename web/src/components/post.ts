// @ts-nocheck
import { html, useEffect, useMemo } from '../vendor/preact-htm-entry.ts';
import { FilePill } from './file-pill.ts';

function renderMarkdown(content) {
  try {
    const md = globalThis.marked;
    if (md?.parse) return md.parse(String(content || ''));
  } catch (err) {
    console.error('[post] markdown render failed', err);
  }
  return String(content || '').replace(/[&<>]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[c]));
}

export function Post({ msg, icon, logPrefix = '[post]' }) {
  useEffect(() => {
    console.debug(`${logPrefix} mounted`, { id: msg.id, role: msg.role });
    return () => console.debug(`${logPrefix} unmounted`, { id: msg.id, role: msg.role });
  }, [msg.id, msg.role]);

  const refs = Array.isArray(msg.payload?.file_refs) ? msg.payload.file_refs : [];
  const renderedHtml = useMemo(() => renderMarkdown(msg.content), [msg.content]);
  return html`<article class=${`post gi-post role-${msg.role}`}><div class=${`post-avatar ${msg.role}`}>${icon}</div><div class="post-body"><div class="post-header"><span class="post-author">${msg.role}</span><span class="post-meta">${msg.payload?.source || msg.payload?.kind || ''}</span></div><div class="post-content"><div class="message-content" dangerouslySetInnerHTML=${{ __html: renderedHtml }}></div>${refs.length ? html`<div class="compose-file-refs">${refs.map((ref) => html`<${FilePill} prefix="post" label=${ref} title=${ref} />`)}</div>` : null}</div></div></article>`;
}
