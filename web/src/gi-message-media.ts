// Native media references are ordered and session-owned. Keep the supplied
// Post component's parallel media_ids/content_blocks projection in one place.
import { projectResourceLinks } from './gi-message-links.js';
export function projectMessageMedia(payload: any, sessionId: string) {
    const blocks = projectResourceLinks(Array.isArray(payload?.content_blocks) ? payload.content_blocks : []);
    const refs = Array.isArray(payload?.media) ? payload.media.filter((ref: any) =>
        Number.isSafeInteger(ref?.media_id) && ref.media_id > 0 &&
        (!ref.session_id || ref.session_id === sessionId)) : [];
    if (!refs.length) return { media_ids: [], content_blocks: blocks.length ? blocks : null };
    const mediaBlocks = refs.map((ref: any) => ({
        type: /^image\/(png|jpeg|gif|webp|avif|bmp|svg\+xml)$/i.test(ref.content_type || '') ? 'image' : 'file',
        name: ref.filename || `attachment-${ref.media_id}`,
        mime_type: ref.content_type || 'application/octet-stream',
    }));
    return {
        media_ids: refs.map((ref: any) => ref.media_id),
        content_blocks: [...blocks.filter((block: any) => block?.type !== 'image' && block?.type !== 'file'), ...mediaBlocks],
    };
}
