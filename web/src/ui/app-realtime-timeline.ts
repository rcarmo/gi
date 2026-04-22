export interface TimelineViewStateLike {
  currentHashtag?: unknown;
  searchQuery?: unknown;
  searchOpen?: unknown;
}

export interface TimelinePostLike {
  id?: unknown;
  [key: string]: unknown;
}

export function isMainTimelineView(viewState: TimelineViewStateLike | null | undefined): boolean {
  return !viewState?.currentHashtag && !viewState?.searchQuery && !viewState?.searchOpen;
}

export function shouldAppendRealtimeTimelinePost(
  eventType: unknown,
  isCurrentChatEvent: boolean,
  isMainTimeline: boolean,
): boolean {
  return Boolean(
    isCurrentChatEvent
      && isMainTimeline
      && (eventType === 'new_post' || eventType === 'new_reply' || eventType === 'agent_response'),
  );
}

export function shouldMutateInteractionTimeline(
  isCurrentChatEvent: boolean,
  isMainTimeline: boolean,
): boolean {
  return isCurrentChatEvent && isMainTimeline;
}

export function appendUniqueTimelinePost<T extends TimelinePostLike>(
  posts: T[] | null | undefined,
  nextPost: T,
): T[] {
  if (!Array.isArray(posts) || posts.length === 0) {
    return [nextPost];
  }
  if (posts.some((post) => post?.id === nextPost?.id)) {
    return posts;
  }
  return [...posts, nextPost];
}

export function replaceTimelinePostById<T extends TimelinePostLike>(
  posts: T[] | null | undefined,
  nextPost: T,
): T[] | null | undefined {
  if (!Array.isArray(posts)) return posts;
  if (!posts.some((post) => post?.id === nextPost?.id)) return posts;
  return posts.map((post) => (post?.id === nextPost?.id ? nextPost : post));
}

export function removeTimelinePostsByIds<T extends TimelinePostLike>(
  posts: T[] | null | undefined,
  ids: unknown,
): T[] | null | undefined {
  if (!Array.isArray(posts)) return posts;
  const idList = Array.isArray(ids) ? ids : [];
  if (idList.length === 0) return posts;

  const idSet = new Set(idList);
  const filtered = posts.filter((post) => !idSet.has(post?.id));
  return filtered.length === posts.length ? posts : filtered;
}
