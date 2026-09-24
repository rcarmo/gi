// Native browser SSE adapter: each transport instance owns its callbacks.
const API_BASE = "";
export class SSEClient {
    onEvent: any;
    onStatusChange: any;
    chatJid: string | null;
    eventSource: EventSource | null;
    reconnectTimeout: any;
    reconnectDelay: number;
    status: string;
    connecting: boolean;
    staleMonitor: any;

    constructor(onEvent: any, onStatusChange: any, options: any = {}) {
        this.onEvent = onEvent;
        this.onStatusChange = onStatusChange;
        this.chatJid = typeof options?.chatJid === 'string' && options.chatJid.trim() ? options.chatJid.trim() : null;
        this.eventSource = null;
        this.reconnectTimeout = null;
        this.reconnectDelay = 1000;
        this.status = 'disconnected';
        this.connecting = false;
        this.staleMonitor = null;
    }

    connect() {
        if (this.connecting) return;
        if (this.eventSource && this.status === 'connected') return;
        this.connecting = true;
        if (this.eventSource) this.eventSource.close();
        this.clearStaleMonitor();

        const query = this.chatJid ? `?chat_jid=${encodeURIComponent(this.chatJid)}` : '';
        const source = new EventSource(API_BASE + '/sse/stream' + query);
        this.eventSource = source;
        const current = () => this.eventSource === source;

        const bindJsonEvent = (eventType: string) => {
            source.addEventListener(eventType, (e: any) => {
                if (!current()) return;
                this.resetStaleMonitor();
                try {
                    const data = JSON.parse(e.data);
                    this.onEvent(eventType, data);
                } catch {}
            });
        };

        source.addEventListener('connected', (event: any) => {
            if (!current()) return;
            this.connecting = false;
            this.reconnectDelay = 1000;
            this.setStatus('connected');
            this.resetStaleMonitor();
            try { this.onEvent('connected', JSON.parse(event.data)); } catch {}
        });

        source.addEventListener('heartbeat', () => {
            if (!current()) return;
            this.resetStaleMonitor();
        });

        bindJsonEvent('new_post');
        bindJsonEvent('new_reply');
        bindJsonEvent('agent_response');
        bindJsonEvent('interaction_updated');
        bindJsonEvent('interaction_deleted');
        bindJsonEvent('agent_status');
        bindJsonEvent('agent_steer_queued');
        bindJsonEvent('agent_followup_queued');
        bindJsonEvent('agent_followup_consumed');
        bindJsonEvent('agent_followup_removed');
        bindJsonEvent('queue_changed');
        bindJsonEvent('tool_activity_changed');
        for (const event of ['compaction_started', 'compaction_completed', 'compaction_cancelled', 'compaction_suppressed', 'compaction_failed']) bindJsonEvent(event);
        bindJsonEvent('workspace_update');
        bindJsonEvent('agent_draft');
        bindJsonEvent('agent_draft_delta');
        bindJsonEvent('agent_thought');
        bindJsonEvent('agent_thought_delta');
        bindJsonEvent('routing_decision');
        bindJsonEvent('routing_incoming');
        bindJsonEvent('model_changed');
        bindJsonEvent('ui_theme');
        bindJsonEvent('ui_meters');

        source.onerror = () => {
            if (!current()) return;
            this.eventSource = null;
            source.close();
            this.clearStaleMonitor();
            this.connecting = false;
            this.setStatus('disconnected');
            this.scheduleReconnect();
        };
    }

    disconnect() {
        this.connecting = false;
        this.clearStaleMonitor();
        if (this.reconnectTimeout) { clearTimeout(this.reconnectTimeout); this.reconnectTimeout = null; }
        const source = this.eventSource;
        this.eventSource = null;
        source?.close();
        this.setStatus('disconnected');
    }

    reconnectIfNeeded() {
        if (this.status !== 'connected') this.connect();
    }

    forceReconnect() {
        this.disconnect();
        this.connect();
    }

    setStatus(status: string) {
        if (this.status === status) return;
        this.status = status;
        this.onStatusChange?.(status);
    }

    scheduleReconnect() {
        if (this.reconnectTimeout) return;
        this.reconnectTimeout = setTimeout(() => {
            this.reconnectTimeout = null;
            this.reconnectDelay = Math.min(this.reconnectDelay * 1.5, 30000);
            this.connect();
        }, this.reconnectDelay);
    }

    resetStaleMonitor() {
        this.clearStaleMonitor();
        this.staleMonitor = setTimeout(() => {
            this.setStatus('stale');
            this.forceReconnect();
        }, 60000);
    }

    clearStaleMonitor() {
        if (this.staleMonitor) { clearTimeout(this.staleMonitor); this.staleMonitor = null; }
    }
}
