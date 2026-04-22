// @ts-nocheck

export function parseBtwCommand(input) {
    const trimmed = String(input || '').trim();
    if (!trimmed.startsWith('/btw')) return null;

    const rest = trimmed.slice(4).trim();
    if (!rest) {
        return { type: 'help' };
    }

    if (rest === 'clear' || rest === 'close') {
        return { type: 'clear' };
    }

    return {
        type: 'ask',
        question: rest,
    };
}

export function resolveBtwChatJid(chatJid) {
    const normalized = String(chatJid || '').trim();
    return normalized || 'web:default';
}

export function shouldShowBtwAnswer(session) {
    if (!session) return false;
    const answer = String(session.answer || '').trim();
    return session.status !== 'running' && Boolean(answer);
}

export function shouldShowBtwControls(session) {
    if (!session) return false;
    return session.status !== 'running';
}

export function buildBtwInjectionText(session) {
    const question = String(session?.question || '').trim();
    const answer = String(session?.answer || '').trim();
    if (!question && !answer) return '';

    return [
        'BTW side conversation',
        question ? `Question: ${question}` : null,
        answer ? `Answer:\n${answer}` : null,
    ].filter(Boolean).join('\n\n');
}
