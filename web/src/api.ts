// @ts-nocheck
import { notifyModelSettlement } from './gi-model-invalidation.js';
/**
 * api.ts — Gi API adapter.
 *
 * Implements the same exported function signatures as Piclaw's api.ts so
 * that all Piclaw components work without modification, but calls Gi's
 * REST endpoints underneath.
 *
 * Unimplemented endpoints return sensible empty/no-op responses rather
 * than throwing, so the UI degrades gracefully until Gi grows the feature.
 */

import { recordAppPerfRequest } from './ui/app-perf-tracing.js';
import { sessionPickerAgents } from './gi-session-state.js';
import { projectMessageMedia } from './gi-message-media.js';
import { projectLinkPreviews } from './gi-message-links.js';
import { composeTransfers } from './gi-compose-transfer.js';

const API_BASE = '';

// ── Core request helper ────────────────────────────────────────────────────

async function request(url: string, options: RequestInit = {}) {
    const startedAt = typeof performance !== 'undefined' && typeof performance.now === 'function'
        ? performance.now()
        : Date.now();
    let response: Response;
    try {
        response = await fetch(API_BASE + url, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                ...((options as any).headers || {}),
            },
        });
    } catch (error) {
        recordAppPerfRequest({
            method: String((options as any).method || 'GET').toUpperCase(),
            url, startedAt,
            durationMs: performance.now() - startedAt,
            ok: false,
            detail: { failedBeforeResponse: true },
        });
        throw error;
    }
    const durationMs = performance.now() - startedAt;
    recordAppPerfRequest({
        method: String((options as any).method || 'GET').toUpperCase(),
        url, startedAt, durationMs,
        status: response.status,
        ok: response.ok,
        requestId: response.headers?.get?.('x-request-id') || null,
        serverTiming: response.headers?.get?.('Server-Timing') || null,
    });
    if (!response.ok) {
        const err = await response.json().catch(() => ({ error: 'Unknown error' }));
        throw Object.assign(new Error(err.error || `HTTP ${response.status}`), {status:response.status});
    }
    return response.json();
}

// Same frozen component contract; conservative failure hides unsupported rows.
let quickActionsReady = false;
let quickActionsRevision = 0;
export const isQuickActionsReady = () => quickActionsReady;
export function resetQuickActionsReadiness() { quickActionsReady = false; ++quickActionsRevision; }
export async function getQuickActionsSettings() {
    const revision = ++quickActionsRevision;
    quickActionsReady = false;
    try { return { settings: await request('/api/quick-actions') }; }
    catch { return { settings: { workspaceCommands: [], slashCommands: [] } }; }
    finally { if (revision === quickActionsRevision) quickActionsReady = true; }
}
export async function getAgentCommands(_chatJid: string | null = null) {
    return request('/api/quick-actions').then(data => ({ commands: data.commands || [] }));
}

// ── SSE helper ────────────────────────────────────────────────────────────

function parseEventStreamBlock(block: string) {
    const lines = String(block || '').split('\n');
    let event = 'message';
    const dataLines: string[] = [];
    for (const line of lines) {
        if (line.startsWith('event:')) { event = line.slice(6).trim() || 'message'; }
        else if (line.startsWith('data:')) { dataLines.push(line.slice(5).trim()); }
    }
    const rawData = dataLines.join('\n');
    if (!rawData) return null;
    try { return { event, data: JSON.parse(rawData) }; }
    catch { return { event, data: rawData }; }
}

export async function consumeEventStream(response: Response, onEvent: (event: string, data: unknown) => void) {
    if (!response.body) throw new Error('Missing event stream body');
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    while (true) {
        const { value, done } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        const parts = buffer.split('\n\n');
        buffer = parts.pop() || '';
        for (const part of parts) {
            const parsed = parseEventStreamBlock(part);
            if (parsed) onEvent(parsed.event, parsed.data);
        }
    }
    buffer += decoder.decode();
    const tail = parseEventStreamBlock(buffer);
    if (tail) onEvent(tail.event, tail.data);
}

// ── Gi session → Piclaw chat_jid mapping ──────────────────────────────────
// Gi uses session IDs; Piclaw components use chat_jid. We use a fixed
// default JID and map Gi session IDs onto it for now.

const DEFAULT_CHAT_JID = 'web:default';

function sessionToChatJid(sessionId: string | null) {
    return sessionId ? `gi:${sessionId}` : DEFAULT_CHAT_JID;
}

