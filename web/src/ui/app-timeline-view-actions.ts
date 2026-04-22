import { useCallback, useMemo } from '../vendor/preact-htm.js';
import {
  backToTimeline,
  deleteTimelinePost,
  loadHashtagTimeline,
  searchTimeline,
} from './app-timeline-actions.js';

type StateSetter<T> = (next: T | ((prev: T) => T)) => void;

export function isMainTimelineViewState(options: {
  currentHashtag: string | null;
  searchQuery: string | null;
  searchOpen: boolean;
}): boolean {
  const { currentHashtag, searchQuery, searchOpen } = options;
  return !currentHashtag && !searchQuery && !searchOpen;
}

interface UseTimelineViewActionsOptions {
  currentHashtag: string | null;
  searchQuery: string | null;
  searchOpen: boolean;
  searchScope: string;
  currentChatJid: string;
  currentRootChatJid: string;
  posts: any[] | null;
  loadPosts: (hashtag?: string | null) => Promise<void>;
  searchPosts: (
    query: string,
    limit: number,
    offset: number,
    chatJid: string,
    scope: string,
    rootChatJid: string,
  ) => Promise<{ results?: any[] }>;
  setCurrentHashtag: (value: string | null) => void;
  setSearchQuery: (value: string | null) => void;
  setSearchOpen: (open: boolean) => void;
  setSearchScope: (scope: string) => void;
  setPosts: StateSetter<any[] | null>;
  setHasMore: (next: boolean) => void;
  preserveTimelineScrollTop: (mutate: () => void) => void;
  setRemovingPostIds: StateSetter<Set<string | number>>;
  deletePost: (postId: string | number, deleteReplies: boolean, chatJid: string) => Promise<{ ids?: Array<string | number> }>;
  hasMoreRef: { current: boolean };
  loadMoreRef: { current: ((options?: Record<string, unknown>) => void) | null };
}

export function useTimelineViewActions(options: UseTimelineViewActionsOptions) {
  const {
    currentHashtag,
    searchQuery,
    searchOpen,
    searchScope,
    currentChatJid,
    currentRootChatJid,
    posts,
    loadPosts,
    searchPosts,
    setCurrentHashtag,
    setSearchQuery,
    setSearchOpen,
    setSearchScope,
    setPosts,
    setHasMore,
    preserveTimelineScrollTop,
    setRemovingPostIds,
    deletePost,
    hasMoreRef,
    loadMoreRef,
  } = options;

  const handleHashtagClick = useCallback(async (hashtag: string) => {
    await loadHashtagTimeline({
      hashtag,
      setCurrentHashtag: setCurrentHashtag as (value: string) => void,
      setPosts,
      loadPosts,
    });
  }, [loadPosts, setCurrentHashtag, setPosts]);

  const handleBackToTimeline = useCallback(async () => {
    await backToTimeline({
      setCurrentHashtag,
      setSearchQuery,
      setPosts,
      loadPosts: loadPosts as () => Promise<void>,
    });
  }, [loadPosts, setCurrentHashtag, setPosts, setSearchQuery]);

  const handleSearch = useCallback(async (query: unknown, scope = searchScope) => {
    await searchTimeline({
      query,
      scope,
      currentChatJid,
      currentRootChatJid,
      searchPosts,
      setSearchScope,
      setSearchQuery,
      setCurrentHashtag,
      setPosts,
      setHasMore,
    });
  }, [currentChatJid, currentRootChatJid, searchPosts, searchScope, setCurrentHashtag, setHasMore, setPosts, setSearchQuery, setSearchScope]);

  const handleSearchScopeChange = useCallback((newScope: string) => {
    setSearchScope(newScope);
    // Re-run search with the new scope if there's an active query
    if (searchQuery && searchQuery.trim()) {
      void searchTimeline({
        query: searchQuery,
        scope: newScope,
        currentChatJid,
        currentRootChatJid,
        searchPosts,
        setSearchScope,
        setSearchQuery,
        setCurrentHashtag,
        setPosts,
        setHasMore,
      });
    }
  }, [currentChatJid, currentRootChatJid, searchPosts, searchQuery, setCurrentHashtag, setHasMore, setPosts, setSearchQuery, setSearchScope]);

  const enterSearchMode = useCallback(() => {
    setSearchOpen(true);
    setSearchQuery(null);
    setCurrentHashtag(null);
    setSearchScope('current');
    setPosts([]);
  }, [setCurrentHashtag, setPosts, setSearchOpen, setSearchQuery, setSearchScope]);

  const exitSearchMode = useCallback(() => {
    setSearchOpen(false);
    setSearchQuery(null);
    void loadPosts();
  }, [loadPosts, setSearchOpen, setSearchQuery]);

  const isMainTimelineView = useMemo(() => isMainTimelineViewState({
    currentHashtag,
    searchQuery,
    searchOpen,
  }), [currentHashtag, searchOpen, searchQuery]);

  const handleDeletePost = useCallback(async (post: any) => {
    await deleteTimelinePost({
      post,
      posts,
      currentChatJid,
      deletePost,
      preserveTimelineScrollTop,
      setPosts,
      setRemovingPostIds,
      hasMoreRef,
      loadMoreRef,
    });
  }, [currentChatJid, deletePost, hasMoreRef, loadMoreRef, posts, preserveTimelineScrollTop, setPosts, setRemovingPostIds]);

  return {
    handleHashtagClick,
    handleBackToTimeline,
    handleSearch,
    handleSearchScopeChange,
    enterSearchMode,
    exitSearchMode,
    isMainTimelineView,
    handleDeletePost,
  };
}
