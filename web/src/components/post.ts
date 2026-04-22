// @ts-nocheck
import { html, useEffect } from '../vendor/preact-htm-entry.ts';
import { FilePill } from './file-pill.ts';

export function Post({ msg, icon, logPrefix = '[post]' }) {
  useEffect(() => {
    console.debug(`${logPrefix} mounted`, { id: msg.id, role: msg.role });
    return () => console.debug(`${logPrefix} unmounted`, { id: msg.id, role: msg.role });
  }, [msg.id, msg.role]);

  const refs = Array.isArray(msg.payload?.file_refs) ? msg.payload.file_refs : [];
  return html`<article class=${`post gi-post role-${msg.role}`}><div class=${`post-avatar ${msg.role}`}>${icon}</div><div class="post-body"><div class="post-header"><span class="post-author">${msg.role}</span><span class="post-meta">${msg.payload?.source || msg.payload?.kind || ''}</span></div><div class="post-content"><div class="message-content">${msg.content}</div>${refs.length ? html`<div class="compose-file-refs">${refs.map((ref) => html`<${FilePill} prefix="post" label=${ref} title=${ref} />`)}</div>` : null}</div></div></article>`;
}
