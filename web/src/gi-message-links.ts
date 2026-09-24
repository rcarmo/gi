// Stored link metadata only: no discovery, server-side fetch or remote image load.
const MAX_LINKS = 8;
function text(value: unknown, max: number): string {
    return typeof value === 'string' ? value.slice(0, max) : '';
}
export function remoteLinkUrl(value: unknown): string | null {
    if (typeof value !== 'string' || value.length > 2048 || /[\s\\\u0000-\u001f\u007f]/.test(value)) return null;
    try {
        const url = new URL(value);
        return ['https:', 'http:'].includes(url.protocol) && !url.username && !url.password ? url.href : null;
    } catch { return null; }
}
export function projectResourceLinks(blocks: unknown[]): unknown[] {
    let count = 0;
    return blocks.flatMap((block: any) => {
        if (block?.type !== 'resource_link') return [block];
        const uri = remoteLinkUrl(block.uri);
        if (!uri || count++ >= MAX_LINKS) return [];
        return [{ type: 'resource_link', uri, title: text(block.title || block.name, 200),
            description: text(block.description, 1000), mimeType: text(block.mimeType, 100),
            ...(Number.isSafeInteger(block.size) && block.size >= 0 ? { size: block.size } : {}) }];
    });
}
export function projectLinkPreviews(payload: any) {
    if (!Array.isArray(payload?.link_previews)) return null;
    const seen = new Set<string>(), previews = [];
    for (const entry of payload.link_previews) {
        const url = remoteLinkUrl(entry?.url);
        if (!url || seen.has(url)) continue;
        seen.add(url);
        previews.push({ url, title: text(entry.title, 200), description: text(entry.description, 1000),
            site_name: new URL(url).hostname });
        if (previews.length === MAX_LINKS) break;
    }
    return previews.length ? previews : null;
}
