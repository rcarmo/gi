import { test, expect } from 'bun:test';
import { SSEClient } from '../../../web/src/gi-sse-client';

class FakeSource {
  static sources: FakeSource[] = [];
  handlers = new Map<string, Function>();
  onerror: Function | null = null;
  closed = false;
  constructor(public url: string) { FakeSource.sources.push(this); }
  addEventListener(type: string, handler: Function) { this.handlers.set(type, handler); }
  close() { this.closed = true; }
  emit(type: string, data: any = {}) { this.handlers.get(type)?.({ data: JSON.stringify(data) }); }
}

test('closed and superseded SSE sources cannot deliver status/data or create reconnect timers', () => {
  const old = globalThis.EventSource;
  globalThis.EventSource = FakeSource as any;
  const events: string[] = [], statuses: string[] = [];
  const client = new SSEClient(type => events.push(type), status => statuses.push(status), { chatJid: 'gi:A' });
  try {
    client.connect(); const first = FakeSource.sources.at(-1)!;
    // Reconnect while still connecting must create a fresh connection.
    client.forceReconnect(); const second = FakeSource.sources.at(-1)!;
    expect(second).not.toBe(first);
    first.emit('connected'); first.emit('agent_draft_delta'); first.onerror?.();
    expect(events).toEqual([]); expect(statuses).toEqual([]); expect(client.reconnectTimeout).toBeNull();
    second.emit('connected'); second.emit('queue_changed');
    expect(events).toEqual(['queue_changed']); expect(statuses).toEqual(['connected']);
    second.onerror?.(); expect(second.closed).toBe(true); expect(client.status).toBe('disconnected');
    second.emit('agent_draft_delta'); expect(events).toEqual(['queue_changed']);
    client.disconnect(); expect(client.reconnectTimeout).toBeNull();
    second.onerror?.(); expect(client.reconnectTimeout).toBeNull();
  } finally { client.disconnect(); globalThis.EventSource = old; }
});
