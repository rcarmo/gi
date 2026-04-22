export interface FloatingWidgetDashboardBuildInput {
  generatedAt: string;
  request: unknown;
  currentChatJid: string;
  currentRootChatJid: string | null;
  statusPayload?: Record<string, any> | null;
  contextPayload?: Record<string, any> | null;
  queuePayload?: Record<string, any> | null;
  modelsPayload?: Record<string, any> | null;
  activeChatsPayload?: Record<string, any> | null;
  branchesPayload?: Record<string, any> | null;
  timelinePayload?: Record<string, any> | null;
  rawPosts?: Array<Record<string, any>> | null;
  activeChatAgents?: Array<Record<string, any>> | null;
  currentChatBranches?: Array<Record<string, any>> | null;
  contextUsage?: Record<string, any> | null;
  followupQueueItems?: Array<Record<string, any>> | null;
  activeModel?: unknown;
  activeThinkingLevel?: unknown;
  supportsThinking?: unknown;
  isAgentTurnActive?: boolean;
}

export function readFulfilledResult<T>(result: PromiseSettledResult<T>): T | null {
  return result.status === 'fulfilled' ? result.value : null;
}

function clampPercent(value: number): number {
  return Math.max(0, Math.min(100, value));
}

export function buildFloatingWidgetDashboardSnapshot(input: FloatingWidgetDashboardBuildInput): Record<string, unknown> {
  const posts = Array.isArray(input.timelinePayload?.posts)
    ? input.timelinePayload.posts
    : (Array.isArray(input.rawPosts) ? input.rawPosts : []);
  const latestPost = posts.length ? posts[posts.length - 1] : null;
  const botPosts = posts.filter((post) => post?.data?.is_bot_message).length;
  const userPosts = posts.filter((post) => !post?.data?.is_bot_message).length;

  const queueCount = Number(input.queuePayload?.count ?? input.followupQueueItems?.length ?? 0) || 0;
  const activeChatsCount = Array.isArray(input.activeChatsPayload?.chats)
    ? input.activeChatsPayload.chats.length
    : (Array.isArray(input.activeChatAgents) ? input.activeChatAgents.length : 0);
  const branchCount = Array.isArray(input.branchesPayload?.chats)
    ? input.branchesPayload.chats.length
    : (Array.isArray(input.currentChatBranches) ? input.currentChatBranches.length : 0);

  const contextPercent = Number(input.contextPayload?.percent ?? input.contextUsage?.percent ?? 0) || 0;
  const contextTokens = Number(input.contextPayload?.tokens ?? input.contextUsage?.tokens ?? 0) || 0;
  const contextWindow = Number(input.contextPayload?.contextWindow ?? input.contextUsage?.contextWindow ?? 0) || 0;

  const modelName = input.modelsPayload?.current ?? input.activeModel ?? null;
  const thinkingLevel = input.modelsPayload?.thinking_level ?? input.activeThinkingLevel ?? null;
  const supportsThinkingValue = input.modelsPayload?.supports_thinking ?? input.supportsThinking;
  const agentState = input.statusPayload?.status || (input.isAgentTurnActive ? 'active' : 'idle');
  const agentPhase = input.statusPayload?.data?.type || input.statusPayload?.type || null;

  return {
    generatedAt: input.generatedAt,
    request: input.request,
    chat: {
      currentChatJid: input.currentChatJid,
      rootChatJid: input.currentRootChatJid,
      activeChats: activeChatsCount,
      branches: branchCount,
    },
    agent: {
      status: agentState,
      phase: agentPhase,
      running: Boolean(input.isAgentTurnActive),
    },
    model: {
      current: modelName,
      thinkingLevel,
      supportsThinking: Boolean(supportsThinkingValue),
    },
    context: {
      tokens: contextTokens,
      contextWindow,
      percent: contextPercent,
    },
    queue: {
      count: queueCount,
    },
    timeline: {
      loadedPosts: posts.length,
      botPosts,
      userPosts,
      latestPostId: latestPost?.id ?? null,
      latestTimestamp: latestPost?.timestamp ?? null,
    },
    bars: [
      { key: 'context', label: 'Context', value: clampPercent(Math.round(contextPercent)) },
      { key: 'queue', label: 'Queue', value: clampPercent(queueCount * 18) },
      { key: 'activeChats', label: 'Active chats', value: clampPercent(activeChatsCount * 12) },
      { key: 'posts', label: 'Timeline load', value: clampPercent(posts.length * 5) },
    ],
  };
}
