// @ts-nocheck
/**
 * app.ts — Gi entry point.
 *
 * Uses Piclaw's web shell components verbatim. The only Gi-specific
 * concern is mapping Gi sessions onto Piclaw's chat_jid model so every
 * component receives what it expects without modification.
 */
import { html, render, useState, useEffect, useMemo, useCallback } from './vendor/preact-htm.js';
import { getLocalStorageBoolean, getLocalStorageItem, setLocalStorageItem } from './utils/storage.js';
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
import { formatBranchPickerLabel } from './ui/branch-lifecycle.js';
import { paneRegistry, tabStore } from './panes/index.js';
import {
    getTimeline,
    getThread,
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
} from './api.js';
import { Timeline } from './components/timeline.js';
import { ComposeBox } from './components/compose-box.js';
import { AgentStatus } from './components/status.js';
import { WorkspaceExplorer } from './components/workspace-explorer.js';
import { TabStrip } from './components/tab-strip.js';

// ── Constants ──────────────────────────────────────────────────────────────

const DEFAULT_AGENT_ID = 'gi';
const SESSION_STORAGE_KEY = 'gi_current_session_id';
const POLL_INTERVAL_MS = 1200;

// ── Gi-specific session bridge ─────────────────────────────────────────────

function sessionToChatJid(sessionId: string | null) {
    return sessionId ? `gi:${sessionId}` : 'web:default';
}

async function listSessions() {
    const res = await fetch('/api/sessions');
    if (!res.ok) return [];
    const data = await res.json();
    return data.sessions || [];
}

async function createSession(title: string) {
    const res = await fetch('/api/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title }),
    });
    if (!res.ok) throw new Error('Failed to create session');
    return res.json();
}

async function getRuntimeConfig() {
    const res = await fetch('/api/runtime/config');
    if (!res.ok) return {};
    return res.json();
}

// ── Session list sidebar ───────────────────────────────────────────────────

function SessionSidebar({ sessions, currentSessionId, onSelect, onCreate }) {
    const [title, setTitle] = useState('');
    return html`
        <div class="gi-session-sidebar">
            <div class="workspace-header">
                <span class="workspace-header-left">Sessions</span>
                <div class="workspace-header-actions">
                    <button
                        class="workspace-create"
                        title="New session"
                        onClick=${async () => {
                            const s = await onCreate(title || 'Session');
                            setTitle('');
                            onSelect(s.id);
                        }}
                    >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
                        </svg>
                    </button>
                </div>
            </div>
            <div class="workspace-tree">
                ${sessions.map((s: any) => html`
                    <button
                        key=${s.id}
                        class=${`workspace-row${currentSessionId === s.id ? ' workspace-row-active' : ''}`}
                        style="padding-left:12px"
                        onClick=${() => onSelect(s.id)}
                    >
                        <span class="workspace-row-icon">
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                            </svg>
                        </span>
                        <span class="workspace-row-label">${s.title || s.id}</span>
                    </button>
                `)}
            </div>
        </div>
    `;
}

// ── Main app ───────────────────────────────────────────────────────────────

