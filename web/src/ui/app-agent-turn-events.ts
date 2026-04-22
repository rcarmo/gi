export function shouldIgnoreMismatchedTurn(
  turnId: unknown,
  currentTurnId: unknown,
): boolean {
  return Boolean(turnId) && Boolean(currentTurnId) && turnId !== currentTurnId;
}

export function shouldAdoptIncomingTurn(
  turnId: unknown,
  currentTurnId: unknown,
): boolean {
  return Boolean(turnId) && !Boolean(currentTurnId);
}

export function resolveSteerQueuedTurnId(
  turnId: unknown,
  currentTurnId: unknown,
): unknown {
  return turnId || currentTurnId || null;
}
