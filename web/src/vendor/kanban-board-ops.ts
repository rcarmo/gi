export interface KanbanCardData {
  id: string;
  title: string;
  checked: boolean;
  checkChar: string;
}

export interface KanbanLaneData {
  id: string;
  title: string;
  cards: KanbanCardData[];
  collapsed?: boolean;
}

export interface KanbanBoardData {
  lanes: KanbanLaneData[];
  archive: KanbanCardData[];
  settings: Record<string, any>;
}

export interface MoveCardParams {
  cardId: string;
  fromLaneId: string;
  toLaneId: string;
  toIndex?: number;
}

export interface MoveLaneParams {
  fromLaneId: string;
  toLaneId: string;
}

export function moveCardInBoard(board: KanbanBoardData, params: MoveCardParams): KanbanBoardData {
  if (!board) return board;

  const fromLane = board.lanes.find((lane) => lane.id === params.fromLaneId);
  const toLane = board.lanes.find((lane) => lane.id === params.toLaneId);
  if (!fromLane || !toLane) return board;

  const fromIndex = fromLane.cards.findIndex((entry) => entry.id === params.cardId);
  if (fromIndex === -1) return board;

  const card = fromLane.cards[fromIndex];
  if (!card) return board;

  const rawToIndex = Number.isFinite(params.toIndex)
    ? Number(params.toIndex)
    : toLane.cards.length;

  if (params.fromLaneId === params.toLaneId) {
    const nextCards = [...fromLane.cards];
    const [movedCard] = nextCards.splice(fromIndex, 1);
    if (!movedCard) return board;
    const clampedIndex = Math.max(0, Math.min(rawToIndex, fromLane.cards.length));
    const insertIndex = clampedIndex > fromIndex ? clampedIndex - 1 : clampedIndex;
    if (insertIndex === fromIndex) return board;
    nextCards.splice(insertIndex, 0, movedCard);
    return {
      ...board,
      lanes: board.lanes.map((lane) => (
        lane.id === params.fromLaneId ? { ...lane, cards: nextCards } : lane
      )),
    };
  }

  const targetCards = [...toLane.cards.filter((entry) => entry.id !== params.cardId)];
  const insertIndex = Math.max(0, Math.min(rawToIndex, targetCards.length));
  targetCards.splice(insertIndex, 0, card);

  return {
    ...board,
    lanes: board.lanes.map((lane) => {
      if (lane.id === params.fromLaneId) {
        return { ...lane, cards: lane.cards.filter((entry) => entry.id !== params.cardId) };
      }
      if (lane.id === params.toLaneId) {
        return { ...lane, cards: targetCards };
      }
      return lane;
    }),
  };
}

export function moveLaneInBoard(board: KanbanBoardData, params: MoveLaneParams): KanbanBoardData {
  if (!board || params.fromLaneId === params.toLaneId) return board;

  const fromIndex = board.lanes.findIndex((lane) => lane.id === params.fromLaneId);
  const toIndex = board.lanes.findIndex((lane) => lane.id === params.toLaneId);
  if (fromIndex === -1 || toIndex === -1) return board;

  const nextLanes = [...board.lanes];
  const [moved] = nextLanes.splice(fromIndex, 1);
  if (!moved) return board;
  nextLanes.splice(toIndex, 0, moved);
  return { ...board, lanes: nextLanes };
}
