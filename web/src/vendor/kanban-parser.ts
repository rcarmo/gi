/**
 * Kanban Markdown Parser/Serializer
 * 
 * Format (compatible with Obsidian Kanban):
 * 
 * ---
 * kanban-plugin: board
 * ---
 * 
 * ## Lane Title
 * 
 * - [ ] Card 1
 * - [x] Completed card
 * - [ ] Card with
 *   multiple lines
 * 
 * ## Another Lane
 * 
 * - [ ] More cards
 * 
 * ---
 * 
 * ## Archive
 * 
 * - [x] Archived item
 * 
 * %% kanban:settings
 * ```
 * {"lane-width": 300}
 * ```
 * %%
 */

import { BoardData, BoardSettings, CardData, LaneData, createBoard, createCard, createLane } from './kanban-types.js';

const ARCHIVE_MARKER = '---';

function stripFrontmatter(md: string): string {
  if (!md.startsWith('---\n')) {
    return md;
  }

  const endIndex = md.indexOf('\n---\n', 4);
  if (endIndex === -1) {
    return md;
  }

  return md.slice(endIndex + 5);
}

function extractSettings(content: string): { settings: BoardSettings; content: string } {
  const settingsStart = content.indexOf('%% kanban:settings');
  if (settingsStart === -1) {
    return { settings: {}, content };
  }

  const codeStart = content.indexOf('```', settingsStart);
  if (codeStart === -1) {
    return { settings: {}, content };
  }

  const codeEnd = content.indexOf('```', codeStart + 3);
  if (codeEnd === -1) {
    return { settings: {}, content };
  }

  const settingsText = content.slice(codeStart + 3, codeEnd).trim();
  let settings: BoardSettings = {};
  try {
    settings = JSON.parse(settingsText) as BoardSettings;
  } catch (e) {
    console.warn('Failed to parse kanban settings:', e);
  }

  const endMarker = content.indexOf('%%', codeEnd + 3);
  const removeEnd = endMarker === -1 ? codeEnd + 3 : endMarker + 2;
  const cleaned = `${content.slice(0, settingsStart).trimEnd()}\n${content.slice(removeEnd).trimStart()}`.trim();

  return { settings, content: cleaned };
}

function escapeMarkdownLine(line: string): string {
  if (line.startsWith('\\#') || line.startsWith('\\---')) {
    return line;
  }
  if (line.startsWith('#') || line.startsWith('---')) {
    return `\\${line}`;
  }
  return line;
}

function unescapeMarkdownLine(line: string): string {
  if (line.startsWith('\\#') || line.startsWith('\\---')) {
    return line.slice(1);
  }
  return line;
}

export function parseMarkdown(md: string): BoardData {
  const board = createBoard();

  // Remove frontmatter
  let content = stripFrontmatter(md);

  // Extract settings
  const extracted = extractSettings(content);
  board.settings = extracted.settings;
  content = extracted.content;

  // Split by archive marker
  const parts = content.split(/\n---\n/);
  const mainContent = parts[0];
  const archiveContent = parts.length > 1 ? parts.slice(1).join('\n---\n') : '';
  
  // Parse lanes and cards from main content
  board.lanes = parseLanes(mainContent);
  
  // Parse archive
  if (archiveContent) {
    const archiveLanes = parseLanes(archiveContent);
    const archiveLane = archiveLanes.find(l => l.title.toLowerCase() === 'archive');
    if (archiveLane) {
      board.archive = archiveLane.cards;
    }
  }
  
  return board;
}

function parseLanes(content: string): LaneData[] {
  const lanes: LaneData[] = [];
  
  // Split content by lane headers
  const sections = content.split(/(?=^## )/gm).filter(s => s.trim());
  
  for (const section of sections) {
    const lines = section.split('\n');
    const headerLine = lines[0];
    
    // Check if this is a lane header
    const headerMatch = headerLine.match(/^## (.+)$/);
    if (!headerMatch) continue;
    
    const lane = createLane(headerMatch[1].trim());
    
    // Parse cards from the rest of the section
    const cardContent = lines.slice(1).join('\n');
    lane.cards = parseCards(cardContent);
    
    lanes.push(lane);
  }
  
  return lanes;
}

function parseCards(content: string): CardData[] {
  const cards: CardData[] = [];
  
  // Match card items (- [ ] or - [x])
  const lines = content.split('\n');
  let currentCard: CardData | null = null;
  let currentTitle = '';
  
  for (const line of lines) {
    const cardMatch = line.match(/^- \[(.)\] (.*)$/);
    
    if (cardMatch) {
      // Save previous card if exists
      if (currentCard) {
        currentCard.title = currentTitle.trim();
        cards.push(currentCard);
      }
      
      // Start new card
      const checkChar = cardMatch[1];
      const firstLine = unescapeMarkdownLine(cardMatch[2]);
      currentCard = createCard(firstLine);
      currentCard.checkChar = checkChar;
      currentCard.checked = checkChar !== ' ';
      currentTitle = firstLine;
    } else if (currentCard && line.match(/^\s+\S/)) {
      // Continuation of multi-line card (indented)
      const continuation = unescapeMarkdownLine(line.replace(/^\s+/, ''));
      currentTitle += '\n' + continuation;
    }
  }
  
  // Don't forget the last card
  if (currentCard) {
    currentCard.title = currentTitle.trim();
    cards.push(currentCard);
  }
  
  return cards;
}

export function serializeMarkdown(board: BoardData): string {
  const lines: string[] = [];
  
  // Frontmatter
  lines.push('---');
  lines.push('');
  lines.push('kanban-plugin: board');
  lines.push('');
  lines.push('---');
  lines.push('');
  
  // Lanes
  for (const lane of board.lanes) {
    lines.push(`## ${lane.title}`);
    lines.push('');
    
    for (const card of lane.cards) {
      const checkbox = `[${card.checked ? card.checkChar || 'x' : ' '}]`;
      const titleLines = card.title.split('\n');
      lines.push(`- ${checkbox} ${escapeMarkdownLine(titleLines[0])}`);
      
      // Add indented continuation lines
      for (let i = 1; i < titleLines.length; i++) {
        lines.push(`  ${escapeMarkdownLine(titleLines[i])}`);
      }
    }
    
    lines.push('');
  }
  
  // Archive
  if (board.archive.length > 0) {
    lines.push('---');
    lines.push('');
    lines.push('## Archive');
    lines.push('');
    
    for (const card of board.archive) {
      const checkbox = `[${card.checked ? card.checkChar || 'x' : ' '}]`;
      const titleLines = card.title.split('\n');
      lines.push(`- ${checkbox} ${escapeMarkdownLine(titleLines[0])}`);
      for (let i = 1; i < titleLines.length; i++) {
        lines.push(`  ${escapeMarkdownLine(titleLines[i])}`);
      }
    }
    
    lines.push('');
  }
  
  // Settings
  if (Object.keys(board.settings).length > 0) {
    lines.push('%% kanban:settings');
    lines.push('```');
    lines.push(JSON.stringify(board.settings));
    lines.push('```');
    lines.push('%%');
  }
  
  return lines.join('\n');
}
