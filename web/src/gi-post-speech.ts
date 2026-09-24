// Text normalisation and button behaviour follow Piclaw 70d33bc93 post-speech.ts.
// Gi owns playback per mounted post occurrence and selected-session generation.
import { buildPostMarkdownCopyPayload } from './utils/post-copy-markdown.js';

export function buildSpeakablePostText(post) {
    const raw = buildPostMarkdownCopyPayload(post);
    if (!raw) return '';
    return String(raw)
        .replace(/```[\s\S]*?```/g, ' Code block omitted. ')
        .replace(/`([^`]+)`/g, '$1')
        .replace(/!\[([^\]]*)\]\(([^)]+)\)/g, '$1')
        .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '$1')
        .replace(/^#{1,6}\s+/gm, '')
        .replace(/^>\s?/gm, '')
        .replace(/^[-*+]\s+/gm, '• ')
        .replace(/\n{3,}/g, '\n\n')
        .replace(/\n\n+/g, '. ')
        .replace(/\s+/g, ' ')
        .replace(/\s+([.,;:!?])/g, '$1')
        .trim().slice(0, 1600);
}

export function createSpeechPlayback(runtime = () => typeof window === 'undefined' ? null : window) {
    let scope = null, token = 0, active = null;
    const listeners = new Set();
    const state = () => ({ owner: active?.owner ?? null, speaking: Boolean(active) });
    const emit = () => { for (const listener of listeners) listener(state()); };
    const supported = () => {
        const win = runtime();
        return Boolean(win && typeof win.SpeechSynthesisUtterance === 'function' &&
            typeof win.speechSynthesis?.speak === 'function' && typeof win.speechSynthesis?.cancel === 'function');
    };
    const stop = (owner?) => {
        if (owner !== undefined && active?.owner !== owner) return;
        const previous = active;
        ++token; active = null;
        // Clear ownership before cancel: some engines call onend synchronously.
        if (previous) { try { previous.synth.cancel(); } catch {} }
        emit();
    };
    return {
        state, supported, stop,
        setScope(next) { if (scope !== next) { stop(); scope = next; } },
        subscribe(listener) { listeners.add(listener); listener(state()); return () => listeners.delete(listener); },
        speak(owner, text) {
            if (!scope || !supported()) return false;
            const value = String(text || '').trim().slice(0, 1600);
            if (!value) return false;
            stop();
            const win = runtime(), synth = win.speechSynthesis, current = ++token;
            try {
                const utterance = new win.SpeechSynthesisUtterance(value);
                const finish = () => { if (current === token) { active = null; emit(); } };
                utterance.onend = finish; utterance.onerror = finish;
                active = { owner, synth, utterance }; // Retain utterance until terminal callback.
                emit(); synth.speak(utterance); return true;
            } catch {
                if (current === token) { active = null; emit(); }
                return false;
            }
        },
    };
}

export const speechPlayback = createSpeechPlayback();