function GiApp() {
    const [runtimeConfig, setRuntimeConfig] = useState<any>({});
    const [sessions, setSessions] = useState<any[]>([]);
    const [currentSessionId, setCurrentSessionId] = useState<string | null>(() =>
        getLocalStorageItem(SESSION_STORAGE_KEY) || null
    );

    const currentChatJid = useMemo(() => sessionToChatJid(currentSessionId), [currentSessionId]);

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
        lastSilenceNoticeRef,
        isAgentRunningRef,
        draftBufferRef,
        thoughtBufferRef,
        previewResyncPendingRef,
        previewResyncGenerationRef,
        pendingRequestRef,
        stalledPostIdRef,
        currentTurnIdRef,
        steerQueuedTurnIdRef,
        thoughtExpandedRef,
        draftExpandedRef,
    } = useAgentState();

    // Timeline posts
    const [posts, setPosts] = useState<any[]>([]);
    const [hasMorePosts, setHasMorePosts] = useState(false);

    // Workspace / pane state
    const [workspaceOpen, setWorkspaceOpen] = useState(false);
    const [tabs, setTabs] = useState<any[]>([]);
    const [activeTabId, setActiveTabId] = useState<string | null>(null);

    // Agents / models
    const [agents, setAgents] = useState<any>({});
    const [userProfile, setUserProfile] = useState<any>(null);

    // Agent running/silence
    const [isAgentTurnActive, setIsAgentTurnActive] = useState(false);

    // ── Initialise ──────────────────────────────────────────────────────────

    useEffect(() => {
        const cleanup = initTheme();
        getRuntimeConfig().then((cfg: any) => {
            setRuntimeConfig(cfg);
            setUserProfile({ name: cfg.user_name, avatar_url: cfg.user_avatar });
            setAgents({
                [DEFAULT_AGENT_ID]: {
                    id: DEFAULT_AGENT_ID,
                    name: cfg.assistant_name || 'Gi',
                    avatar_url: cfg.assistant_avatar || null,
                },
            });
        });
        listSessions().then(setSessions);
        return cleanup;
    }, []);

    // ── Session selection ───────────────────────────────────────────────────

    useEffect(() => {
        if (currentSessionId) {
            setLocalStorageItem(SESSION_STORAGE_KEY, currentSessionId);
        }
    }, [currentSessionId]);

    async function handleSelectSession(id: string) {
        setCurrentSessionId(id);
        setPosts([]);
        await loadTimeline(id);
    }

    async function handleCreateSession(title: string) {
        const s = await createSession(title);
        const updated = await listSessions();
        setSessions(updated);
        return s;
    }

    // ── Timeline loading ────────────────────────────────────────────────────

    async function loadTimeline(sessionId = currentSessionId, beforeId: number | null = null) {
        if (!sessionId) return;
        const chatJid = sessionToChatJid(sessionId);
        const data = await getTimeline(50, beforeId, chatJid);
        const incoming = data.posts || [];
        if (beforeId) {
            setPosts((prev: any[]) => dedupePosts([...incoming, ...prev]));
        } else {
            setPosts(dedupePosts(incoming));
        }
        setHasMorePosts(incoming.length >= 50);
    }

    // ── Polling ─────────────────────────────────────────────────────────────

    useEffect(() => {
        if (!currentSessionId) return;
        const id = setInterval(async () => {
            await loadTimeline(currentSessionId);
            const chatJid = sessionToChatJid(currentSessionId);
            const status = await getAgentStatus(DEFAULT_AGENT_ID, chatJid).catch(() => null);
            if (status) {
                setAgentStatus(status);
                const active = status.status === 'running' || status.status === 'cancelling';
                setIsAgentTurnActive(active);
                isAgentRunningRef.current = active;
            } else {
                setAgentStatus(null);
                setIsAgentTurnActive(false);
                isAgentRunningRef.current = false;
            }
        }, POLL_INTERVAL_MS);
        loadTimeline(currentSessionId);
        return () => clearInterval(id);
    }, [currentSessionId]);

    // ── Send message ────────────────────────────────────────────────────────

    async function handleSend(content: string, options: any = {}) {
        if (!currentSessionId) return;
        const chatJid = sessionToChatJid(currentSessionId);
        await sendAgentMessage(DEFAULT_AGENT_ID, content, null, [], options.mode || 'auto', chatJid);
    }

    async function handleAbort() {
        if (!currentSessionId) return;
        const chatJid = sessionToChatJid(currentSessionId);
        await sendAgentMessage(DEFAULT_AGENT_ID, '/abort', null, [], 'steer', chatJid).catch(() => null);
    }

    // ── Pane helpers ────────────────────────────────────────────────────────

    function openEditorTab(path: string) {
        const existing = tabs.find((t: any) => t.path === path);
        if (existing) { setActiveTabId(existing.id); return; }
        const id = `tab-${Date.now()}`;
        setTabs((prev: any[]) => [...prev, { id, path, label: path.split('/').pop() || path }]);
        setActiveTabId(id);
    }

    function closeTab(id: string) {
        setTabs((prev: any[]) => {
            const next = prev.filter((t: any) => t.id !== id);
            if (activeTabId === id) setActiveTabId(next[next.length - 1]?.id || null);
            return next;
        });
    }

    // ── Render ──────────────────────────────────────────────────────────────

    const hasSession = Boolean(currentSessionId);

    return html`
        <div class="app-shell ${workspaceOpen ? '' : 'workspace-collapsed'}">
            <${WorkspaceExplorer}
                currentChatJid=${currentChatJid}
                onOpenFile=${openEditorTab}
            />
            <div class="workspace-splitter"></div>
            <div class="container">
                <${SessionSidebar}
                    sessions=${sessions}
                    currentSessionId=${currentSessionId}
                    onSelect=${handleSelectSession}
                    onCreate=${handleCreateSession}
                />
                <div class="chat-window">
                    <div class="chat-window-header">
                        <div class="chat-window-header-main">
                            <div class="chat-window-header-title">
                                ${runtimeConfig.assistant_name || 'Gi'}
                            </div>
                            <div class="chat-window-header-subtitle">
                                ${currentSessionId
                                    ? sessions.find((s: any) => s.id === currentSessionId)?.title || currentSessionId
                                    : 'No session — create or select one'}
                            </div>
                        </div>
                        <div class="chat-window-header-actions">
                            <button
                                class="chat-window-header-button"
                                title="Toggle workspace"
                                onClick=${() => setWorkspaceOpen((v: boolean) => !v)}
                            >
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                                    <polyline points="9 22 9 12 15 12 15 22"/>
                                </svg>
                            </button>
                        </div>
                    </div>
                    ${tabs.length > 0 ? html`
                        <${TabStrip}
                            tabs=${tabs}
                            activeId=${activeTabId}
                            onActivate=${(id: string) => setActiveTabId(id)}
                            onClose=${closeTab}
                            onCloseOthers=${(id: string) => setTabs((prev: any[]) => prev.filter((t: any) => t.id === id))}
                            onCloseAll=${() => setTabs([])}
                            onTogglePin=${() => {}}
                        />
                    ` : null}
                    <${AgentStatus}
                        status=${agentStatus}
                        draft=${agentDraft}
                        plan=${agentPlan}
                        thought=${agentThought}
                        pendingRequest=${pendingRequest}
                        turnId=${currentTurnId}
                        steerQueued=${Boolean(steerQueuedTurnId)}
                    />
                    <${Timeline}
                        posts=${posts}
                        hasMore=${hasMorePosts}
                        onLoadMore=${async ({ preserveScroll }: any) => {
                            const oldest = posts[0];
                            if (oldest) await loadTimeline(currentSessionId, oldest.id);
                        }}
                        onPostClick=${() => {}}
                        onHashtagClick=${() => {}}
                        onMessageRef=${() => {}}
                        onScrollToMessage=${() => {}}
                        onFileRef=${openEditorTab}
                        onOpenWidget=${() => {}}
                        onOpenAttachmentPreview=${() => {}}
                        emptyMessage=${hasSession ? 'No messages yet.' : 'Create or select a session to start.'}
                        agents=${agents}
                        user=${userProfile}
                        onDeletePost=${() => {}}
                        reverse=${true}
                    />
                    <${ComposeBox}
                        currentChatJid=${currentChatJid}
                        isAgentActive=${isAgentTurnActive}
                        onSend=${(content: string, opts: any) => handleSend(content, opts)}
                        onAbort=${handleAbort}
                        agents=${agents}
                        currentSessionAgent=${agents[DEFAULT_AGENT_ID] ? {
                            ...agents[DEFAULT_AGENT_ID],
                            chat_jid: currentChatJid,
                        } : null}
                        agentStatus=${agentStatus}
                        agentDraft=${agentDraft}
                        contextUsage=${null}
                        disabled=${!hasSession}
                    />
                </div>
            </div>
        </div>
    `;
}

render(html`<${GiApp} />`, document.getElementById('app'));
