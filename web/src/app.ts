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
import { useAgentState } from './ui/use-agent-state.js';
import { initTheme } from './ui/theme.js';
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

// ── Gi session bridge ──────────────────────────────────────────────────────
// Piclaw components expect chat_jid strings. We map Gi sessions onto that
// model: the default session becomes 'gi:default'.

const DEFAULT_SESSION_TITLE = 'default';
const SESSION_KEY = 'gi_session_id';
const POLL_INTERVAL_MS = 1200;
const DEFAULT_AGENT_ID = 'gi';

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
    // Create a new default session
    const r = await fetch('/api/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: DEFAULT_SESSION_TITLE }),
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
    const [activeModel, setActiveModel] = useState<string>('');
    const [agentModelsPayload, setAgentModelsPayload] = useState<any>(null);
    const [activeThinkingLevel, setActiveThinkingLevel] = useState<string>('');
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

    // ── Bootstrap ────────────────────────────────────────────────────────────

    useEffect(() => {
        const cleanupTheme = initTheme();
        Promise.all([
            ensureDefaultSession(),
            getRuntimeConfig(),
        ]).then(([sid, cfg]) => {
            setSessionId(sid);
            setRuntimeConfig(cfg);
            setUserProfile({ name: cfg.user_name, avatarUrl: cfg.user_avatar });
            setAgents({
                [DEFAULT_AGENT_ID]: {
                    id: DEFAULT_AGENT_ID,
                    name: cfg.assistant_name || 'Gi',
                    avatar_url: cfg.assistant_avatar || null,
                },
            });
            setActiveModel(cfg.default_model || '');
            setActiveThinkingLevel(cfg.default_thinking_level || '');
            setReady(true);
        }).catch((err) => {
            console.error('[gi] Bootstrap failed:', err);
        });
        return cleanupTheme;
    }, []);

    // ── Timeline loading ─────────────────────────────────────────────────────

    const loadPosts = useCallback(async (opts: any = {}) => {
        if (!sessionId) return;
        const chatJid = sessionToChatJid(sessionId);
        const data = await getTimeline(50, opts.beforeId || null, chatJid);
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

    // ── Polling ───────────────────────────────────────────────────────────────

    useEffect(() => {
        if (!ready || !sessionId) return;
        loadPosts();
        const id = setInterval(async () => {
            await loadPosts();
            const chatJid = sessionToChatJid(sessionId);
            const status = await getAgentStatus(DEFAULT_AGENT_ID, chatJid).catch(() => null);
            if (status) {
                setAgentStatus(status);
                const active = status?.status === 'running' || status?.status === 'cancelling';
                setIsAgentTurnActive(active);
                isAgentRunningRef.current = active;
            } else {
                setAgentStatus(null);
                setIsAgentTurnActive(false);
                isAgentRunningRef.current = false;
            }
        }, POLL_INTERVAL_MS);
        return () => clearInterval(id);
    }, [ready, sessionId]);

    // ── Send ──────────────────────────────────────────────────────────────────

    const handlePost = useCallback(async (response: any) => {
        // Called by ComposeBox after a successful send
        await loadPosts();
        scrollToBottom();
    }, [loadPosts, scrollToBottom]);

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
            <${WorkspaceExplorer}
                onFileSelect=${(path: string) => setFileRefs((p: string[]) => [...p, path])}
                visible=${workspaceOpen}
                active=${workspaceOpen || editorOpen}
                onOpenEditor=${openEditor}
                onOpenTerminalTab=${() => {}}
                onOpenVncTab=${() => {}}
            />
            <button
                class=${`workspace-toggle-tab${workspaceOpen ? ' open' : ' closed'}`}
                onClick=${() => setWorkspaceOpen((v: boolean) => !v)}
                title=${workspaceOpen ? 'Hide workspace' : 'Show workspace'}
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
                <${ComposeBox}
                    currentChatJid=${currentChatJid}
                    isAgentActive=${isAgentTurnActive}
                    onPost=${handlePost}
                    onFocus=${() => { if (!isIOSDevice()) scrollToBottom(); }}
                    agents=${agents}
                    currentSessionAgent=${agents[DEFAULT_AGENT_ID] ? {
                        ...agents[DEFAULT_AGENT_ID],
                        chat_jid: currentChatJid,
                        agent_name: agents[DEFAULT_AGENT_ID].name,
                    } : null}
                    agentStatus=${agentStatus}
                    agentDraft=${agentDraft}
                    contextUsage=${contextUsage}
                    fileRefs=${fileRefs}
                    messageRefs=${messageRefs}
                    onRemoveFileRef=${(p: string) => setFileRefs((prev: string[]) => prev.filter(x => x !== p))}
                    onClearFileRefs=${() => setFileRefs([])}
                    onRemoveMessageRef=${() => {}}
                    onClearMessageRefs=${() => setMessageRefs([])}
                    connectionStatus=${connectionStatus}
                    activeChatAgents=${[]}
                    currentChatBranches=${[]}
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
                    activeModelUsage=${null}
                    activeThinkingLevel=${activeThinkingLevel}
                    supportsThinking=${false}
                    followupQueueCount=${followupQueueItems.length}
                    notificationsEnabled=${false}
                    notificationPermission="default"
                    onToggleNotifications=${() => {}}
                    onComposeSubmitError=${() => {}}
                    pendingRequestRef=${pendingRequestRef}
                    setPendingRequest=${setPendingRequest}
                />
            </div>
        </div>
    `;
}

render(html`<${GiApp} />`, document.getElementById('app'));
