// Narrow compatibility override for the pinned Post component. Keep every
// other supplied helper, but absence of writeText is not clipboard success.
export * from './components/post-runtime-safety.js';

export async function writeClipboardTextBestEffort(
    clipboard: { writeText?: (value: string) => Promise<unknown> } | null | undefined,
    value: string,
): Promise<boolean> {
    try {
        if (typeof clipboard?.writeText !== 'function') return false;
        await clipboard.writeText(value);
        return true;
    } catch { return false; }
}