// ── Timeline / posts ──────────────────────────────────────────────────────

export async function getTimeline(limit = 50, beforeId: string | null = null, chatJid: string | null = null, after: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) return { posts: [] };
    let url = `/api/sessions/${encodeURIComponent(sessionId)}/messages?limit=${limit}`;
    if (beforeId) url += `&before=${encodeURIComponent(beforeId)}`;
    if (after) url += `&after=${encodeURIComponent(after)}`;
    const data = await request(url);
    const messages: any[] = data.messages || [];
    return {
        hasMore: data.has_more === true, before: data.before || null, after: data.after || null,
        posts: messages.map((m: any) => ({
            id: m.id,
            chat_jid: chatJid,
            content: m.content,
            timestamp: m.created_at,
            sender: m.role === 'user' ? 'user' : 'agent',
            is_from_me: m.role === 'user',
            is_bot_message: m.role === 'assistant',
            data: {
                type: m.role === 'assistant' ? 'agent_response' : 'user_message',
                content: m.content,
                thread_id: null,
                agent_id: m.payload?.agent_id || (m.role === 'assistant' ? 'agent' : null),
                ...projectMessageMedia(m.payload, sessionId),
                content_meta: null,
                link_previews: projectLinkPreviews(m.payload),
                kind: m.payload?.kind || null,
                source: m.payload?.source || null,
                clipped: m.payload?.clipped || false,
            },
        })),
    };
}

export async function getPostsByHashtag(_hashtag: string, _limit = 50, _offset = 0, _chatJid: string | null = null) {
    return { posts: [] };
}

export async function searchPosts(query: string, limit = 50, offset = 0, chatJid: string | null = null, scope = 'current', _rootChatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) return { posts: [] };
    const params=new URLSearchParams({q:query,scope,limit:String(limit),offset:String(offset)});
    const data = await request(`/api/sessions/${encodeURIComponent(sessionId)}/search?${params}`);
    const messages: any[] = data.messages || [];
    return { posts: messages.map((m: any) => ({
        id: m.id, chat_jid: sessionToChatJid(m.session_id), content: m.content, timestamp: m.created_at,
        sender: m.role === 'user' ? 'user' : 'agent',
        is_from_me: m.role === 'user', is_bot_message: m.role === 'assistant',
        data: { type: m.role === 'assistant' ? 'agent_response' : 'user_message', content: m.content, thread_id: null, agent_id: m.payload?.agent_id || (m.role === 'assistant' ? 'agent' : null), ...projectMessageMedia(m.payload, m.session_id), link_previews: projectLinkPreviews(m.payload) },
    })) };
}

export async function getThread(threadId: number, _chatJid: string | null = null) {
    return { posts: [] };
}

// ── Agent / status ────────────────────────────────────────────────────────

export async function getSystemMetrics() {
    return request('/api/system-metrics').catch(() => null);
}

export async function getAgents() {
    const data = await request('/api/sessions').catch(() => ({ sessions: [] }));
    const runtime = await request('/api/runtime/config').catch(() => ({}));
    const sessions: any[] = data.sessions || [];
    const agents = new Map();
    for (const s of sessions) {
        const id = s.scope?.agent_id || 'agent';
        if (!agents.has(id)) {
            agents.set(id, {
                id,
                name: s.title || `@${id}`,
                avatar_url: runtime.assistant_avatar || null,
                chat_jid: sessionToChatJid(s.id),
            });
        }
    }
    if (agents.size === 0) {
        agents.set('agent', { id: 'agent', name: runtime.assistant_name || '@agent', avatar_url: runtime.assistant_avatar || null, chat_jid: DEFAULT_CHAT_JID });
    }
    return { agents: Array.from(agents.values()) };
}

export async function getAgentStatus(agentId: string, chatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) return null;
    const data = await request(`/api/sessions/${encodeURIComponent(sessionId)}/activity`);
    return { ...data, type: data.status === 'running' ? 'tool_call' : 'intent',
        title: data.status === 'cancelling' ? 'Cancelling…' : data.status === 'running' ? 'Working…' : '' };
}

export async function getSessionCompaction(chatJid: string) {
    if (!chatJid?.startsWith('gi:')) return {available:false};
    return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/compaction`);
}
export async function compactSession(chatJid: string, token: string) {
    if (!chatJid?.startsWith('gi:') || !token) throw new Error('No compaction snapshot');
    return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/compaction`, {method:'POST',body:JSON.stringify({token})});
}

