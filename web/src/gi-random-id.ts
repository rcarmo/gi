/** UUIDv4 for non-security UI/draft identifiers on supported HTTP origins.
 * getRandomValues is available outside secure contexts; randomUUID is not.
 * Never substitute Math.random or weaken secure-origin authentication gates.
 */
export function randomClientId(source: Pick<Crypto, 'getRandomValues'> = globalThis.crypto): string {
    if (!source || typeof source.getRandomValues !== 'function') {
        throw new Error('Secure random generation is unavailable; message not sent.');
    }
    const bytes = source.getRandomValues(new Uint8Array(16));
    bytes[6] = (bytes[6] & 0x0f) | 0x40;
    bytes[8] = (bytes[8] & 0x3f) | 0x80;
    const hex = Array.from(bytes, byte => byte.toString(16).padStart(2, '0')).join('');
    return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}
