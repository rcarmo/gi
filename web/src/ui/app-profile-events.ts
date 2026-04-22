export interface AgentProfileEntryLike {
  id?: unknown;
  name?: unknown;
  avatar_url?: unknown;
  avatarUrl?: unknown;
  avatar?: unknown;
  [key: string]: unknown;
}

export interface AgentProfilePatch {
  agentId: string;
  nameChanged: boolean;
  avatarChanged: boolean;
  resolvedName: unknown;
  resolvedAvatar: string | null;
}

export interface UserProfileLike {
  name?: string | null;
  avatar_url?: string | null;
  avatar_background?: string | null;
}

export interface AgentBrandingPayload {
  name: unknown;
  avatarUrl: unknown;
}

export function resolveDefaultAgentBrandingPayload(agents: unknown): AgentBrandingPayload {
  const list = Array.isArray(agents) ? agents : [];
  const defaultAgent = list.find((agent) => agent?.id === 'default');
  return {
    name: defaultAgent?.name,
    avatarUrl: defaultAgent?.avatar_url,
  };
}

export function resolveAgentProfilePatch(
  payload: Record<string, unknown> | null | undefined,
  currentEntry: AgentProfileEntryLike | null | undefined,
): AgentProfilePatch | null {
  if (!payload || typeof payload !== 'object') return null;

  const rawAgentId = payload.agent_id;
  if (!rawAgentId) return null;
  const agentId = String(rawAgentId);

  const nextName = payload.agent_name;
  const nextAvatar = payload.agent_avatar;
  if (!nextName && nextAvatar === undefined) return null;

  const current = currentEntry || { id: agentId };
  let resolvedName = current.name || null;
  let resolvedAvatar = current.avatar_url ?? current.avatarUrl ?? current.avatar ?? null;
  let nameChanged = false;
  let avatarChanged = false;

  if (nextName && nextName !== current.name) {
    resolvedName = nextName;
    nameChanged = true;
  }

  if (nextAvatar !== undefined) {
    const normalizedAvatar = typeof nextAvatar === 'string' ? nextAvatar.trim() : null;
    const normalizedCurrent = typeof resolvedAvatar === 'string' ? resolvedAvatar.trim() : null;
    const nextValue = normalizedAvatar || null;
    const currentValue = normalizedCurrent || null;
    if (nextValue !== currentValue) {
      resolvedAvatar = nextValue;
      avatarChanged = true;
    }
  }

  if (!nameChanged && !avatarChanged) return null;

  return {
    agentId,
    nameChanged,
    avatarChanged,
    resolvedName,
    resolvedAvatar: resolvedAvatar as string | null,
  };
}

export function resolveUserProfileFromAgentsPayload(
  previous: UserProfileLike,
  payload: Record<string, unknown> | null | undefined,
): UserProfileLike {
  const nextName = typeof payload?.name === 'string' && payload.name.trim() ? payload.name.trim() : 'You';
  const nextAvatar = typeof payload?.avatar_url === 'string' ? payload.avatar_url.trim() : null;
  const nextBackground = typeof payload?.avatar_background === 'string' && payload.avatar_background.trim()
    ? payload.avatar_background.trim()
    : null;

  if (
    previous.name === nextName
    && previous.avatar_url === nextAvatar
    && previous.avatar_background === nextBackground
  ) {
    return previous;
  }

  return {
    name: nextName,
    avatar_url: nextAvatar,
    avatar_background: nextBackground,
  };
}

export function resolveUserProfileUpdate(
  previous: UserProfileLike,
  payload: Record<string, unknown> | null | undefined,
): UserProfileLike {
  if (!payload || typeof payload !== 'object') return previous;

  const nextName = payload.user_name ?? payload.userName;
  const nextAvatar = payload.user_avatar ?? payload.userAvatar;
  const nextBackground = payload.user_avatar_background ?? payload.userAvatarBackground;
  if (nextName === undefined && nextAvatar === undefined && nextBackground === undefined) return previous;

  const resolvedName = typeof nextName === 'string' && nextName.trim()
    ? nextName.trim()
    : previous.name || 'You';
  const resolvedAvatar = nextAvatar === undefined
    ? previous.avatar_url
    : (typeof nextAvatar === 'string' && nextAvatar.trim() ? nextAvatar.trim() : null);
  const resolvedBackground = nextBackground === undefined
    ? previous.avatar_background
    : (typeof nextBackground === 'string' && nextBackground.trim() ? nextBackground.trim() : null);

  if (
    previous.name === resolvedName
    && previous.avatar_url === resolvedAvatar
    && previous.avatar_background === resolvedBackground
  ) {
    return previous;
  }

  return {
    name: resolvedName,
    avatar_url: resolvedAvatar,
    avatar_background: resolvedBackground,
  };
}