export async function cancelSessionRun(chatJid: string, turnId: string) {
    if (!chatJid?.startsWith('gi:') || !turnId) throw new Error('No active run to stop');
    return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/activity`, {method:'POST', body:JSON.stringify({turn_id:turnId})});
}

export async function getAgentContext(_agentId: string, chatJid: string | null = null) {
    if (!chatJid?.startsWith('gi:')) return null;
    return (await getAgentModels(chatJid)).context_usage || null;
}

export async function getAgentThought(_agentId: string, _chatJid: string | null = null) {
    return null;
}

export async function setAgentThoughtVisibility(_agentId: string, _visible: boolean, _chatJid: string | null = null) {
    return null;
}

export async function getGiProviders() {
    return request('/api/settings/providers');
}

export async function saveGiProviderKey(provider: string, revision: string, key: string) {
    return request('/api/settings/providers', { method: 'PATCH', body: JSON.stringify({ provider, revision, key }) });
}

export async function removeGiProviderKey(provider: string, revision: string) {
    return request('/api/settings/providers', { method: 'DELETE', body: JSON.stringify({ provider, revision }) });
}

export async function getGiCompactionPolicy() {
    return request('/api/settings/compaction');
}

export async function saveGiCompactionPolicy(value: any) {
    return request('/api/settings/compaction', { method: 'PATCH', body: JSON.stringify(value) });
}

export async function getGiIdentity() {
    return request('/api/settings/identity');
}

export async function saveGiIdentity(value: { revision: string; assistant_name: string; user_name: string }) {
    return request('/api/settings/identity', { method: 'PATCH', body: JSON.stringify(value) });
}

export async function getGiSettingsSnapshot() {
    return request('/api/runtime/config');
}

export async function getAgentModels(chatJid: string | null = null) {
    if (chatJid?.startsWith('gi:')) return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/model`);
    const data = await request('/api/runtime/config');
    const modelOptions = Array.isArray(data.model_options) ? data.model_options : [];
    const models: any[] = modelOptions.length > 0
        ? modelOptions
        : (data.enabled_models || []).map((id: string) => ({
            id,
            provider: data.default_provider || '',
            label: id,
        }));
    return {
        models,
        model_options: modelOptions,
        provider_options: Array.isArray(data.provider_options) ? data.provider_options : [],
        current: data.current || data.default_model || '',
        thinking_level: data.default_thinking_level || data.thinking_level || '',
        supports_thinking: Boolean(data.supports_thinking),
    };
}

export async function selectAgentModel(chatJid: string, model: string) {
    if (!chatJid?.startsWith('gi:')) throw new Error('No model destination session');
    try {
        return await request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/model`, { method: 'PATCH', body: JSON.stringify({ model }) });
    } finally {
        // A transport failure can follow a committed native write. Notify only
        // the captured session; consumers reread rather than replay responses.
        notifyModelSettlement(chatJid);
    }
}

export async function getAgentQueueState(chatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) return { items: [] };
    const data = await request(`/api/sessions/${encodeURIComponent(sessionId)}/queue`);
    return { activeTurnId: data.active_turn_id || null, items: (data.items || []).map((turn: any) => ({
        id: turn.id, text: turn.prompt, content: turn.prompt, chat_jid: chatJid,
        metadata: turn.metadata, created_at: turn.created_at, phase: turn.phase,
    })) };
}

export async function steerAgentQueueItem(itemId: string, chatJid: string, activeTurnId: string) {
    if (!chatJid?.startsWith('gi:') || !activeTurnId) throw new Error('Steer requires a matching active run');
    return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/queue/${encodeURIComponent(itemId)}/steer`, {
        method: 'POST', body: JSON.stringify({ active_turn_id: activeTurnId }),
    });
}

export async function removeAgentQueueItem(turnId: string, chatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) throw new Error('No queue session');
    return request(`/api/sessions/${encodeURIComponent(sessionId)}/queue/${encodeURIComponent(turnId)}`, { method: 'DELETE' });
}

export async function getAutoresearchStatus(_chatJid: string | null = null) {
    return null;
}

export async function stopAutoresearch(_chatJid: string | null = null) {
    return null;
}

export async function dismissAutoresearch(_chatJid: string | null = null) {
    return null;
}

