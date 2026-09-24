// The Gi host mounts previews, not an editor. Preserve supplied source bytes
// while making its existing callback's visible description truthful.
export function patchWorkspaceReadonly(source) {
    for (const [from, to] of [
        ["? 'Open in editor'", "? 'Open read-only tab'"],
        ["? 'File too large to edit'", "? 'File too large for a read-only tab'"],
        [": 'File is not editable'", ": 'File preview only'"],
        ['disabled=${!canEdit}>Open in editor</button>', 'disabled=${!canEdit}>Open read-only tab</button>'],
    ]) {
        if (source.split(from).length !== 2) throw new Error(`Read-only workspace adapter anchor changed: ${from}`);
        source = source.replace(from, to);
    }
    return source;
}
