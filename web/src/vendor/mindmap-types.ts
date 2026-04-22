/**
 * TypeScript interfaces for the mindmap YAML schema
 */

export interface MindmapPosition {
  x: number;
  y: number;
}

export interface MindmapNode {
  id: string;
  text: string;
  children?: MindmapNode[];
  position?: MindmapPosition;  // Manual override position
  image?: string;              // Base64 or URL
  collapsed?: boolean;         // Whether children are hidden
}

export interface MindmapLink {
  from: string;   // Node ID
  to: string;     // Node ID
  label?: string; // Optional label
}

export type LayoutType = 'horizontal-tree' | 'vertical-tree' | 'radial' | 'force-directed';

export interface MindmapDocument {
  version: number;
  layout: LayoutType;
  root: MindmapNode;
  links?: MindmapLink[];
}

/**
 * Messages sent from extension to webview
 */
export type ExtensionToWebviewMessage =
  | { type: 'update'; content: string }
  | { type: 'setTheme'; theme: 'light' | 'dark' };

/**
 * Messages sent from webview to extension
 */
export type WebviewToExtensionMessage =
  | { type: 'ready' }
  | { type: 'edit'; content: string }
  | { type: 'log'; text: string }
  | { type: 'error'; text: string };