export async function getActiveChatAgents() {
    const data = await request('/api/sessions');
    const sessions: any[] = data.sessions || [];
    return {
        agents: sessionPickerAgents(sessions),
    };
}

export async function getChatBranches(rootChatJid: string | null = null, _options: any = {}) {
    const data = await request('/api/sessions').catch(() => ({ sessions: [] }));
    let sessions: any[] = data.sessions || [];
    if (rootChatJid?.startsWith('gi:')) {
        const rootId = rootChatJid.slice(3);
        const byParent = new Map();
        for (const s of sessions) {
            const key = s.parent_session_id || '';
            const bucket = byParent.get(key) || [];
            bucket.push(s);
            byParent.set(key, bucket);
        }
        const wanted = new Set([rootId]);
        const queue = [rootId];
        while (queue.length > 0) {
            const current = queue.shift();
            const children = byParent.get(current) || [];
            for (const child of children) {
                if (!wanted.has(child.id)) {
                    wanted.add(child.id);
                    queue.push(child.id);
                }
            }
        }
        sessions = sessions.filter((s: any) => wanted.has(s.id));
    }
    const mapped = sessions.map((s: any) => ({
        chat_jid: sessionToChatJid(s.id),
        label: s.title || `@${s.scope?.agent_id || s.id}`,
        updated_at: s.updated_at,
        parent_chat_jid: s.parent_session_id ? sessionToChatJid(s.parent_session_id) : null,
        agent_id: s.scope?.agent_id || 'agent',
    }));
    return { branches: mapped, chats: mapped };
}

export async function forkChatBranch(sourceChatJid: string, options: any = {}) {
    const sessionId = sourceChatJid?.startsWith('gi:') ? sourceChatJid.slice(3) : null;
    if (!sessionId) throw new Error('No source session to fork');
    return request(`/api/sessions/${encodeURIComponent(sessionId)}/fork`, {
        method: 'POST',
        body: JSON.stringify({ title: options?.title || null, agent_id: options?.agent_id || null }),
    });
}

