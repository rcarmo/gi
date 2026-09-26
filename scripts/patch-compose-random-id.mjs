// Keep the supplied composer byte-identical; adapt its non-security queue ID.
export function patchComposeRandomId(source) {
    const anchor = 'crypto.randomUUID()';
    if (source.split(anchor).length !== 2) throw new Error('Compose UUID adapter anchor changed');
    return `import { randomClientId } from '../gi-random-id.js';\n` + source.replace(anchor, 'randomClientId()');
}
