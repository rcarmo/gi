// Frozen timeline026 orders timestamp before existing outcome chips.
export function patchPostOutcomes(source) {
    const start = '                    ${recoveryMarker && html`';
    const timestamp = '                    <a class="post-time"';
    const end = '                    }}>${formatTime(post.timestamp)}</a>';
    for (const anchor of [start, timestamp, end]) {
        if (source.split(anchor).length !== 2) throw new Error('Post outcome anchor changed: ' + anchor);
    }
    const a = source.indexOf(start), b = source.indexOf(timestamp), c = source.indexOf(end) + end.length;
    if (!(a < b && b < c)) throw new Error('Post outcome ordering already adapted or changed');
    return source.slice(0, a) + source.slice(b, c) + '\n' + source.slice(a, b).trimEnd() + source.slice(c);
}