async function mutateChatSession(chatJid: string, mutation: any) {
    if (!chatJid?.startsWith('gi:') || !chatJid.slice(3)) throw new Error('Invalid session identifier');
    return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}`, {
        method: 'PATCH', body: JSON.stringify(mutation),
    });
}

export async function renameChatBranch(chatJid: string, options: any = {}) {
    return mutateChatSession(chatJid, { action: 'rename', title: options.title });
}

export async function pinChatSession(chatJid: string, pinned: boolean) {
    return mutateChatSession(chatJid, { action: 'pin', pinned });
}

export async function pruneChatBranch(chatJid: string) {
    return mutateChatSession(chatJid, { action: 'archive' });
}

export async function restoreChatBranch(chatJid: string, _options: any = {}) {
    return mutateChatSession(chatJid, { action: 'restore' });
}

export async function renameChatJid(_oldJid: string, _newJid: string) {
    return null;
}

export async function sendPeerAgentMessage(sourceChatJid: string, target: string, content: string, mode = 'auto', options: any = {}) {
    const sessionId = sourceChatJid?.startsWith('gi:') ? sourceChatJid.slice(3) : null;
    if (!sessionId) throw new Error('No source session');
    return request(`/api/sessions/${encodeURIComponent(sessionId)}/peer-message`, {
        method: 'POST',
        body: JSON.stringify({
            target_agent_id: String(target || '').replace(/^@/, ''),
            content,
            mode,
            model: options?.model || null,
            parent_turn_id: options?.parent_turn_id || null,
        }),
    });
}

export async function completeInstanceOobe(_chatJid: string | null = null) {
    return null;
}

// ── Posts / messages ──────────────────────────────────────────────────────

export async function createPost(content: string, _mediaIds: number[] = [], chatJid: string | null = null, options: any = {}) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) throw new Error('No active session');
    const payload: any = { prompt: content, intent: 'prompt' };
    if (options?.parent_turn_id) {
        payload.parent_turn_id = options.parent_turn_id;
    }
    return request(`/api/sessions/${encodeURIComponent(sessionId)}/prompt`, {
        method: 'POST',
        body: JSON.stringify(payload),
    });
}

export async function createReply(threadId: number, content: string, _mediaIds: number[] = [], chatJid: string | null = null, options: any = {}) {
    return createPost(content, [], chatJid, options);
}

export async function deletePost(postId: string, cascade = false, chatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) throw new Error('No message destination session');
    return request(`/api/sessions/${encodeURIComponent(sessionId)}/messages/${encodeURIComponent(postId)}?cascade=${cascade}`, { method: 'DELETE' });
}

export async function sendAgentMessage(agentId: string, content: string, _threadId: number | null = null, _mediaIds: number[] = [], mode: string | null = null, chatJid: string | null = null, options: any = {}) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) throw new Error('No active session');
    const intent = mode === 'steer' ? 'steer' : mode === 'queue' ? 'queue' : 'prompt';
    const targetAgentId = agentId && agentId !== 'default' ? String(agentId).replace(/^@/, '') : null;
    const payload: any = {
        prompt: content,
        intent,
        target_agent_id: targetAgentId,
        media: _mediaIds.map(media_id => ({ media_id, session_id: sessionId })),
        client_request_id: options?.client_request_id || undefined,
    };
    if (options?.parent_turn_id) {
        payload.parent_turn_id = options.parent_turn_id;
    }
    const activity = composeTransfers.begin(sessionId, 'send');
    try {
        return await request(`/api/sessions/${encodeURIComponent(sessionId)}/prompt`, {
            method: 'POST', body: JSON.stringify(payload),
        });
    } finally { activity.end(); }
}

export async function streamSidePrompt(content: string, chatJid: string | null = null, _options: any = {}) {
    return null;
}

// ── Media ─────────────────────────────────────────────────────────────────

export async function uploadMedia(file: File, chatJid: string | null = null, options: { signal?: AbortSignal } = {}) {
    const signal = options.signal;
    signal?.throwIfAborted();
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) throw new Error('No attachment destination session');
    if (file.size > 10 * 1024 * 1024) throw new Error('Media exceeds 10 MiB limit');
    const activity = composeTransfers.begin(sessionId, 'upload');
    try {
        const form = new FormData();
        form.append('file', file, file.name);
        // Materialise the browser-generated multipart body within the existing
        // 10 MiB file bound. This also avoids WebKit file-backed XHR bodies
        // becoming empty when a service/automation interceptor forwards them.
        const encoded = new Response(form);
        const contentType = encoded.headers.get('content-type');
        const body = await encoded.arrayBuffer();
        signal?.throwIfAborted();
        // XHR exposes real upload bytes; fetch does not.
        return await new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();
            let settled = false;
            const abort = () => { finish(new DOMException('Upload aborted', 'AbortError')); xhr.abort(); };
            const finish = (error?: unknown, value?: unknown) => {
                if (settled) return; settled = true;
                signal?.removeEventListener('abort', abort);
                xhr.upload.onprogress = null;
                error ? reject(error) : resolve(value);
            };
            signal?.addEventListener('abort', abort, { once: true });
            if (signal?.aborted) { abort(); return; }
            try {
                xhr.open('POST', `/api/sessions/${encodeURIComponent(sessionId)}/media`);
                xhr.setRequestHeader('Content-Type', contentType);
            } catch (error) { finish(error); return; }
            xhr.upload.onprogress = event => activity.progress(event.loaded, event.total, event.lengthComputable);
            xhr.onerror = () => finish(new TypeError('Upload network request failed'));
            xhr.onabort = () => finish(new DOMException('Upload aborted', 'AbortError'));
            xhr.onload = () => {
                let data = {};
                try { data = JSON.parse(xhr.responseText); } catch {}
                if (xhr.status < 200 || xhr.status >= 300) { finish(new Error(data.error || `Upload failed: HTTP ${xhr.status}`)); return; }
                if (!data.media?.id) { finish(new Error('Upload returned no media identifier')); return; }
                finish(undefined, { ...data.media, id: data.media.id });
            };
            try { xhr.send(body); } catch (error) { finish(error); }
        });
    } finally { activity.end(); }
}

export async function getMediaInfo(mediaId: number) {
    return request(`/api/media/${mediaId}`).catch(() => null);
}

export function getMediaUrl(mediaId: number) {
    return `/api/media/${mediaId}/raw`;
}

export function getThumbnailUrl(mediaId: number) {
    // Native storage has no derived thumbnails; keep the original image bytes.
    return getMediaUrl(mediaId);
}

export async function submitAdaptiveCardAction(_payload: unknown) {
    // Rendering is supported; accepting card actions requires a native identity,
    // authorization and persistence contract. Never acknowledge a no-op submit.
    throw new Error('Card submissions are not supported by Gi yet. Your inputs have not been submitted.');
}

// ── Workspace ─────────────────────────────────────────────────────────────

export async function getSessionRouteEvents(chatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) return { route_events: [] };
    return request(`/api/sessions/${encodeURIComponent(sessionId)}/route-events`);
}

export async function getWorkspaceTree(path = '', depth = 1, showHidden = false) {
    const query = new URLSearchParams({path: path || '.', depth: String(depth), show_hidden: String(showHidden)});
    return { root: await request(`/api/workspace/tree?${query}`) };
}

export async function getWorkspaceFile(path: string, maxBytes = 20000) {
    return request(`/api/workspace/file?path=${encodeURIComponent(path)}&max_bytes=${maxBytes}`);
}

export async function getWorkspaceIndexStatus(scope = 'all') {
    return request(`/api/workspace/index?scope=${encodeURIComponent(scope)}`);
}

export async function reindexWorkspace(scope = 'all') {
    return request(`/api/workspace/index?scope=${encodeURIComponent(scope)}`, { method: 'POST' });
}

export async function createWorkspaceFile(path: string, content: string, _chatJid: string | null = null) {
    return request('/api/workspace/file', { method: 'POST', body: JSON.stringify({ path, content }) }).catch(() => null);
}

export async function renameWorkspaceFile(_oldPath: string, _newPath: string, _chatJid: string | null = null) {
    return null;
}

export async function moveWorkspaceEntry(_from: string, _to: string, _chatJid: string | null = null) {
    return null;
}

export async function deleteWorkspaceFile(_path: string, _chatJid: string | null = null) {
    return null;
}

export async function uploadWorkspaceFile(_path: string, _file: File, _chatJid: string | null = null) {
    return null;
}

export async function setWorkspaceVisibility(visible: boolean, showHidden: boolean) {
    // Gi has no workspace push subscription to reconfigure. The explorer owns
    // local persistence and supplies this flag on every subsequent tree pull.
    return { visible, show_hidden: showHidden };
}

export function getWorkspaceDownloadUrl(path: string) {
    return `/api/workspace/file?path=${encodeURIComponent(path)}`;
}

export function getWorkspaceFileDownloadUrl(path: string) {
    return `/api/workspace/file?path=${encodeURIComponent(path)}`;
}

export async function getWorkspaceBranch(_chatJid: string | null = null) {
    return null;
}

// ── Push notifications ────────────────────────────────────────────────────

export async function getWebPushPublicKey() { return null; }
export async function saveWebPushSubscription(_sub: unknown, _opts: any = {}) { return null; }
export async function deleteWebPushSubscription(_sub: unknown, _opts: any = {}) { return null; }

// ── Agent whitelist / ACP ─────────────────────────────────────────────────

export async function addToWhitelist(_target: string, _chatJid: string | null = null) { return null; }
export async function respondToAgentRequest(_requestId: string, _allow: boolean, _chatJid: string | null = null) { return null; }

// ── Performance tracing stub (consumed by this file itself) ───────────────
// app-perf-tracing.ts in Piclaw is a real module; we provide a no-op here
// only if the import fails — but since we copied the file it should resolve.

// ── Additional exports required by Piclaw components ─────────────────────

export async function reorderAgentQueueItem(payload: { chatJid: string; expected: string[]; order: string[] }) {
    if (!payload.chatJid?.startsWith('gi:')) throw new Error('No queue session');
    return request(`/api/sessions/${encodeURIComponent(payload.chatJid.slice(3))}/queue`, { method: 'PATCH', body: JSON.stringify({ expected: payload.expected, order: payload.order }) });
}

export function getWorkspaceRawUrl(path: string, options: any = {}) {
    const q = new URLSearchParams({ path: String(path || '') });
    if (options?.download) q.set('download', '1');
    return `/api/workspace/raw?${q.toString()}`;
}

export function getWorkspaceFileDownloadUrl(path: string) {
    return getWorkspaceRawUrl(path, { download: true });
}

export async function getWorkspacePreviewContent(_path: string, _chatJid: string | null = null) { return null; }
export async function createWorkspaceFolder(_path: string, _chatJid: string | null = null) { return null; }
export async function recordAppPerfRequest(_payload: unknown) {}

export { SSEClient } from './gi-sse-client.js';

export async function getWorkspaceFileStat(_path: string, _chatJid: string | null = null) { return null; }
export async function getMediaBlob(..._args: any[]) { return null; }
