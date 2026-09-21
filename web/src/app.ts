// @ts-nocheck
/**
 * app.ts — Gi entry point.
 *
 * Uses Piclaw's web components verbatim. The shell structure mirrors
 * app-main-shell-render.ts exactly: app-shell > container > timeline +
 * status + compose. No session creation UI — a default session is
 * auto-created on startup, matching Piclaw's always-ready UX.
 */
import { html, render, useState, useEffect, useMemo, useCallback, useRef } from './vendor/preact-htm.js';
import { getLocalStorageItem, setLocalStorageItem } from './utils/storage.js';
import { dedupePosts } from './ui/timeline-utils.js';
import { appendUniqueTimelinePost } from './ui/app-realtime-timeline.js';
import { useAgentState } from './ui/use-agent-state.js';
import { useSseConnection } from './ui/use-sse-connection.js';
import { handleAppSseEvent } from './ui/app-sse-events.js';
import { initTheme } from './ui/theme.js';
import { installPwaDisplayScaleSync } from './ui/pwa-display-scale.js';
import {
    LAST_ACTIVITY_TTL_MS,
    SILENCE_FINALIZE_MS,
    SILENCE_REFRESH_MS,
    SILENCE_WARNING_MS,
    isIOSDevice,
} from './ui/app-helpers.js';
import { isCompactionStatus } from './ui/status-duration.js';
import { paneRegistry, tabStore } from './panes/index.js';
import {
    getTimeline,
    searchPosts,
    deletePost,
    getAgents,
    getAgentThought,
    setAgentThoughtVisibility,
    getAgentStatus,
    getAgentContext,
    getAutoresearchStatus,
    stopAutoresearch,
    dismissAutoresearch,
    getAgentModels,
    completeInstanceOobe,
    getActiveChatAgents,
    getChatBranches,
    renameChatBranch,
    pruneChatBranch,
    restoreChatBranch,
    getAgentQueueState,
    steerAgentQueueItem,
    removeAgentQueueItem,
    streamSidePrompt,
    getWorkspaceFile,
    sendAgentMessage,
    forkChatBranch,
    getThread,
} from './api.js';
import { Timeline } from './components/timeline.js';
import { ComposeBox, QueuedFollowupStack } from './components/compose-box.js';
import { AgentStatus, AgentRequestModal } from './components/status.js';
import { WorkspaceExplorer } from './components/workspace-explorer.js';
import { TabStrip } from './components/tab-strip.js';
import { FloatingWidgetPane } from './components/floating-widget-pane.js';
import { AttachmentPreviewModal } from './components/attachment-preview-modal.js';
import { SystemMetersHud } from './components/system-meters-hud.js';
import { TimelineMenu } from './components/timeline-menu.js';
import { createSelectionScope } from './gi-session-state.js';

// ── Gi session bridge ──────────────────────────────────────────────────────
// Piclaw components expect chat_jid strings. We map Gi sessions onto that
// model: the default session becomes 'gi:default'.

const DEFAULT_SESSION_TITLE = 'default';
const SESSION_KEY = 'gi_session_id';
const POLL_INTERVAL_MS = 1200;
const DEFAULT_AGENT_ID = 'web';

function sessionToChatJid(id: string) {
    return `gi:${id}`;
}

async function ensureDefaultSession() {
    const stored = getLocalStorageItem(SESSION_KEY);
    if (stored) {
        // Verify it still exists
        try {
            const r = await fetch(`/api/sessions/${encodeURIComponent(stored)}`);
            if (r.ok) return stored;
        } catch {}
    }

    try {
        const existing = await fetch('/api/sessions');
        if (existing.ok) {
            const payload = await existing.json();
            const sessions = Array.isArray(payload?.sessions) ? payload.sessions : [];
            const matching = sessions
                .filter((session: any) => session?.scope?.agent_id === 'web' && !session?.parent_session_id)
                .sort((a: any, b: any) => String(b?.updated_at || '').localeCompare(String(a?.updated_at || '')));
            if (matching[0]?.id) {
                setLocalStorageItem(SESSION_KEY, matching[0].id);
                return matching[0].id;
            }
        }
    } catch {}

    // Create a new default web session
    const r = await fetch('/api/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: '@web', agent_id: 'web' }),
    });
    if (!r.ok) throw new Error('Failed to create default session');
    const s = await r.json();
    setLocalStorageItem(SESSION_KEY, s.id);
    return s.id;
}

