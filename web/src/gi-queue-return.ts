import { emptyDraft, type Draft } from './gi-drafts.js';

// Awaiting media may cross a selection change or concurrent edit. The caller
// merges only after this returns, reading the latest origin-session repository.
export async function recoverQueueDraft(item: any, parsed: any, fetcher: typeof fetch = fetch): Promise<Draft> {
    if (!item?.chat_jid?.startsWith('gi:') || !item?.id) throw new Error('Missing queue origin');
    const session = item.chat_jid.slice(3);
    const attachments = new Map<string, any>();
    for (const media of item.metadata?.media || []) {
        if (media.session_id && media.session_id !== session) throw new Error('Attachment belongs to another session');
        const id = media.media_id || String(media.id || '').replace(/^media:/, '');
        if (id) attachments.set(String(id), { ...media, id });
    }
    for (const ref of parsed.attachmentRefs || []) {
        if (!attachments.has(String(ref.id))) attachments.set(String(ref.id), ref);
    }
    const media: File[] = [];
    for (const ref of attachments.values()) {
        if (!/^\d+$/.test(String(ref.id))) throw new Error('Invalid queued attachment identifier');
        const response = await fetcher(`/api/sessions/${encodeURIComponent(session)}/media/${ref.id}`);
        if (!response.ok) throw new Error(`Cannot restore queued attachment: HTTP ${response.status}`);
        const blob = await response.blob();
        if (blob.size > 10 * 1024 * 1024) throw new Error('Queued attachment exceeds 10 MiB limit');
        media.push(new File([blob], ref.filename || ref.label || `attachment-${ref.id}`, {
            type: ref.content_type || blob.type,
            lastModified: Date.parse(ref.created_at || '') || 0,
        }));
    }
    return { ...emptyDraft(), text: parsed.text || '', fileRefs: parsed.fileRefs || [], messageRefs: parsed.messageRefs || [], media };
}
