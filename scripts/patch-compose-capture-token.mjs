// Persisted draft capture and transport must identify the same submission.
// Adapt the immutable component at build time; unknown upstream drift fails.
export function patchComposeCaptureToken(source) {
    const anchor = '{ client_request_id: queueToken }';
    if (source.split(anchor).length !== 2) throw new Error('Compose capture-token adapter anchor changed');
    return source.replace(anchor, '{ client_request_id: capture?.token || queueToken }');
}
