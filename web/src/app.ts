// @ts-nocheck
/**
 * app.ts — Gi entry point.
 *
 * Uses Piclaw's web components verbatim. The shell structure mirrors
 * app-main-shell-render.ts exactly: app-shell > container > timeline +
 * status + compose. No session creation UI — a default session is
 * auto-created on startup, matching Piclaw's always-ready UX.
 */
import { html, render, useState, useEffect, useLayoutEffect, useMemo, useCallback, useRef } from './vendor/preact-htm.js';
import { getLocalStorageItem, setLocalStorageItem } from './utils/storage.js';
import { dedupePosts } from './ui/timeline-utils.js';
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
import { paneRegistry, tabStore, workspacePreviewPaneExtension, workspaceMarkdownPreviewPaneExtension } from './panes/index.js';

// Register only read-only previews; editor/specialised tab lifecycle is separate.
paneRegistry.register(workspacePreviewPaneExtension);
paneRegistry.register(workspaceMarkdownPreviewPaneExtension);
import {
    getTimeline,
    searchPosts,
    deletePost,
    getAgents,
    getAgentThought,
    setAgentThoughtVisibility,
    getAgentStatus,
    cancelSessionRun,
    getSessionCompaction, compactSession,
    getAgentContext,
    getAutoresearchStatus,
    stopAutoresearch,
    dismissAutoresearch,
    getAgentModels,
    completeInstanceOobe,
    getActiveChatAgents,
    getChatBranches,
    renameChatBranch,
    pinChatSession,
    pruneChatBranch,
    restoreChatBranch,
    getAgentQueueState,
    removeAgentQueueItem,
    reorderAgentQueueItem,
    steerAgentQueueItem,
    streamSidePrompt,
    getWorkspaceFile,
    sendAgentMessage,
    forkChatBranch,
    getThread,
} from './api.js';
import { Timeline } from './components/timeline.js';
import { ComposeBox, QueuedFollowupStack, parseQueuedContent } from './components/compose-box.js';
import { AgentStatus, AgentRequestModal } from './components/status.js';
import { WorkspaceExplorer } from './components/workspace-explorer.js';
import { TabStrip } from './components/tab-strip.js';
import { FloatingWidgetPane } from './components/floating-widget-pane.js';
import { AttachmentPreviewModal } from './components/attachment-preview-modal.js';
import { SystemMetersHud } from './components/system-meters-hud.js';
import { TimelineMenu } from './components/timeline-menu.js';
import { createSelectionScope } from './gi-session-state.js';
import { createDraftRepository, indexedDraftStorage, emptyDraft } from './gi-drafts.js';
import { recoverQueueDraft } from './gi-queue-return.js';

// ── Gi session bridge ──────────────────────────────────────────────────────
// Piclaw components expect chat_jid strings. We map Gi sessions onto that
// model: the default session becomes 'gi:default'.

import { createActivityRevision, compactionNotice, compactionElapsed } from './gi-compaction-state.js';
import { contextPresentation } from './gi-context-usage.js';
import {createActivationRefreshGate,createTimelineRevision,createAssetVersionGuard,loadedAssetVersion} from './gi-refresh-guards.js';
import {createSearchView} from './gi-search-state.js';
import {newMessageWindow,mergeMessagePages,captureTimelineAnchor,restoreTimelineAnchor} from './gi-message-pages.js';

import {bindWorkspaceVisibility} from './gi-workspace-visibility.js';

const DEFAULT_SESSION_TITLE = 'default';
const SESSION_KEY = 'gi_session_id';
const POLL_INTERVAL_MS = 1200;
const DEFAULT_AGENT_ID = 'web';

// The pinned stack has no disabled-Steer prop. Apply native button state at
// the host boundary, without changing the supplied component or appearance.
function RunBoundQueueStack({ steerEnabled, ...props }: any) {
    const root = useRef(null);
    useLayoutEffect(() => {
        const pending = new Set(props.items.filter(item => item.pending).map(item => String(item.id)));
        root.current?.querySelectorAll('.compose-queue-stack-steer-btn').forEach(button => {
            button.disabled = !steerEnabled || props.busy || pending.has(button.closest('[data-queue-id]')?.dataset.queueId);
        });
    });
    return html`<div ref=${root} style="display:contents"><${QueuedFollowupStack} ...${props} /></div>`;
}

// Folder clicks remain navigation. This host-owned action uses the existing
// workspace header without editing the supplied explorer or composer.
function useWorkspaceFolderReference(visible:boolean, sessionId:string, fileRefs:string[], attach:(path:string)=>void) {
    const latest=useRef({fileRefs,attach});latest.current={fileRefs,attach};
    const syncRef=useRef<(()=>void)|null>(null);
    useLayoutEffect(()=>{
        if(!visible)return;
        const sidebar=document.querySelector('.workspace-sidebar');
        const actions=sidebar?.querySelector('.workspace-header-actions');
        if(!sidebar||!actions)return;
        const button=document.createElement('button');
        button.type='button';button.className='menu-action-btn';button.textContent='+ folder';
        button.setAttribute('aria-label','Reference selected folder');
        const selected=()=>sidebar.querySelector<HTMLElement>('.workspace-row.selected[data-type="dir"]')?.dataset.path||'';
        const sync=()=>{
            const path=selected();
            button.hidden=!path;
            button.disabled=!path||latest.current.fileRefs.includes(path);
            button.title=path?`Reference folder: ${path}`:'Reference selected folder';
        };
        const click=()=>{const path=selected();if(path&&!latest.current.fileRefs.includes(path))latest.current.attach(path);};
        button.addEventListener('click',click);actions.prepend(button);sync();syncRef.current=sync;
        // Ignore the button's own attributes to avoid observer feedback.
        const observer=new MutationObserver(records=>{if(records.some(record=>!button.contains(record.target as Node)))sync();});
        observer.observe(sidebar,{subtree:true,childList:true,attributes:true,attributeFilter:['class','data-path','data-type']});
        return()=>{observer.disconnect();button.removeEventListener('click',click);button.remove();syncRef.current=null;};
    },[visible,sessionId]);
    useLayoutEffect(()=>{syncRef.current?.();},[fileRefs]);
}

