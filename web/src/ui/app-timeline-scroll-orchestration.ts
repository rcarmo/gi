import { useCallback } from '../vendor/preact-htm.js';
import { filterQueuedTimelinePosts } from './app-followup-queue.js';

interface RefBox<T> {
  current: T;
}

export function shouldAutoScrollToBottom(scrollTop: number, threshold = 150): boolean {
  return Math.abs(scrollTop) <= threshold;
}

export function useTimelineScrollOrchestration(options: {
  timelineRef: RefBox<HTMLElement | null>;
  viewStateRef: RefBox<Record<string, unknown> | null | undefined>;
  followupQueueRowIdsRef: RefBox<Set<string | number>>;
}) {
  const {
    timelineRef,
    viewStateRef,
    followupQueueRowIdsRef,
  } = options;

  const scrollToBottom = useCallback(() => {
    const el = timelineRef.current;
    if (!el) return;
    if (shouldAutoScrollToBottom(el.scrollTop)) {
      el.scrollTop = 0;
    }
  }, [timelineRef]);

  const preserveTimelineScroll = useCallback((mutate?: (() => void) | null) => {
    const container = timelineRef.current;
    if (!container || typeof mutate !== 'function') {
      mutate?.();
      return;
    }
    const { currentHashtag: activeHashtag, searchQuery: activeSearch, searchOpen: activeSearchOpen } = viewStateRef.current || {};
    const reverseTimeline = !((activeSearch || activeSearchOpen) && !activeHashtag);
    const anchor = reverseTimeline
      ? container.scrollHeight - container.scrollTop
      : container.scrollTop;

    mutate();

    requestAnimationFrame(() => {
      const target = timelineRef.current;
      if (!target) return;
      if (reverseTimeline) {
        const nextTop = Math.max(target.scrollHeight - anchor, 0);
        target.scrollTop = nextTop;
      } else {
        const maxScroll = Math.max(target.scrollHeight - target.clientHeight, 0);
        const nextTop = Math.min(anchor, maxScroll);
        target.scrollTop = nextTop;
      }
    });
  }, [timelineRef, viewStateRef]);

  const preserveTimelineScrollTop = useCallback((mutate?: (() => void) | null) => {
    const container = timelineRef.current;
    if (!container || typeof mutate !== 'function') {
      mutate?.();
      return;
    }
    const anchor = container.scrollTop;
    mutate();
    requestAnimationFrame(() => {
      const target = timelineRef.current;
      if (!target) return;
      const maxScroll = Math.max(target.scrollHeight - target.clientHeight, 0);
      target.scrollTop = Math.min(anchor, maxScroll);
    });
  }, [timelineRef]);

  const filterQueuedPosts = useCallback((items: any[] | null) => {
    return filterQueuedTimelinePosts(items, followupQueueRowIdsRef.current);
  }, [followupQueueRowIdsRef]);

  return {
    scrollToBottom,
    preserveTimelineScroll,
    preserveTimelineScrollTop,
    filterQueuedPosts,
  };
}
