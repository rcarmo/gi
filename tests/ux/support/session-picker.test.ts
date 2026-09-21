import { expect, test } from 'bun:test';
import { sessionPickerAgents } from '../../../web/src/gi-session-state';
import { filterSessionPickerChats, groupSessionPickerChats, moveSessionPickerIndex, resolveSessionPickerSearchInitialIndex } from '../../../web/src/ui/compose-session-switcher';

const sessions = [
  { id: 'root', title: '@root' },
  { id: 'child', title: '@child', parent_session_id: 'root' },
  { id: 'leaf', title: '@leaf', parent_session_id: 'child', state: { model: 'test-model' } },
  { id: 'other', title: '@other' },
];

test('native ancestry maps nested children to the true root and preserves searchable metadata', () => {
  const chats = sessionPickerAgents(sessions);
  expect(chats[2]).toMatchObject({ root_chat_jid: 'gi:root', parent_branch_id: 'child', model: 'test-model' });
  expect(groupSessionPickerChats(chats, 'gi:child').map(group => [group.key, group.items.map(chat => chat.chat_jid)])).toEqual([
    ['current', ['gi:child']], ['tree', ['gi:root', 'gi:leaf']], ['other', ['gi:other']],
  ]);
  expect(sessionPickerAgents([{ id: 'orphan', parent_session_id: 'missing' }])[0].root_chat_jid).toBe('gi:orphan');
  expect(sessionPickerAgents([{ id: 'cycle', parent_session_id: 'cycle' }])[0].root_chat_jid).toBe('gi:cycle');
});

test('pinned Piclaw grouping precedence and ordering remain intact (helper contract, not live mutation evidence)', () => {
  const chats = [
    { chat_jid: 'current', root_chat_jid: 'root', is_active: true },
    { chat_jid: 'pin', is_active: true },
    { chat_jid: 'active', root_chat_jid: 'root', is_active: true },
    { chat_jid: 'tree', root_chat_jid: 'root' },
    { chat_jid: 'other' },
    { chat_jid: 'archive', is_active: true, archived_at: '2026-09-21' },
  ];
  const groups = groupSessionPickerChats(chats, 'current', ['current', 'pin', 'archive']);
  expect(groups.map(group => group.key)).toEqual(['current', 'pinned', 'active', 'tree', 'other', 'archived']);
  expect(groups.flatMap(group => group.items)).toEqual(chats);
});

test('search is case-insensitive and all-term, preserves ancestors, and highlights a handle match', () => {
  const chats = sessionPickerAgents(sessions);
  const filtered = filterSessionPickerChats(chats, 'LEAF TEST-MODEL');
  expect(filtered.map(chat => chat.chat_jid)).toEqual(['gi:root', 'gi:child', 'gi:leaf']);
  expect(resolveSessionPickerSearchInitialIndex(filtered, 'leaf')).toBe(2);
  expect(filterSessionPickerChats(chats, 'gi:other')).toEqual([chats[3]]);
  expect(filterSessionPickerChats(chats, 'leaf absent')).toEqual([]);
  expect(filterSessionPickerChats(chats, '   ')).toEqual(chats);
});

test('native mutation capabilities and activity never depend on selected chat identity', () => {
  const [main, busy, archived] = sessionPickerAgents([
    { id: 'main', state: { status: 'idle', pinned: true } },
    { id: 'busy', parent_session_id: 'main', state: { status: 'queued', queue_count: 1 } },
    { id: 'archived', parent_session_id: 'main', state: { archived_at: '2026-09-21', pinned: true } },
  ]);
  expect(main).toMatchObject({ is_active: false, pinned: true, capabilities: { archive: false, pin: true, rename: true, restore: false } });
  expect(busy.is_active).toBe(true);
  expect(archived.capabilities).toEqual({ archive: false, pin: false, rename: false, restore: true });
  expect(groupSessionPickerChats([main, busy, archived], 'gi:main', ['gi:archived']).map(group => group.key)).toEqual(['current', 'active', 'archived']);
});

test('picker navigation wraps arrows and bounds page movement and empty lists', () => {
  expect(moveSessionPickerIndex(0, 3, 'ArrowUp')).toBe(2);
  expect(moveSessionPickerIndex(2, 3, 'ArrowDown')).toBe(0);
  expect(moveSessionPickerIndex(0, 20, 'PageDown')).toBe(8);
  expect(moveSessionPickerIndex(2, 20, 'PageUp')).toBe(0);
  expect(moveSessionPickerIndex(0, 3, 'End')).toBe(2);
  expect(moveSessionPickerIndex(2, 3, 'Home')).toBe(0);
  expect(moveSessionPickerIndex(4, 0, 'ArrowDown')).toBe(0);
});