// Keep the supplied component untouched. Its native title also supplies the
// tooltip-data contract; observe child-owned updates (e.g. model selection).
function useContextTooltip(root: any, usage: any, notice: any, now: number, canStop: boolean, stop: any, compact: any) {
    useLayoutEffect(() => {
        const compose = root.current?.querySelector('.compose-box');
        if (!compose) return;
        const sync = () => {
            compose.querySelectorAll('.send-btn.abort-mode').forEach(button => { if (button.disabled === canStop) button.disabled = !canStop; });
            compose.querySelectorAll('.compose-context-pie').forEach(button => {
                const active = notice?.intent_key === 'compaction';
                const normal = contextPresentation(usage, typeof compact === 'function');
                const canCompact = typeof compact === 'function' && !active;
                if (button.disabled === canCompact) button.disabled = !canCompact;
                const title = active ? `${notice.title} — ${compactionElapsed(notice, now)}` : normal.title + (canCompact ? '' : ' — Context usage');
                const label = active ? `${notice.title} — ${normal.label}` : normal.label;
                if (button.getAttribute('title') !== title) button.setAttribute('title', title);
                if (button.getAttribute('aria-label') !== label) button.setAttribute('aria-label', label);
                if (button.getAttribute('data-tooltip') !== title) button.setAttribute('data-tooltip', title);
                if (button.classList.contains('is-compacting') !== active) button.classList.toggle('is-compacting', active);
                let elapsed = compose.querySelector('.gi-compaction-elapsed');
                if (active) {
                    if (!elapsed) { elapsed = document.createElement('span'); elapsed.className = 'gi-compaction-elapsed'; button.after(elapsed); }
                    const text = compactionElapsed(notice, now);
                    if (elapsed.textContent !== text) elapsed.textContent = text;
                } else elapsed?.remove();
            });
        };
        sync();
        const observer = new MutationObserver(sync);
        observer.observe(compose, {subtree: true, childList: true, attributes: true, attributeFilter: ['title', 'disabled', 'class']});
        // Capture before the component's /abort handler clears its draft.
        const onClick = (event: any) => {
            if (event.target.closest?.('.compose-context-pie')) {
                event.preventDefault(); event.stopImmediatePropagation(); compact?.(); return;
            }
            if (event.target.closest?.('.send-btn.abort-mode')) {
                event.preventDefault(); event.stopImmediatePropagation(); stop();
            }
        };
        compose.addEventListener('click', onClick, true);
        return () => { observer.disconnect(); compose.removeEventListener('click', onClick, true); };
    });
}

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
    const containerRef = useRef(null);
    const [ready, setReady] = useState(false);
    const [sessionId, setSessionId] = useState<string | null>(null);
    const selection = useRef(createSelectionScope()).current;
    const [sessionError, setSessionError] = useState<string | null>(null);
    const [draftStorageError, setDraftStorageError] = useState('');
    const [draftRestore, setDraftRestore] = useState<any>(null);
    const draftsRef = useRef<any>(null);
    if (!draftsRef.current) draftsRef.current = createDraftRepository(indexedDraftStorage(), error => setDraftStorageError(`Draft not saved: ${error.message}`));
    const drafts = draftsRef.current;
    const getDraft = (sid: string) => drafts.get(sid);
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
    const messageWindow=useRef(newMessageWindow());
    const pageRequest=useRef<any>(null);
    const pageRefreshPending=useRef(false);
    const scrollRestore=useRef<any>(null);
    const readingAnchor=useRef<any>(null);
    const searchView=useRef(createSearchView()).current;
    const [searchState,setSearchState]=useState(searchView.capture());
    const [searchError,setSearchError]=useState('');
    const timelineRevision=useRef(createTimelineRevision()).current;
    const versionGuard=useRef(createAssetVersionGuard(loadedAssetVersion(document))).current;
    const [newUIVersion,setNewUIVersion]=useState('');
    const timelineRef = useRef<any>(null);

    // Compose
    const [fileRefs, setFileRefs] = useState<string[]>([]);
    const [messageRefs, setMessageRefs] = useState<any[]>([]);
    const [followupQueueItems, setFollowupQueueItems] = useState<any[]>([]);
    const [queueError, setQueueError] = useState('');
    const [queueBusy, setQueueBusy] = useState(false);
    const queueMutation = useRef<any>(null);
    const queueRevision = useRef(0);
    const [queueActiveTurnId, setQueueActiveTurnId] = useState(null);
    const modelRevision = useRef(0);
    const modelMutation = useRef<any>(null);
    const connectionRevision = useRef(0);
    const streamDisconnected = useRef(true);
    const activationRefresh=useRef(createActivationRefreshGate()).current;
    const refreshAfterConnection = useRef<() => void>(() => {});
    const refreshTimer = useRef<any>(null);
    const [optimisticQueue, setOptimisticQueue] = useState<any[]>([]);
    const [floatingWidget, setFloatingWidget] = useState<any>(null);
    const [attachmentPreview, setAttachmentPreview] = useState<any>(null);
    const [contextUsage, setContextUsage] = useState<any>(null);
    const [activity, setActivity] = useState<any>(null);
    const activityRevision = useRef(createActivityRevision()).current;
    const [activityNow, setActivityNow] = useState(Date.now());
    const [stopPending, setStopPending] = useState(false);
    const [stopError, setStopError] = useState('');
    const [activityFresh, setActivityFresh] = useState(false);
    const stopToken = useRef(null);
    const [compactState, setCompactState] = useState<any>(null);
    const [compactPending, setCompactPending] = useState(false);
    const [compactError, setCompactError] = useState('');
    const compactToken = useRef(null);
    const manualCompact = activityFresh && activity?.status === 'idle' && compactState?.available && !compactPending ? async () => {
        if (compactToken.current || streamDisconnected.current) return;
        const scope = selection.capture(); const expected = compactState.token; const token = {};
        compactToken.current=token; setCompactPending(true); setCompactError('');
        try { await compactSession(sessionToChatJid(scope.sessionId),expected); }
        catch (error) { if (selection.isCurrent(scope)) setCompactError(`Compact failed: ${error.message}`); }
        finally {
            if (compactToken.current===token) { compactToken.current=null; setCompactPending(false); }
            if (selection.isCurrent(scope)) {activityRevision.invalidate();setActivityFresh(false);refreshAfterConnection.current();}
        }
    } : null;
    const notice = compactionNotice(activity, activityNow);
    useEffect(() => {
        if (!activity?.compaction) return;
        const timer = setInterval(() => setActivityNow(Date.now()), 1000);
        return () => clearInterval(timer);
    }, [activity]);
    useContextTooltip(containerRef, contextUsage, notice, activityNow, activityFresh && !stopPending && !!activity?.turn_id && ['running','cancelling'].includes(activity?.status), async () => {
        if (stopToken.current || !activityFresh || !activity?.turn_id || streamDisconnected.current) return;
        const scope = selection.capture(); const run = activity.turn_id; const token = {};
        stopToken.current = token; setStopPending(true); setStopError('');
        try { await cancelSessionRun(sessionToChatJid(scope.sessionId), run); }
        catch (error) { if (selection.isCurrent(scope) && error.status !== 409) setStopError(`Stop failed: ${error.message}`); }
        finally {
            if (stopToken.current === token) { stopToken.current = null; setStopPending(false); }
            if (selection.isCurrent(scope)) { activityRevision.invalidate(); setActivityFresh(false); refreshAfterConnection.current(); }
        }
    }, manualCompact);
    const [activeChatAgents, setActiveChatAgents] = useState<any[]>([]);
    const sessionListRevision = useRef(0);
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
            drafts.load().catch(error => setDraftStorageError(`Draft recovery unavailable: ${error.message}`)),
        ]).then(([sid, cfg]) => {
            selection.select(sid);
            setSessionId(sid);
            setFileRefs(getDraft(sid).fileRefs);
            setMessageRefs(getDraft(sid).messageRefs);
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

    useLayoutEffect(() => {
        const pending=scrollRestore.current;scrollRestore.current=null;
        if(!pending||!selection.isCurrent(pending.scope)||!searchView.isCurrent(pending.view)||pending.connection!==connectionRevision.current)return;
        if(pending.bottom&&timelineRef.current){timelineRef.current.scrollTop=0;readingAnchor.current=null;}
        else {restoreTimelineAnchor(pending.anchor);readingAnchor.current=pending.anchor;}
    },[posts]);

    const loadPosts = useCallback(async (opts: any = {}) => {
        if (!sessionId || !activationRefresh.ready(selection.capture().generation)) return;
        const scope=selection.capture(),view=searchView.capture(),connection=connectionRevision.current;
        if(scope.sessionId!==sessionId||view.active)return;
        const older=opts.older===true;
        if(pageRequest.current?.connection===connection&&pageRequest.current?.generation===scope.generation){
            if(!older)pageRefreshPending.current=true;
            return pageRequest.current.promise;
        }
        if(older&&(!messageWindow.current.loaded||!messageWindow.current.hasMore))return;
        const token:any={connection,generation:scope.generation};pageRequest.current=token;
        const request=timelineRevision.begin();
        const valid=()=>selection.isCurrent(scope)&&searchView.isCurrent(view)&&connection===connectionRevision.current&&timelineRevision.accepts(request)&&!streamDisconnected.current;
        token.promise=(async()=>{
            try {
                const initial=!messageWindow.current.loaded||!messageWindow.current.after;
                let cursor=older?messageWindow.current.before:messageWindow.current.after;
                do {
                    const data=await getTimeline(50,older?cursor:null,sessionToChatJid(scope.sessionId),!older&&!initial?cursor:null);
                    if(!valid())return;
                    const root=timelineRef.current;
                    scrollRestore.current={scope,view,connection,anchor:captureTimelineAnchor(root,readingAnchor.current),bottom:!older&&(initial||!root||Math.abs(root.scrollTop)<80)};
                    const incoming=data.posts||[];
                    setPosts(prev=>mergeMessagePages(prev,incoming));
                    if(initial){messageWindow.current={loaded:true,before:data.before,after:data.after,hasMore:data.hasMore};}
                    else if(older){messageWindow.current.before=data.before||cursor;messageWindow.current.hasMore=data.hasMore;}
                    else {messageWindow.current.after=data.after||cursor;}
                    setHasMore(messageWindow.current.hasMore);
                    // Reconnect catch-up can span many bounded pages; never jump
                    // straight to newest and silently lose the intervening rows.
                    if(older||initial||!data.hasMore||!data.after||data.after===cursor)break;
                    cursor=data.after;
                } while(valid());
            } catch(error){if(valid())setSessionError(`Timeline refresh failed: ${error.message}`);}
            finally {
                if(pageRequest.current===token){pageRequest.current=null;if(pageRefreshPending.current){pageRefreshPending.current=false;refreshAfterConnection.current();}}
            }
        })();
        return token.promise;
    },[sessionId]);

    // The supplied reverse timeline's prefetch math assumes positive scrolling.
    // Own native negative-scroll paging at the existing host without editing it.
    useEffect(() => {
        const root=timelineRef.current;if(!root||searchState.active)return;
        const onScroll=()=>{const distance=root.scrollHeight-root.clientHeight+root.scrollTop;
            if(messageWindow.current.hasMore&&distance<200)void loadPosts({older:true});};
        const userScroll=()=>{readingAnchor.current=null;};
        root.addEventListener('scroll',onScroll,{passive:true});
        root.addEventListener('wheel',userScroll,{passive:true});root.addEventListener('touchstart',userScroll,{passive:true});root.addEventListener('pointerdown',userScroll);root.addEventListener('keydown',userScroll);window.addEventListener('resize',userScroll);
        return ()=>{root.removeEventListener('scroll',onScroll);root.removeEventListener('wheel',userScroll);root.removeEventListener('touchstart',userScroll);root.removeEventListener('pointerdown',userScroll);root.removeEventListener('keydown',userScroll);window.removeEventListener('resize',userScroll);};
    },[posts,searchState.active,loadPosts]);

    const runSearch = async (query?:string,scopeValue?:string) => {
        if(query!==undefined)setSearchState(searchView.query(query));
        if(scopeValue!==undefined)setSearchState(searchView.scope(scopeValue));
        const view=searchView.capture(),owner=selection.capture();
        if(!view.active||!owner.sessionId||!activationRefresh.ready(owner.generation))return;
        const request=timelineRevision.begin(),connection=connectionRevision.current;
        setSearchError('');
        if(!view.query){setPosts([]);setHasMore(false);return;}
        try{
            const result=await searchPosts(view.query,50,0,sessionToChatJid(owner.sessionId),view.scope);
            if(!selection.isCurrent(owner)||!searchView.isCurrent(view)||!timelineRevision.accepts(request)||connection!==connectionRevision.current||streamDisconnected.current)return;
            setPosts(dedupePosts(result.posts||[]));setHasMore(false);
        }catch(error){if(selection.isCurrent(owner)&&searchView.isCurrent(view)&&timelineRevision.accepts(request)&&connection===connectionRevision.current&&!streamDisconnected.current)setSearchError(`Search failed: ${error.message}`);}
    };
    const enterSearch = () => {timelineRevision.invalidate();readingAnchor.current=null;pageRequest.current=null;pageRefreshPending.current=false;scrollRestore.current=null;setSearchState(searchView.enter());setPosts([]);setHasMore(false);setSearchError('');};
    const exitSearch = () => {timelineRevision.invalidate();setPosts([]);messageWindow.current=newMessageWindow();pageRequest.current=null;pageRefreshPending.current=false;setSearchState(searchView.close());setSearchError('');void loadPosts();};

    const scrollToBottom = useCallback(() => {
        const el = timelineRef.current;
        if (!el) return;
        if(Math.abs(el.scrollTop)<80)el.scrollTop=0;
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
        const revision = ++sessionListRevision.current;
        const [agentsPayload, branchesPayload] = await Promise.all([
            getActiveChatAgents().catch(() => ({ agents: [] })),
            getChatBranches(chatJid).catch(() => ({ branches: [] })),
        ]);
        if (!selection.isCurrent(scope) || revision !== sessionListRevision.current) return;
        const agentsList = Array.isArray((agentsPayload as any)?.agents) ? (agentsPayload as any).agents : [];
        setAgents(Object.fromEntries(agentsList.map((entry: any) => [entry.agent_id, {
            id: entry.agent_id, name: entry.agent_name, avatar_url: null,
        }])));
        const branchesList = Array.isArray((branchesPayload as any)?.branches)
            ? (branchesPayload as any).branches
            : (Array.isArray((branchesPayload as any)?.chats) ? (branchesPayload as any).chats : []);
        setActiveChatAgents(agentsList);
        setCurrentChatBranches(branchesList);
    }, []);

    // ── SSE connection (replaces polling) ─────────────────────────────────────

    const handleSseEvent = useCallback((eventType: string, data: any) => {
        if (!selection.current() || data?.chat_jid !== sessionToChatJid(selection.current()!)) return;
        if(eventType==='connected'&&versionGuard.observe(data?.app_asset_version))setNewUIVersion(data.app_asset_version);
        if (eventType === 'agent_status' || eventType.startsWith('compaction_') || ['queue_changed', 'agent_response'].includes(eventType)) { activityRevision.invalidate(); setActivityFresh(false); }
        if (eventType.startsWith('compaction_') || ['new_post', 'agent_status', 'agent_response', 'queue_changed', 'agent_followup_queued', 'agent_followup_consumed', 'agent_followup_removed'].includes(eventType)) {
            ++queueRevision.current;
            if (!refreshTimer.current) refreshTimer.current = setTimeout(() => {
                refreshTimer.current = null;
                refreshAfterConnection.current();
            }, 0);
        }
        // Handle new_post events directly for immediate timeline updates
        if (eventType === 'new_post' || eventType === 'agent_response') {
            if (data?.id && data?.data && !searchView.capture().active) {
                const root=timelineRef.current;
                scrollRestore.current={scope:selection.capture(),view:searchView.capture(),connection:connectionRevision.current,anchor:captureTimelineAnchor(root,readingAnchor.current),bottom:!root||Math.abs(root.scrollTop)<80};
                setPosts((prev: any[]) => mergeMessagePages(prev,[data]));
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
        const shouldRefresh=activationRefresh.status(selection.capture().generation,status);
        ++connectionRevision.current;
        timelineRevision.invalidate();
        pageRequest.current=null;pageRefreshPending.current=false;scrollRestore.current=null;
        activityRevision.invalidate(); setActivity(null); setActivityFresh(false);
        ++queueRevision.current;
        setQueueActiveTurnId(null);
        setConnectionStatus(status);
        streamDisconnected.current = status !== 'connected';
        if (status !== 'connected') {
            setAgentStatus(null); setAgentDraft(null); setAgentPlan(null); setAgentThought(null);
            setPendingRequest(null); setCurrentTurnId(null); setSteerQueuedTurnId(null);
            draftBufferRef.current = ''; thoughtBufferRef.current = '';
            pendingRequestRef.current = null; currentTurnIdRef.current = null; steerQueuedTurnIdRef.current = null;
            setIsAgentTurnActive(false); isAgentRunningRef.current = false;
        } else if(shouldRefresh) refreshAfterConnection.current();
    }, []);

    useSseConnection({
        handleSseEvent: (type:string,data:any) => { if(selection.isCurrent(renderedSelection))handleSseEvent(type,data); },
        handleConnectionStatusChange: (status:string) => { if(selection.isCurrent(renderedSelection))handleConnectionStatusChange(status); },
        loadPosts,
        onWake: () => { if(selection.isCurrent(renderedSelection))refreshAfterConnection.current(); },
        chatJid: currentChatJid,
        selectionKey: renderedSelection.generation,
    });

    const refreshSelectedState = useCallback(async () => {
        const scope = selection.capture();
        if (!sessionId || scope.sessionId !== sessionId || !activationRefresh.ready(scope.generation)) return;
        const chat = sessionToChatJid(sessionId);
        const revision = ++queueRevision.current;
        const connection = connectionRevision.current;
        const modelVersion = modelRevision.current;
        const activityVersion = activityRevision.capture();
        try {
            const [models, queue, status, compact] = await Promise.all([
                getAgentModels(chat), getAgentQueueState(chat), getAgentStatus('', chat), getSessionCompaction(chat),
            ]);
            if (!selection.isCurrent(scope) || connection !== connectionRevision.current || streamDisconnected.current) return;
            if (!activityRevision.accepts(activityVersion)) return;
            setActivity(status); setCompactState(compact); setActivityFresh(true); setActivityNow(Date.now());
            if (modelVersion === modelRevision.current && !modelMutation.current) {
                setAgentModelsPayload(models);
                setActiveModel(models.current);
                setActiveThinkingLevel(models.thinking_level);
                setSupportsThinking(models.supports_thinking);
                setContextUsage(models.context_usage || null);
            }
            if (revision === queueRevision.current && !queueMutation.current) {
                setQueueActiveTurnId(queue.activeTurnId || null);
                setFollowupQueueItems(queue.items || []);
                const admitted = new Set((queue.items || []).map(item => item.metadata?.client_request_id).filter(Boolean));
                setOptimisticQueue(items => items.filter(item => !admitted.has(item.id)));
            }
            setAgentStatus(status);
            const running = status?.status === 'running' || status?.status === 'cancelling';
            setIsAgentTurnActive(running);
            isAgentRunningRef.current = running;
            setSessionError(null);
        } catch (error) {
            if (selection.isCurrent(scope)&&connection===connectionRevision.current&&activityRevision.accepts(activityVersion)&&!streamDisconnected.current) setSessionError(error.message || 'Unable to refresh session');
        }
    }, [sessionId]);

    refreshAfterConnection.current = () => {
        if(!activationRefresh.ready(selection.capture().generation))return;
        if(searchView.capture().active)void runSearch();else void loadPosts();
        void refreshSelectedState();
    };
    useEffect(() => () => { if (refreshTimer.current) clearTimeout(refreshTimer.current); }, []);

    // ── Initial load + light periodic refresh ─────────────────────────────────

    useEffect(() => {
        if (!ready || !sessionId) return;
        if(activationRefresh.activate(selection.capture().generation))refreshAfterConnection.current();
        void refreshSessionLists(sessionId);
        // Light refresh every 10s as a safety net (SSE handles real-time)
        const id = setInterval(() => {
            refreshAfterConnection.current();
            void refreshSessionLists(sessionId);
        }, 10000);
        return () => clearInterval(id);
    }, [ready, sessionId, loadPosts, refreshSessionLists, refreshSelectedState]);

    // ── Send ──────────────────────────────────────────────────────────────────

    const handlePost = useCallback((_response: any) => {
        // Acknowledgement belongs to the captured session, but refresh belongs
        // to the current view. Paging/SSE own near-bottom and reading anchors.
        if (!selection.isCurrent(renderedSelection)) return;
        refreshAfterConnection.current();
        void refreshSessionLists(sessionId);
    }, [refreshSessionLists, sessionId]);

    const handleSwitchChat = useCallback((chatJid: string | null) => {
        const nextSessionId = typeof chatJid === 'string' && chatJid.startsWith('gi:') ? chatJid.slice(3) : null;
        if (!nextSessionId || nextSessionId === sessionId) return;
        if (sessionId) drafts.update(sessionId, { fileRefs, messageRefs });
        // Advance synchronously, before rendering, to invalidate already pending work.
        selection.select(nextSessionId);
        activationRefresh.select(selection.capture().generation);streamDisconnected.current=true;setConnectionStatus('disconnected');
        setSearchState(searchView.close());setSearchError('');
        messageWindow.current=newMessageWindow();readingAnchor.current=null;pageRequest.current=null;pageRefreshPending.current=false;scrollRestore.current=null;
        timelineRevision.invalidate();
        stopToken.current = null; setStopPending(false); setStopError('');
        compactToken.current=null; setCompactPending(false); setCompactError(''); setCompactState(null);
        activityRevision.invalidate(); setActivity(null); setActivityFresh(false);
        setLocalStorageItem(SESSION_KEY, nextSessionId);
        setSessionId(nextSessionId);
        setPosts([]); setHasMore(false); setFollowupQueueItems([]); setQueueActiveTurnId(null); setCurrentChatBranches([]);
        queueMutation.current = null; ++queueRevision.current; setQueueBusy(false); setQueueError('');
        setOptimisticQueue([]);
        ++modelRevision.current; modelMutation.current = null;
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

    const handleSessionMutation = async (chatJid: string, action: string, value?: any) => {
        // Row-owned mutations: never change selection/drafts optimistically.
        ++sessionListRevision.current;
        if (action === 'rename') await renameChatBranch(chatJid, { title: value });
        else if (action === 'pin') await pinChatSession(chatJid, value);
        else if (action === 'archive') await pruneChatBranch(chatJid);
        else if (action === 'restore') await restoreChatBranch(chatJid);
        else throw new Error('Unsupported session action');
        const revision = ++sessionListRevision.current;
        const data = await getActiveChatAgents();
        if (revision === sessionListRevision.current) setActiveChatAgents(data.agents || []);
    };

    const mutateQueue = async (action: 'remove' | 'move' | 'return' | 'steer', itemOrIndex: any, toIndex?: number) => {
        if (queueMutation.current) return;
        if (action === 'steer' && (streamDisconnected.current || !isAgentTurnActive || !queueActiveTurnId || itemOrIndex.pending)) return;
        const expectedActiveTurnId = queueActiveTurnId;
        const scope = selection.capture();
        if (!scope.sessionId) return;
        const token = {}; queueMutation.current = token; ++queueRevision.current;
        setQueueBusy(true); setQueueError('');
        const before = [...followupQueueItems];
        const chat = sessionToChatJid(scope.sessionId);
        try {
            if (action === 'steer') {
                if (itemOrIndex.chat_jid !== chat) throw new Error('Queued item belongs to another session');
                await steerAgentQueueItem(itemOrIndex.id, chat, expectedActiveTurnId);
                if (selection.isCurrent(scope)) setFollowupQueueItems(items => items.filter(item => item.id !== itemOrIndex.id));
            } else if (action === 'return') {
                const item = itemOrIndex;
                if (item.chat_jid !== chat || item.pending) throw new Error('Queued item belongs to another session or has no durable ID');
                const recovered = drafts.hasQueueReturn(scope.sessionId, item.id)
                    ? emptyDraft() : await recoverQueueDraft(item, parseQueuedContent(item.content));
                const prepared = drafts.prepareQueueReturn(scope.sessionId, item.id, recovered);
                // Publish the merge immediately; typing while persistence is in
                // flight then writes on top of it instead of replacing it.
                if (selection.current() === scope.sessionId) {
                    setFileRefs(prepared.draft.fileRefs); setMessageRefs(prepared.draft.messageRefs);
                    setDraftRestore({sessionId: scope.sessionId, ...prepared.draft, token: crypto.randomUUID()});
                }
                await prepared.ready;
                await drafts.flushStable();
                await removeAgentQueueItem(item.id, chat);
                await drafts.completeQueueReturn(scope.sessionId, item.id);
            } else if (action === 'remove') {
                if (itemOrIndex.chat_jid !== chat) throw new Error('Queued item belongs to another session');
                setFollowupQueueItems(before.filter(item => item.id !== itemOrIndex.id));
                await removeAgentQueueItem(itemOrIndex.id, chat);
            } else {
                const after = [...before];
                if (!after[itemOrIndex] || toIndex! < 0 || toIndex! >= after.length) throw new Error('Invalid queue position');
                const [moved] = after.splice(itemOrIndex, 1); after.splice(toIndex!, 0, moved);
                setFollowupQueueItems(after);
                await reorderAgentQueueItem({chatJid: chat, expected: before.map(item => item.id), order: after.map(item => item.id)});
            }
        } catch (error) {
            if (action === 'return') drafts.queueReturnFailed(scope.sessionId, itemOrIndex.id, error.message);
            if (selection.isCurrent(scope)) { setFollowupQueueItems(before); setQueueError(`Queue action failed: ${error.message}`); }
        } finally {
            try {
                const fresh = await getAgentQueueState(chat);
                if (selection.isCurrent(scope)) {
                    setFollowupQueueItems(fresh.items || []);
                    if (!streamDisconnected.current) setQueueActiveTurnId(fresh.activeTurnId || null);
                }
            } catch (error) {
                if (selection.isCurrent(scope)) setQueueError(`Queue refresh failed: ${error.message}`);
            }
            if (queueMutation.current === token) {
                ++queueRevision.current; // Invalidate polls captured during the mutation too.
                queueMutation.current = null; setQueueBusy(false);
            }
        }
    };

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

    useLayoutEffect(()=>{
        if(!ready || !workspaceOpen)return;
        const sidebar=document.querySelector<HTMLElement>('.workspace-sidebar');
        if(sidebar)return bindWorkspaceVisibility(sidebar);
    },[ready,workspaceOpen]);

    useWorkspaceFolderReference(ready && workspaceOpen, sessionId, fileRefs, (path:string)=>{
        if (!selection.isCurrent(renderedSelection)) return;
        const refs=[...new Set([...getDraft(sessionId).fileRefs,path])];
        drafts.update(sessionId,{fileRefs:refs});setFileRefs(refs);
    });

    // ── Render ────────────────────────────────────────────────────────────────

    if (!ready) {
        return html`<div id="app"><div style="padding:20px;text-align:center;color:var(--text-secondary,#888)">Loading…</div></div>`;
    }

    return html`
        <div class=${appShellClass}>
            <style>${`.app-shell .post-content:has(table) { overflow-x: auto; } .app-shell .post-content table { display: table; width: 100%; table-layout: auto; }`}</style>
            <${SystemMetersHud} mode="overlay" />
            <${TimelineMenu}
                workspaceOpen=${workspaceOpen}
                toggleWorkspace=${() => setWorkspaceOpen((v: boolean) => !v)}
                chatOnlyMode=${false}
                openEditor=${openEditor}
            />
            <${WorkspaceExplorer}
                onFileSelect=${(path: string) => {
                    const refs = [...new Set([...getDraft(sessionId).fileRefs, path])];
                    drafts.update(sessionId, { fileRefs: refs }); setFileRefs(refs);
                }}
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
            <div class="container" ref=${containerRef}>
                <${Timeline}
                    posts=${posts}
                    hasMore=${false}
                    onLoadMore=${() => loadPosts({older:true})}
                    timelineRef=${timelineRef}
                    onHashtagClick=${() => {}}
                    onMessageRef=${(id: any) => {
                        const refs = [...new Set([...getDraft(sessionId).messageRefs, id])];
                        drafts.update(sessionId, { messageRefs: refs }); setMessageRefs(refs);
                    }}
                    onScrollToMessage=${() => {}}
                    onFileRef=${openEditor}
                    onPostClick=${undefined}
                    onDeletePost=${() => {}}
                    onOpenWidget=${(w: any) => setFloatingWidget(w)}
                    onOpenAttachmentPreview=${setAttachmentPreview}
                    emptyMessage=${searchState.active ? (searchState.query ? 'No matching messages.' : 'Enter a search query.') : 'Send a message to get started.'}
                    agents=${agents}
                    user=${userProfile}
                    reverse=${true}
                    removingPostIds=${new Set()}
                    searchQuery=${searchState.active ? searchState.query : ''}
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
                <${RunBoundQueueStack}
                    steerEnabled=${connectionStatus === 'connected' && isAgentTurnActive && !!queueActiveTurnId}
                    onInjectQueuedFollowup=${(item: any) => mutateQueue('steer', item)}
                    items=${[...followupQueueItems, ...optimisticQueue.filter(item => item.chat_jid === currentChatJid && !followupQueueItems.some(stored => stored.id === item.id || stored.metadata?.client_request_id === item.id))]}
                    busy=${queueBusy}
                    onReturnQueuedFollowup=${(item: any) => mutateQueue('return', item)}
                    onRemoveQueuedFollowup=${(item: any) => mutateQueue('remove', item)}
                    onMoveQueuedFollowup=${(from: number, to: number) => mutateQueue('move', from, to)}
                    onOpenFilePill=${openEditor}
                />
                ${followupQueueItems.some(item => item.phase === 'steer_returned') && html`<div role="alert">Steer was not consumed by its target run. The item remains queued and will not auto-send; return it to the editor, remove it, or Steer a new active run.</div>`}
                ${queueError && html`<div role="alert">${queueError}</div>`}
                ${newUIVersion && html`<div role="status" class="gi-version-warning">New UI available. Reload manually when ready; unsaved editor work may be lost.</div>`}
                ${sessionError && html`<div role="alert">${sessionError}</div>`}
                ${searchError && html`<div role="alert">${searchError}</div>`}
                ${searchState.active && html`<div role="status">Search${searchState.query ? `: ${searchState.query}` : ''} · ${searchState.scope} · up to 50 results</div>`}
                ${stopError && html`<div role="alert">${stopError}</div>`}
                ${compactError && html`<div role="alert">${compactError}</div>`}
                ${draftStorageError && html`<div role="alert">${draftStorageError}</div>`}
                ${drafts.error(sessionId) && html`<div role="alert">${drafts.error(sessionId)}</div>`}
                <${ComposeBox}
                    statusNotice=${notice}
                    showQueueStack=${false}
                    key=${`${sessionId}:${draftRestore?.sessionId === sessionId ? draftRestore.token : ''}`}
                    draftValue=${getDraft(sessionId).text}
                    draftMediaFiles=${getDraft(sessionId).media}
                    onContentChange=${(text: string) => drafts.update(sessionId, { text })}
                    onDraftMediaChange=${(media: File[]) => drafts.update(sessionId, { media })}
                    focusRestoredDraft=${draftRestore?.sessionId === sessionId}
                    onCaptureDraft=${(draft: any) => drafts.begin(sessionId, draft)}
                    onQueuedSubmissionStart=${(token: string, text: string) => {
                        if (!selection.isCurrent(renderedSelection)) return;
                        setOptimisticQueue(items => [...items, {id: token, content: text, chat_jid: currentChatJid, pending: true}]);
                    }}
                    onQueuedSubmissionEnd=${(token: string) => {
                        if (!selection.isCurrent(renderedSelection)) return;
                        setOptimisticQueue(items => items.filter(item => item.id !== token));
                        void refreshSelectedState();
                    }}
                    onDraftAccepted=${(token: string) => drafts.accepted(sessionId, token)}
                    onDraftFailed=${(token: string, error: string) => {
                        const draft = drafts.failed(sessionId, token, error);
                        if (selection.current() === sessionId) {
                            setFileRefs(draft.fileRefs); setMessageRefs(draft.messageRefs);
                            setDraftRestore({ sessionId, ...draft, token: crypto.randomUUID() });
                        }
                    }}
                    onDraftStorageError=${(error: any) => setDraftStorageError(`Send acknowledged, but draft cleanup failed: ${error.message}. Reload recovery may contain already-delivered text.`)}
                    currentChatJid=${currentChatJid}
                    isAgentActive=${isAgentTurnActive}
                    onPost=${handlePost}
                    onFocus=${() => { if (!isIOSDevice()) scrollToBottom(); }}
                    onModelMutationStart=${() => {
                        const token = {}; ++modelRevision.current; modelMutation.current = token; return token;
                    }}
                    onModelMutationEnd=${(token: any) => {
                        if (modelMutation.current === token) { ++modelRevision.current; modelMutation.current = null; }
                    }}
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
                            if (state.context_usage !== undefined) setContextUsage(state.context_usage);
                        }
                    }}
                    agents=${agents}
                    currentSessionAgent=${activeChatAgents.find((entry: any) => entry?.chat_jid === currentChatJid) || null}
                    agentStatus=${agentStatus}
                    agentDraft=${agentDraft}
                    contextUsage=${contextUsage}
                    activeEditorPath=${activeTabId}
                    onAttachEditorFile=${() => {
                        if (!activeTabId) return;
                        const refs = [...new Set([...getDraft(sessionId).fileRefs, activeTabId])];
                        drafts.update(sessionId, { fileRefs: refs }); setFileRefs(refs);
                    }}
                    fileRefs=${fileRefs}
                    messageRefs=${messageRefs}
                    onRemoveFileRef=${(p: string) => {
                        const refs = getDraft(sessionId).fileRefs.filter((x: string) => x !== p);
                        drafts.update(sessionId, { fileRefs: refs }); setFileRefs(refs);
                    }}
                    onClearFileRefs=${() => {
                        drafts.update(sessionId, { fileRefs: [] });
                        if (selection.current() === sessionId) setFileRefs([]);
                    }}
                    onSetFileRefs=${(refs: string[]) => {
                        drafts.update(sessionId, { fileRefs: refs });
                        if (selection.current() === sessionId) setFileRefs(refs);
                    }}
                    onRemoveMessageRef=${(id: any) => {
                        const refs = getDraft(sessionId).messageRefs.filter((x: any) => x !== id);
                        drafts.update(sessionId, { messageRefs: refs }); setMessageRefs(refs);
                    }}
                    onClearMessageRefs=${() => {
                        drafts.update(sessionId, { messageRefs: [] });
                        if (selection.current() === sessionId) setMessageRefs([]);
                    }}
                    onSetMessageRefs=${(refs: any[]) => {
                        drafts.update(sessionId, { messageRefs: refs });
                        if (selection.current() === sessionId) setMessageRefs(refs);
                    }}
                    connectionStatus=${connectionStatus}
                    activeChatAgents=${activeChatAgents}
                    currentChatBranches=${currentChatBranches}
                    onSwitchChat=${handleSwitchChat}
                    onCreateSession=${handleCreateSession}
                    onRenameSession=${(chatJid, title) => handleSessionMutation(chatJid, 'rename', title)}
                    onPinSession=${(chatJid, pinned) => handleSessionMutation(chatJid, 'pin', pinned)}
                    onArchiveSession=${chatJid => handleSessionMutation(chatJid, 'archive')}
                    onRestoreSession=${chatJid => handleSessionMutation(chatJid, 'restore')}
                    formatBranchPickerLabel=${(b: any) => b?.label || b?.chat_jid || ''}
                    handleBranchPickerChange=${() => {}}
                    searchMode=${searchState.active}
                    onEnterSearch=${enterSearch}
                    onExitSearch=${exitSearch}
                    onSearch=${(query:string) => runSearch(query)}
                    searchScope=${searchState.scope}
                    onSearchScopeChange=${(scope:string) => runSearch(undefined,scope)}
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
