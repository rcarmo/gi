// Gi-owned batch lifetime around the supplied composer. Source bytes stay pinned.
export function patchUploadCancel(source) {
    const edits = [
        ["// @ts-nocheck\n", "// @ts-nocheck\nimport { composeTransfers } from '../gi-compose-transfer.js';\n"],
        ['const capturedChatJid = currentChatJid;', "const capturedChatJid = currentChatJid;\n        const uploadBatch = capturedMediaFiles.length ? composeTransfers.beginUploadBatch(capturedChatJid?.replace(/^gi:/, '') || '') : null;"],
        ['// Upload media files first', "// Upload media files first\n                uploadBatch?.signal.throwIfAborted();"],
        ['await uploadMedia(file, capturedChatJid)', 'await uploadMedia(file, capturedChatJid, { signal: uploadBatch?.signal })'],
        ['const fileBlock = capturedFileRefs.length', "uploadBatch?.signal.throwIfAborted();\n                uploadBatch?.end(); // No cancellation once message dispatch can start.\n                const fileBlock = capturedFileRefs.length"],
        ["const detail = error?.message || 'Failed to send message.';", "const detail = uploadBatch?.signal.aborted ? 'Upload cancelled. Draft and attachments retained; send again to retry.' : (error?.message || 'Failed to send message.');"],
        ["} finally {\n                if (queueToken) onQueuedSubmissionEnd?.(queueToken);", "} finally {\n                uploadBatch?.end();\n                if (queueToken) onQueuedSubmissionEnd?.(queueToken);"],
    ];
    for (const [from,to] of edits) {
        if (source.split(from).length !== 2) throw new Error(`Upload cancel adapter anchor changed: ${from}`);
        source = source.replace(from,to);
    }
    return source;
}
