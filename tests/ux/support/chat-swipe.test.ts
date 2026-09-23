import { test, expect } from 'bun:test';
import { resolveSwipeableChatAgents, resolveAdjacentSwipeChatJid } from '../../../web/src/ui/chat-swipe-navigation.ts';

test('swipe resolver omits duplicate and archived IDs before active-first/JID ordering', () => {
    const candidates = [
        { chat_jid: 'gi:z', is_active: false },
        { chat_jid: 'gi:b', is_active: true, is_pinned: true },
        { chat_jid: 'gi:a', is_active: true },
        { chat_jid: 'gi:b', is_active: true }, // pinned/active/ordinary overlap
        { chat_jid: 'gi:c', is_active: false, is_pinned: true },
        { chat_jid: 'gi:z', is_active: false },
        { chat_jid: 'gi:archived', archived_at: '2026-09-23', is_active: true },
        { chat_jid: 'gi:archived', archived_at: '2026-09-23' },
    ];
    const order = ['gi:a', 'gi:b', 'gi:c', 'gi:z'];
    for (const current of order) {
        expect(resolveSwipeableChatAgents(candidates, current)).toEqual(order);
        const i = order.indexOf(current);
        expect(resolveAdjacentSwipeChatJid({ candidates, currentChatJid: current, direction: 'next' })).toBe(order[(i+1)%order.length]);
        expect(resolveAdjacentSwipeChatJid({ candidates, currentChatJid: current, direction: 'prev' })).toBe(order[(i+order.length-1)%order.length]);
    }
});