async function getRuntimeConfig() {
    const r = await fetch('/api/runtime/config');
    if (!r.ok) return {};
    return r.json();
}

// ── App ────────────────────────────────────────────────────────────────────

function GiApp() {
    const [ready, setReady] = useState(false);
    const [sessionId, setSessionId] = useState<string | null>(null);
    const selection = useRef(createSelectionScope()).current;
    const drafts = useRef(new Map<string, any>()).current;
    const [sessionError, setSessionError] = useState<string | null>(null);
    const getDraft = (sid: string) => {
        if (!drafts.has(sid)) drafts.set(sid, { text: '', media: [], fileRefs: [], messageRefs: [] });
        return drafts.get(sid);
    };
    const [runtimeConfig, setRuntimeConfig] = useState<any>({});
    const [agents, setAgents] = useState<any>({});
    const [userProfile, setUserProfile] = useState<any>(null);

    // Workspace / pane state
    const [workspaceOpen, setWorkspaceOpen] = useState(false);
    const [tabs, setTabs] = useState<any[]>([]);
    const [activeTabId, setActiveTabId] = useState<string | null>(null);
    const editorOpen = tabs.length > 0;

    // Timeline
    const [posts, setPosts] = useState<any[]>([]);
    const [hasMore, setHasMore] = useState(false);
    const timelineRef = useRef<any>(null);

    // Compose
    const [fileRefs, setFileRefs] = useState<string[]>([]);
    const [messageRefs, setMessageRefs] = useState<any[]>([]);
    const [followupQueueItems, setFollowupQueueItems] = useState<any[]>([]);
    const [floatingWidget, setFloatingWidget] = useState<any>(null);
    const [attachmentPreview, setAttachmentPreview] = useState<any>(null);
    const [contextUsage, setContextUsage] = useState<any>(null);
    const [activeChatAgents, setActiveChatAgents] = useState<any[]>([]);
    const [currentChatBranches, setCurrentChatBranches] = useState<any[]>([]);
    const [activeModel, setActiveModel] = useState<string>('');
    const [agentModelsPayload, setAgentModelsPayload] = useState<any>(null);
    const [activeThinkingLevel, setActiveThinkingLevel] = useState<string>('');
    const [supportsThinking, setSupportsThinking] = useState(false);
    const [modelUsage, setModelUsage] = useState<any>(null);
    const [connectionStatus, setConnectionStatus] = useState<string>('connected');
    const [isAgentTurnActive, setIsAgentTurnActive] = useState(false);
    const isAgentRunningRef = useRef(false);

    // Agent state (matches Piclaw's useAgentState shape)
    const {
        agentStatus, setAgentStatus,
        agentDraft, setAgentDraft,
        agentPlan, setAgentPlan,
        agentThought, setAgentThought,
        pendingRequest, setPendingRequest,
        currentTurnId, setCurrentTurnId,
        steerQueuedTurnId, setSteerQueuedTurnId,
        lastAgentEventRef,
        draftBufferRef,
        thoughtBufferRef,
        pendingRequestRef,
        currentTurnIdRef,
        steerQueuedTurnIdRef,
        thoughtExpandedRef,
        draftExpandedRef,
    } = useAgentState();

    const currentChatJid = useMemo(() => sessionId ? sessionToChatJid(sessionId) : '', [sessionId]);
    const renderedSelection = selection.capture();

    // ── Bootstrap ────────────────────────────────────────────────────────────

    useEffect(() => {
        const cleanupTheme = initTheme();
        const cleanupDisplayScale = installPwaDisplayScaleSync();
        // Enable meters by default until /meters slash command exists
        if (getLocalStorageItem('piclaw_system_meters_enabled') === null) {
            setLocalStorageItem('piclaw_system_meters_enabled', 'true');
        }
        Promise.all([
            ensureDefaultSession(),
            getRuntimeConfig(),
        ]).then(([sid, cfg]) => {
            selection.select(sid);
            setSessionId(sid);
            setRuntimeConfig(cfg);
            setUserProfile({ name: cfg.user_name, avatarUrl: cfg.user_avatar, avatarBackground: cfg.user_avatar_background });
            setAgents({
                [DEFAULT_AGENT_ID]: {
                    id: DEFAULT_AGENT_ID,
                    name: cfg.assistant_name || 'Gi',
                    avatar_url: cfg.assistant_avatar || null,
                },
            });
            setActiveModel(cfg.current || cfg.default_model || '');
            setActiveThinkingLevel(cfg.default_thinking_level || '');
            setSupportsThinking(Boolean(cfg.supports_thinking));
            setAgentModelsPayload(cfg);
            setReady(true);
        }).catch((err) => {
            console.error('[gi] Bootstrap failed:', err);
        });
        return () => { cleanupTheme?.(); cleanupDisplayScale(); };
    }, []);

    // ── Timeline loading ─────────────────────────────────────────────────────

    const loadPosts = useCallback(async (opts: any = {}) => {
        if (!sessionId) return;
        const scope = selection.capture();
        if (scope.sessionId !== sessionId) return;
        const chatJid = sessionToChatJid(sessionId);
        const data = await getTimeline(50, opts.beforeId || null, chatJid);
        if (!selection.isCurrent(scope)) return;
        const incoming: any[] = data.posts || [];
        if (opts.beforeId) {
            setPosts((prev: any[]) => dedupePosts([...incoming, ...prev]));
        } else {
            setPosts(dedupePosts(incoming));
        }
        setHasMore(incoming.length >= 50);
    }, [sessionId]);

    const scrollToBottom = useCallback(() => {
        const el = timelineRef.current;
        if (!el) return;
        el.scrollTop = el.scrollHeight;
    }, []);

    const refreshSessionLists = useCallback(async (sid: string | null) => {
        if (!sid) {
            setActiveChatAgents([]);
            setCurrentChatBranches([]);
            return;
        }
        const scope = selection.capture();
        if (scope.sessionId !== sid) return;
        const chatJid = sessionToChatJid(sid);
        const [agentsPayload, branchesPayload] = await Promise.all([
            getActiveChatAgents().catch(() => ({ agents: [] })),
            getChatBranches(chatJid).catch(() => ({ branches: [] })),
        ]);
        if (!selection.isCurrent(scope)) return;
        const agentsList = Array.isArray((agentsPayload as any)?.agents) ? (agentsPayload as any).agents : [];
        setAgents(Object.fromEntries(agentsList.map((entry: any) => [entry.agent_id, {
            id: entry.agent_id, name: entry.agent_name, avatar_url: null,
        }])));
        const branchesList = Array.isArray((branchesPayload as any)?.branches)
            ? (branchesPayload as any).branches
            : (Array.isArray((branchesPayload as any)?.chats) ? (branchesPayload as any).chats : []);
        setActiveChatAgents(agentsList.map((entry: any) => ({
            ...entry,
            is_active: entry?.chat_jid === chatJid,
        })));
        setCurrentChatBranches(branchesList);
    }, []);

    // ── SSE connection (replaces polling) ─────────────────────────────────────

    const handleSseEvent = useCallback((eventType: string, data: any) => {
        if (!selection.current() || data?.chat_jid !== sessionToChatJid(selection.current()!)) return;
        // Handle new_post events directly for immediate timeline updates
        if (eventType === 'new_post' || eventType === 'agent_response') {
            if (data && data.id) {
                setPosts((prev: any[]) => appendUniqueTimelinePost(prev, data));
                scrollToBottom();
            }
        }

        // Handle agent status events
        if (eventType === 'agent_status') {
            setAgentStatus(data);
            const active = data?.status === 'running' || data?.status === 'cancelling';
            setIsAgentTurnActive(active);
            isAgentRunningRef.current = active;
        }

        // Handle draft deltas for streaming display
        if (eventType === 'agent_draft_delta') {
            const delta = data?.delta || '';
            if (delta) {
                draftBufferRef.current = (draftBufferRef.current || '') + delta;
                setAgentDraft({ text: draftBufferRef.current, totalLines: 0, fullText: draftBufferRef.current });
            }
        }

        // Handle thought deltas
        if (eventType === 'agent_thought_delta') {
            const delta = data?.delta || '';
            if (delta) {
                thoughtBufferRef.current = (thoughtBufferRef.current || '') + delta;
                setAgentThought({ text: thoughtBufferRef.current, totalLines: 0, fullText: thoughtBufferRef.current });
            }
        }

        // Clear draft/thought on agent_response (turn complete)
        if (eventType === 'agent_response') {
            draftBufferRef.current = '';
            thoughtBufferRef.current = '';
            setAgentDraft(null);
            setAgentThought(null);
        }
    }, [scrollToBottom]);

    const handleConnectionStatusChange = useCallback((status: string) => {
        setConnectionStatus(status);
    }, []);

    useSseConnection({
        handleSseEvent,
        handleConnectionStatusChange,
        loadPosts,
        onWake: () => { loadPosts(); },
        chatJid: currentChatJid,
    });

    const refreshSelectedState = useCallback(async () => {
        const scope = selection.capture();
        if (!sessionId || scope.sessionId !== sessionId) return;
        const chat = sessionToChatJid(sessionId);
        try {
            const [models, queue, status] = await Promise.all([
                getAgentModels(chat), getAgentQueueState(chat), getAgentStatus('', chat),
            ]);
            if (!selection.isCurrent(scope)) return;
            setAgentModelsPayload(models);
            setActiveModel(models.current);
            setActiveThinkingLevel(models.thinking_level);
            setSupportsThinking(models.supports_thinking);
            setFollowupQueueItems(queue.items || []);
            setAgentStatus(status);
            const running = status?.status === 'running' || status?.status === 'cancelling';
            setIsAgentTurnActive(running);
            isAgentRunningRef.current = running;
            setSessionError(null);
        } catch (error) {
            if (selection.isCurrent(scope)) setSessionError(error.message || 'Unable to refresh session');
        }
    }, [sessionId]);

    // ── Initial load + light periodic refresh ─────────────────────────────────

    useEffect(() => {
        if (!ready || !sessionId) return;
        loadPosts();
        void refreshSessionLists(sessionId);
        void refreshSelectedState();
        // Light refresh every 10s as a safety net (SSE handles real-time)
        const id = setInterval(() => {
            loadPosts();
            void refreshSessionLists(sessionId);
            void refreshSelectedState();
        }, 10000);
        return () => clearInterval(id);
    }, [ready, sessionId, loadPosts, refreshSessionLists, refreshSelectedState]);

    // ── Send ──────────────────────────────────────────────────────────────────

    const handlePost = useCallback(async (response: any) => {
        // A pending send remains owned by its origin even after selection changes.
        if (!selection.isCurrent(renderedSelection)) return;
        await loadPosts();
        if (!selection.isCurrent(renderedSelection)) return;
        void refreshSelectedState();
        void refreshSessionLists(sessionId);
        scrollToBottom();
    }, [loadPosts, refreshSessionLists, refreshSelectedState, scrollToBottom, sessionId]);

    const handleSwitchChat = useCallback((chatJid: string | null) => {
        const nextSessionId = typeof chatJid === 'string' && chatJid.startsWith('gi:') ? chatJid.slice(3) : null;
        if (!nextSessionId || nextSessionId === sessionId) return;
        if (sessionId) Object.assign(getDraft(sessionId), { fileRefs, messageRefs });
        // Advance synchronously, before rendering, to invalidate already pending work.
        selection.select(nextSessionId);
        setLocalStorageItem(SESSION_KEY, nextSessionId);
        setSessionId(nextSessionId);
        setPosts([]); setHasMore(false); setFollowupQueueItems([]); setCurrentChatBranches([]);
        setFileRefs(getDraft(nextSessionId).fileRefs);
        setMessageRefs(getDraft(nextSessionId).messageRefs);
        setAgentStatus(null); setAgentDraft(null); setAgentThought(null); setAgentPlan(null);
        setPendingRequest(null); setCurrentTurnId(null); setSteerQueuedTurnId(null);
        draftBufferRef.current = ''; thoughtBufferRef.current = '';
        currentTurnIdRef.current = null; steerQueuedTurnIdRef.current = null;
        setIsAgentTurnActive(false); isAgentRunningRef.current = false;
        setActiveModel(''); setActiveThinkingLevel(''); setSupportsThinking(false);
        setAgentModelsPayload(null); setContextUsage(null); setModelUsage(null);
        setSessionError(null);
    }, [sessionId, fileRefs, messageRefs]);

    const handleCreateSession = useCallback(async () => {
        if (!sessionId) return;
        const scope = selection.capture();
        try {
            // A second main-session POST reuses the agent's existing main session.
            const created = await forkChatBranch(sessionToChatJid(sessionId));
            if (!selection.isCurrent(scope)) return;
            if (!created?.branch?.chat_jid) throw new Error('Missing created chat identifier');
            handleSwitchChat(created.branch.chat_jid);
        } catch (error) {
            if (selection.isCurrent(scope)) setSessionError(error.message || 'Failed to create session');
        }
    }, [sessionId, handleSwitchChat]);

    // ── Pane helpers ──────────────────────────────────────────────────────────

    const openEditor = useCallback((path: string) => {
        const existing = tabs.find((t: any) => t.id === path || t.path === path);
        if (existing) { setActiveTabId(existing.id); return; }
        setTabs((prev: any[]) => [...prev, { id: path, path, label: path.split('/').pop() || path, dirty: false, pinned: false }]);
        setActiveTabId(path);
    }, [tabs]);

    const handleTabClose = useCallback((id: string) => {
        setTabs((prev: any[]) => {
            const next = prev.filter((t: any) => t.id !== id);
            if (activeTabId === id) setActiveTabId(next[next.length - 1]?.id || null);
            return next;
        });
    }, [activeTabId]);

    // ── Shell class ───────────────────────────────────────────────────────────

    const appShellClass = [
        'app-shell',
        workspaceOpen ? '' : 'workspace-collapsed',
        editorOpen ? 'editor-open' : '',
    ].filter(Boolean).join(' ');

    // ── Render ────────────────────────────────────────────────────────────────

    if (!ready) {
        return html`<div id="app"><div style="padding:20px;text-align:center;color:var(--text-secondary,#888)">Loading…</div></div>`;
    }

    return html`
        <div class=${appShellClass}>
            <${SystemMetersHud} mode="overlay" />
            <${TimelineMenu}
                workspaceOpen=${workspaceOpen}
                toggleWorkspace=${() => setWorkspaceOpen((v: boolean) => !v)}
                chatOnlyMode=${false}
                openEditor=${openEditor}
            />
            <${WorkspaceExplorer}
                onFileSelect=${(path: string) => setFileRefs((p: string[]) => [...p, path])}
                visible=${workspaceOpen}
                active=${workspaceOpen || editorOpen}
                onOpenEditor=${openEditor}
                onOpenTerminalTab=${() => {}}
                onOpenVncTab=${() => {}}
            />
            ${workspaceOpen && html`<button
                class="workspace-drawer-backdrop"
                onClick=${() => setWorkspaceOpen(false)}
                aria-label="Hide workspace"
                title="Hide workspace"
            ></button>`}
            <button
                class=${`workspace-toggle-tab${workspaceOpen ? ' open' : ' closed'}`}
                onClick=${() => setWorkspaceOpen((v: boolean) => !v)}
                title=${workspaceOpen ? 'Hide workspace' : 'Show workspace'}
                aria-label=${workspaceOpen ? 'Hide workspace' : 'Show workspace'}
                aria-expanded=${workspaceOpen ? 'true' : 'false'}
            >
                <svg class="workspace-toggle-tab-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="6 3 11 8 6 13" />
                </svg>
            </button>
            <div class="workspace-splitter"></div>
            ${editorOpen && html`
                <div class="editor-pane-container">
                    <${TabStrip}
                        tabs=${tabs}
                        activeId=${activeTabId}
                        onActivate=${(id: string) => setActiveTabId(id)}
                        onClose=${handleTabClose}
                        onCloseOthers=${(id: string) => setTabs((p: any[]) => p.filter((t: any) => t.id === id))}
                        onCloseAll=${() => { setTabs([]); setActiveTabId(null); }}
                        onTogglePin=${() => {}}
                    />
                    <div class="editor-pane-host"></div>
                </div>
                <div class="editor-splitter"></div>
            `}
            <div class="container">
                <${Timeline}
                    posts=${posts}
                    hasMore=${hasMore}
                    onLoadMore=${({ preserveScroll }: any) => {
                        const oldest = posts[0];
                        if (oldest) loadPosts({ beforeId: oldest.id });
                    }}
                    timelineRef=${timelineRef}
                    onHashtagClick=${() => {}}
                    onMessageRef=${() => {}}
                    onScrollToMessage=${() => {}}
                    onFileRef=${openEditor}
                    onPostClick=${undefined}
                    onDeletePost=${() => {}}
                    onOpenWidget=${(w: any) => setFloatingWidget(w)}
                    onOpenAttachmentPreview=${setAttachmentPreview}
                    emptyMessage="Send a message to get started."
                    agents=${agents}
                    user=${userProfile}
                    reverse=${true}
                    removingPostIds=${new Set()}
                    searchQuery=""
                />
                <${AgentStatus}
                    status=${isCompactionStatus(agentStatus) ? null : agentStatus}
                    draft=${agentDraft}
                    plan=${agentPlan}
                    thought=${agentThought}
                    pendingRequest=${pendingRequest}
                    intent=${null}
                    turnId=${currentTurnId}
                    steerQueued=${Boolean(steerQueuedTurnId)}
                    onPanelToggle=${() => {}}
                    showExtensionPanels=${false}
                />
                <${FloatingWidgetPane}
                    widget=${floatingWidget}
                    onClose=${() => setFloatingWidget(null)}
                    onWidgetEvent=${() => {}}
                />
                ${attachmentPreview && html`
                    <${AttachmentPreviewModal}
                        mediaId=${attachmentPreview.mediaId}
                        info=${attachmentPreview.info}
                        onClose=${() => setAttachmentPreview(null)}
                    />
                `}
                <${QueuedFollowupStack}
                    items=${followupQueueItems}
                    onInjectQueuedFollowup=${() => {}}
                    onRemoveQueuedFollowup=${() => {}}
                    onMoveQueuedFollowup=${() => {}}
                    onOpenFilePill=${openEditor}
                />
                ${sessionError && html`<div role="alert">${sessionError}</div>`}
                <${ComposeBox}
                    key=${sessionId}
                    draftValue=${getDraft(sessionId).text}
                    draftMediaFiles=${getDraft(sessionId).media}
                    onContentChange=${(text: string) => { getDraft(sessionId).text = text; }}
                    onDraftMediaChange=${(media: File[]) => { getDraft(sessionId).media = media; }}
                    currentChatJid=${currentChatJid}
                    isAgentActive=${isAgentTurnActive}
                    onPost=${handlePost}
                    onFocus=${() => { if (!isIOSDevice()) scrollToBottom(); }}
                    onModelChange=${(value: string | null) => {
                        if (!selection.isCurrent(renderedSelection)) return;
                        setActiveModel(value || '');
                    }}
                    onModelStateChange=${(state: any) => {
                        if (!selection.isCurrent(renderedSelection)) return;
                        if (state && typeof state === 'object') {
                            setAgentModelsPayload((prev: any) => ({ ...(prev || {}), ...(state || {}) }));
                            if (typeof state.model === 'string') setActiveModel(state.model);
                            if (typeof state.thinking_level_label === 'string' && state.thinking_level_label.trim()) {
                                setActiveThinkingLevel(state.thinking_level_label);
                            } else if (typeof state.thinking_level === 'string' && state.thinking_level.trim()) {
                                setActiveThinkingLevel(state.thinking_level);
                            }
                            if (typeof state.supports_thinking === 'boolean') setSupportsThinking(state.supports_thinking);
                            if (state.provider_usage !== undefined) setModelUsage(state.provider_usage ?? null);
                        }
                    }}
                    agents=${agents}
                    currentSessionAgent=${activeChatAgents.find((entry: any) => entry?.chat_jid === currentChatJid) || null}
                    agentStatus=${agentStatus}
                    agentDraft=${agentDraft}
                    contextUsage=${contextUsage}
                    fileRefs=${fileRefs}
                    messageRefs=${messageRefs}
                    onRemoveFileRef=${(p: string) => setFileRefs((prev: string[]) => prev.filter(x => x !== p))}
                    onClearFileRefs=${() => {
                        getDraft(sessionId).fileRefs = [];
                        if (selection.current() === sessionId) setFileRefs([]);
                    }}
                    onSetFileRefs=${(refs: string[]) => {
                        getDraft(sessionId).fileRefs = refs;
                        if (selection.current() === sessionId) setFileRefs(refs);
                    }}
                    onRemoveMessageRef=${() => {}}
                    onClearMessageRefs=${() => {
                        getDraft(sessionId).messageRefs = [];
                        if (selection.current() === sessionId) setMessageRefs([]);
                    }}
                    onSetMessageRefs=${(refs: any[]) => {
                        getDraft(sessionId).messageRefs = refs;
                        if (selection.current() === sessionId) setMessageRefs(refs);
                    }}
                    connectionStatus=${connectionStatus}
                    activeChatAgents=${activeChatAgents}
                    currentChatBranches=${currentChatBranches}
                    onSwitchChat=${handleSwitchChat}
                    onCreateSession=${handleCreateSession}
                    formatBranchPickerLabel=${(b: any) => b?.label || b?.chat_jid || ''}
                    handleBranchPickerChange=${() => {}}
                    searchOpen=${false}
                    onEnterSearch=${() => {}}
                    onExitSearch=${() => {}}
                    onSearch=${() => {}}
                    searchScope="current"
                    onSearchScopeChange=${() => {}}
                    activeModel=${activeModel}
                    agentModelsPayload=${agentModelsPayload}
                    modelUsage=${modelUsage}
                    thinkingLevel=${activeThinkingLevel}
                    supportsThinking=${supportsThinking}
                    followupQueueCount=${followupQueueItems.length}
                    notificationsEnabled=${false}
                    notificationPermission="default"
                    onComposeSubmitError=${() => {}}
                    pendingRequestRef=${pendingRequestRef}
                    setPendingRequest=${setPendingRequest}
                />
            </div>
        </div>
    `;
}

render(html`<${GiApp} />`, document.getElementById('app'));
