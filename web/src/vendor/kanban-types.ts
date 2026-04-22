/**
 * Kanban Board Types
 * Based on Obsidian Kanban plugin format
 */

export interface CardData {
  id: string;
  title: string;
  checked: boolean;
  checkChar: string;  // ' ' or 'x' or other
}

export interface LaneData {
  id: string;
  title: string;
  cards: CardData[];
  collapsed?: boolean;
}

export interface BoardSettings {
  'lane-width'?: number;
  'show-checkboxes'?: boolean;
  'hide-card-count'?: boolean;
  'show-archive'?: boolean;
}

export interface BoardData {
  lanes: LaneData[];
  archive: CardData[];
  settings: BoardSettings;
}

export function createId(prefix: 'card' | 'lane'): string {
  return `${prefix}-${crypto.randomUUID()}`;
}

// Default templates
export function createCard(title: string = 'New Card'): CardData {
  return {
    id: createId('card'),
    title,
    checked: false,
    checkChar: ' ',
  };
}

export function createLane(title: string = 'New Lane'): LaneData {
  return {
    id: createId('lane'),
    title,
    cards: [],
  };
}

export function createBoard(): BoardData {
  return {
    lanes: [],
    archive: [],
    settings: {},
  };
}
