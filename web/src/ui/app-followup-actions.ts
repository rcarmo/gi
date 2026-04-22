import { removeFollowupQueueRow, type FollowupQueueItemLike } from './app-followup-queue.js';

export interface FollowupQueueRemovalPlan<T extends FollowupQueueItemLike> {
  rowId: string | number;
  items: T[];
  remainingQueueCount: number;
}

export interface FollowupActionFailureToast {
  title: string;
  detail: string;
}

export function resolveFollowupQueueRemovalPlan<T extends FollowupQueueItemLike>(
  items: T[] | null | undefined,
  payload: Record<string, unknown> | null | undefined,
): FollowupQueueRemovalPlan<T> | null {
  const rowId = payload?.row_id;
  if (rowId == null || (typeof rowId !== 'string' && typeof rowId !== 'number')) {
    return null;
  }

  const removal = removeFollowupQueueRow(items, rowId);
  return {
    rowId,
    items: removal.items,
    remainingQueueCount: removal.remainingQueueCount,
  };
}

export function resolveFollowupActionFailureToast(action: 'steer' | 'remove'): FollowupActionFailureToast {
  if (action === 'steer') {
    return {
      title: 'Failed to steer message',
      detail: 'The queued message could not be sent as steering.',
    };
  }

  return {
    title: 'Failed to remove message',
    detail: 'The queued message could not be removed.',
  };
}
