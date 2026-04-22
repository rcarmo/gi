/**
 * Mindmap Editor
 *
 * D3-based mindmap renderer with keyboard navigation, drag-and-drop,
 * and cross-reference links.  Exposes a mount/update/destroy API on
 * window.__mindmapEditor for the pane host to drive.
 */

import {
  cloneMindmapHistoryDocument,
  recordMindmapHistory,
  redoMindmapHistory,
  undoMindmapHistory,
  type MindmapHistoryEntry,
} from './mindmap-history.js';
import { clearStoredNodePositions } from './mindmap-layout-utils.js';

// Declare globals provided by vendored scripts
declare const d3: any;
declare const jsyaml: any;

interface MindmapNode {
  id: string;
  text: string;
  children?: MindmapNode[];
  position?: { x: number; y: number };
  image?: string;
  collapsed?: boolean;
  side?: 'left' | 'right';
}

interface MindmapLink { from: string; to: string; label?: string; }

interface MindmapDocument {
  version: number;
  layout: 'horizontal-tree' | 'vertical-tree' | 'radial' | 'force-directed';
  root: MindmapNode;
  links?: MindmapLink[];
}

interface D3Node extends d3.HierarchyPointNode<MindmapNode> {
  _targetX?: number;
  _targetY?: number;
}

interface D3DragEvent { x: number; y: number; sourceEvent: MouseEvent & { target: Element }; }

interface MindmapMountOpts {
  content: string;
  isDark: boolean;
  onEdit: (yamlContent: string) => void;
  resolveImagePath: (relativePath: string) => string;
}

// ── Module state ────────────────────────────────────────────────

let _onEdit: ((yaml: string) => void) | null = null;
let _resolveImage: ((path: string) => string) | null = null;

// AbortController for all global event listeners — aborted on destroy
let _abort: AbortController | null = null;

// State
let mindmapData: MindmapDocument | null = null;
let selectedNodeId: string | null = null;
let editingNodeId: string | null = null;
let undoStack: MindmapHistoryEntry<MindmapDocument>[] = [];
let redoStack: MindmapHistoryEntry<MindmapDocument>[] = [];
const nodeMap: Map<string, D3Node> = new Map();
const imageCache: Map<string, string> = new Map(); // path -> resolved URI

// Link creation state
let isCreatingLink = false;
let linkSourceNodeId: string | null = null;
let linkPreviewLine: any = null;

// Clipboard state
let clipboardNode: MindmapNode | null = null;
let clipboardIsCut = false;

// D3 selections
let svg: any;
let zoomContainer: any;
let linksGroup: any;
let crosslinksGroup: any;
let nodesGroup: any;
let zoom: any;

// Constants
const NODE_PADDING = { x: 16, y: 8 };
const NODE_MIN_WIDTH = 80;
const NODE_HEIGHT = 32;
const NODE_IMAGE_SIZE = 24;
const BASE_SIBLING_SPACING = 20; // Base gap between siblings
const LEVEL_SPACING = 180;
const VERTICAL_TREE_LEVEL_SPACING = 56;

// Cache for calculated node dimensions
const nodeDimensionsCache: Map<string, { width: number; height: number }> = new Map();

/**
 * Calculate dimensions for a node based on its content
 */
function calculateNodeDimensions(node: MindmapNode): { width: number; height: number } {
  const cached = nodeDimensionsCache.get(node.id);
  if (cached) return cached;
  
  const hasImage = !!node.image;
  const lines = node.text.split('\n');
  const lineHeight = 18;
  
  // Check for inline images in markdown text
  const inlineImageMatches = node.text.match(/!\[.*?\]\(.*?\)/g) || [];
  const hasInlineImages = inlineImageMatches.length > 0;
  const imageHeight = 52;
  const imageWidth = 124;
  
  // Calculate text dimensions excluding image markdown
  let textLineCount = 0;
  let maxTextWidth = 0;
  for (const line of lines) {
    const strippedLine = stripMarkdown(line).trim();
    if (strippedLine) {
      textLineCount++;
      maxTextWidth = Math.max(maxTextWidth, measureText(strippedLine));
    }
  }
  
  // Width
  const textWidth = hasInlineImages 
    ? Math.max(maxTextWidth, imageWidth)
    : Math.max(maxTextWidth, 20);
  const contentWidth = hasImage ? textWidth + NODE_IMAGE_SIZE + 8 : textWidth;
  const nodeWidth = Math.max(NODE_MIN_WIDTH, contentWidth + NODE_PADDING.x * 2);
  
  // Height
  const textHeight = textLineCount * lineHeight;
  const totalImageHeight = inlineImageMatches.length * imageHeight;
  const contentHeight = textHeight + totalImageHeight;
  const nodeHeight = Math.max(NODE_HEIGHT, contentHeight + NODE_PADDING.y * 2, 
    hasImage ? NODE_IMAGE_SIZE + NODE_PADDING.y * 2 : NODE_HEIGHT);
  
  const dimensions = { width: nodeWidth, height: nodeHeight };
  nodeDimensionsCache.set(node.id, dimensions);
  return dimensions;
}

/**
 * Clear dimensions cache (call when nodes change)
 */
function clearDimensionsCache() {
  nodeDimensionsCache.clear();
}

function describeNodeTitle(node: MindmapNode | null | undefined): string {
  const text = String(node?.text || '').split('\n')[0].trim();
  if (!text) return 'node';
  return text.length > 40 ? `${text.slice(0, 39)}…` : text;
}

function resetHistoryState() {
  undoStack = [];
  redoStack = [];
  updateHistoryControls();
}

function pushHistoryDocument(document: MindmapDocument, label: string) {
  const next = recordMindmapHistory({ undoStack, redoStack }, document, label);
  undoStack = next.undoStack;
  redoStack = next.redoStack;
  updateHistoryControls();
}

function recordHistorySnapshot(label: string) {
  if (!mindmapData) return;
  pushHistoryDocument(mindmapData, label);
}

function applyUndo() {
  if (!mindmapData) return;
  const next = undoMindmapHistory({ undoStack, redoStack }, mindmapData);
  if (!next.restored) return;
  undoStack = next.undoStack;
  redoStack = next.redoStack;
  mindmapData = next.restored;
  editingNodeId = null;
  clearDimensionsCache();
  saveDocument();
  render();
  updateHistoryControls();
}

function applyRedo() {
  if (!mindmapData) return;
  const next = redoMindmapHistory({ undoStack, redoStack }, mindmapData);
  if (!next.restored) return;
  undoStack = next.undoStack;
  redoStack = next.redoStack;
  mindmapData = next.restored;
  editingNodeId = null;
  clearDimensionsCache();
  saveDocument();
  render();
  updateHistoryControls();
}

function updateHistoryControls() {
  const undoButton = document.getElementById('mindmap-undo') as HTMLButtonElement | null;
  const redoButton = document.getElementById('mindmap-redo') as HTMLButtonElement | null;
  const layoutSelect = document.getElementById('layout-select') as HTMLSelectElement | null;

  const latestUndo = undoStack.length > 0 ? undoStack[undoStack.length - 1] : null;
  const latestRedo = redoStack.length > 0 ? redoStack[redoStack.length - 1] : null;

  if (layoutSelect && mindmapData) {
    layoutSelect.value = mindmapData.layout;
  }
  if (undoButton) {
    undoButton.disabled = undoStack.length === 0;
    undoButton.title = latestUndo ? `Undo: ${latestUndo.label} (Ctrl+Z)` : 'Undo (Ctrl+Z)';
  }
  if (redoButton) {
    redoButton.disabled = redoStack.length === 0;
    redoButton.title = latestRedo ? `Redo: ${latestRedo.label} (Ctrl+Shift+Z)` : 'Redo (Ctrl+Shift+Z)';
  }
}

/**
 * Initialize the mindmap editor
 */
function init() {
  svg = d3.select('#mindmap-svg');
  if (svg.empty()) {
    console.warn('[mindmap] #mindmap-svg not found, cannot initialize.');
    return;
  }

  // Set up zoom behavior
  zoom = d3.zoom()
    .scaleExtent([0.1, 4])
    .on('zoom', (event: any) => {
      zoomContainer.attr('transform', event.transform);
    });

  svg.call(zoom);

  // Create container groups
  zoomContainer = svg.append('g').attr('class', 'zoom-container');
  linksGroup = zoomContainer.append('g').attr('class', 'links');
  crosslinksGroup = zoomContainer.append('g').attr('class', 'crosslinks');
  nodesGroup = zoomContainer.append('g').attr('class', 'nodes');

  // Add arrow marker for crosslinks
  const defs = svg.append('defs');
  defs.append('marker')
    .attr('id', 'arrow')
    .attr('viewBox', '0 -5 10 10')
    .attr('refX', 10)
    .attr('refY', 0)
    .attr('markerWidth', 6)
    .attr('markerHeight', 6)
    .attr('orient', 'auto')
    .append('path')
    .attr('d', 'M0,-5L10,0L0,5')
    .attr('class', 'arrow-marker');

  // Add link preview line
  linkPreviewLine = zoomContainer.append('line')
    .attr('class', 'link-preview')
    .attr('stroke', 'var(--crosslink-color)')
    .attr('stroke-width', 2)
    .attr('stroke-dasharray', '6 3')
    .style('display', 'none');

  // Set up event listeners
  setupKeyboardHandlers();
  setupToolbar();
  setupLinkCreation();
  setupContextMenu();

  // Configure marked for inline markdown (if available)
  if (typeof (globalThis as any).marked !== 'undefined' && (globalThis as any).marked.setOptions) {
    (globalThis as any).marked.setOptions({ breaks: true, gfm: true });
  }
}

/**
 * Set up keyboard event handlers
 */
function setupKeyboardHandlers() {
  const signal = _abort?.signal;
  document.addEventListener('keydown', (e: KeyboardEvent) => {
    // Only handle keys when the mindmap container (or its children) is focused
    const container = document.getElementById('mindmap-container');
    if (!container?.contains(e.target as Node)) return;

    // Ignore if editing text
    if (editingNodeId && e.target instanceof HTMLElement && e.target.isContentEditable) {
      if (e.key === 'Escape') {
        finishEditing();
        e.preventDefault();
      } else if (e.key === 'Enter' && e.shiftKey) {
        // Shift+Enter creates newline (let default behavior happen)
        return;
      } else if (e.key === 'Enter') {
        // Enter just finishes editing
        finishEditing();
        e.preventDefault();
      } else if (e.key === 'Tab') {
        e.preventDefault();
        finishEditing();
        if (e.shiftKey) {
          moveNodeUp();
        } else {
          createChildNode();
        }
      }
      return;
    }

    if (!mindmapData) return;

    // Handle Ctrl/Cmd shortcuts
    if (e.ctrlKey || e.metaKey) {
      switch (e.key.toLowerCase()) {
        case 'x':
          if (selectedNodeId && selectedNodeId !== 'root') {
            e.preventDefault();
            cutNode();
          }
          break;
        case 'c':
          if (selectedNodeId) {
            e.preventDefault();
            copyNode();
          }
          break;
        case 'v':
          if (selectedNodeId && clipboardNode) {
            e.preventDefault();
            pasteNode();
          }
          break;
        case 'z':
          e.preventDefault();
          if (e.shiftKey) applyRedo();
          else applyUndo();
          break;
        case 'y':
          e.preventDefault();
          applyRedo();
          break;
      }
      return;
    }

    switch (e.key) {
      case 'Tab':
        e.preventDefault();
        if (e.shiftKey) {
          moveNodeUp();
        } else {
          createChildNode();
        }
        break;

      case 'Enter':
        e.preventDefault();
        if (selectedNodeId) {
          createSiblingNode();
        }
        break;

      case 'Delete':
      case 'Backspace':
        if (selectedNodeId && selectedNodeId !== 'root') {
          e.preventDefault();
          deleteSelectedNode();
        }
        break;

      case 'Escape':
        selectedNodeId = null;
        render();
        break;

      case 'ArrowUp':
        e.preventDefault();
        navigateNode('up');
        break;

      case 'ArrowDown':
        e.preventDefault();
        navigateNode('down');
        break;

      case 'ArrowLeft':
        e.preventDefault();
        navigateNode('left');
        break;

      case 'ArrowRight':
        e.preventDefault();
        navigateNode('right');
        break;

      case 'F2':
        if (selectedNodeId) {
          startEditing(selectedNodeId);
        }
        break;

      default:
        // Don't start editing on direct typing - require double-click or F2
        break;
    }
  }, { signal });
}

/**
 * Set up toolbar event handlers
 */
function setupToolbar() {
  const layoutSelect = document.getElementById('layout-select') as HTMLSelectElement;
  layoutSelect?.addEventListener('change', () => {
    if (mindmapData && mindmapData.layout !== layoutSelect.value) {
      const nextLayout = layoutSelect.value as MindmapDocument['layout'];
      recordHistorySnapshot(`Changed layout to ${nextLayout}`);
      mindmapData.layout = nextLayout;
      clearStoredNodePositions(mindmapData.root);
      saveAndRender();
      updateHistoryControls();
      setTimeout(fitToView, 0);
    }
  });

  document.getElementById('zoom-fit')?.addEventListener('click', fitToView);
  document.getElementById('zoom-in')?.addEventListener('click', () => {
    svg.transition().duration(300).call(zoom.scaleBy, 1.3);
  });
  document.getElementById('zoom-out')?.addEventListener('click', () => {
    svg.transition().duration(300).call(zoom.scaleBy, 0.7);
  });
  document.getElementById('reset-layout')?.addEventListener('click', resetLayout);
  document.getElementById('mindmap-undo')?.addEventListener('click', applyUndo);
  document.getElementById('mindmap-redo')?.addEventListener('click', applyRedo);
  updateHistoryControls();
}

/**
 * Reset layout by removing all saved node positions
 */
function resetLayout() {
  if (!mindmapData) return;

  let changed = false;
  const clearPositions = (node: MindmapNode) => {
    if (node.position) {
      changed = true;
      delete node.position;
    }
    if (node.children) {
      node.children.forEach(clearPositions);
    }
  };

  const historySnapshot = cloneMindmapHistoryDocument(mindmapData);
  clearPositions(mindmapData.root);
  if (!changed) return;
  pushHistoryDocument(historySnapshot, 'Reset layout');
  saveAndRender();
  setTimeout(fitToView, 100);
}

/**
 * Set up context menu
 */
function setupContextMenu() {
  const contextMenu = document.getElementById('context-menu');
  if (!contextMenu) return;

  // Show context menu on right-click on nodes
  svg.on('contextmenu', (event: MouseEvent) => {
    event.preventDefault();
    
    // Check if clicked on a node
    const target = event.target as Element;
    const nodeElement = target.closest('.node');
    
    if (nodeElement) {
      const nodeData = d3.select(nodeElement).datum() as D3Node;
      if (nodeData) {
        selectedNodeId = nodeData.data.id;
        render();
      }
    }
    
    if (selectedNodeId) {
      showContextMenu(event.clientX, event.clientY);
    }
  });

  // Hide context menu on click elsewhere
  document.addEventListener('click', (event: MouseEvent) => {
    if (!contextMenu.contains(event.target as Node)) {
      hideContextMenu();
    }
  }, { signal: _abort?.signal });

  // Handle context menu actions
  contextMenu.addEventListener('click', (event: MouseEvent) => {
    const button = (event.target as Element).closest('button');
    if (!button) return;

    const action = button.getAttribute('data-action');
    hideContextMenu();

    switch (action) {
      case 'cut':
        cutNode();
        break;
      case 'copy':
        copyNode();
        break;
      case 'paste':
        pasteNode();
        break;
      case 'add-child':
        createChildNode();
        break;
      case 'add-sibling':
        createSiblingNode();
        break;
      case 'delete':
        deleteSelectedNode();
        break;
    }
  });
}

function showContextMenu(x: number, y: number) {
  const contextMenu = document.getElementById('context-menu');
  if (!contextMenu) return;

  // Update paste button state
  const pasteBtn = contextMenu.querySelector('[data-action="paste"]') as HTMLButtonElement;
  if (pasteBtn) {
    pasteBtn.disabled = !clipboardNode;
  }

  // Position menu
  contextMenu.style.left = `${x}px`;
  contextMenu.style.top = `${y}px`;
  contextMenu.classList.remove('hidden');

  // Adjust if menu goes off screen
  const rect = contextMenu.getBoundingClientRect();
  if (rect.right > window.innerWidth) {
    contextMenu.style.left = `${x - rect.width}px`;
  }
  if (rect.bottom > window.innerHeight) {
    contextMenu.style.top = `${y - rect.height}px`;
  }
}

function hideContextMenu() {
  const contextMenu = document.getElementById('context-menu');
  if (contextMenu) {
    contextMenu.classList.add('hidden');
  }
}

/**
 * Deep clone a node (for copy/paste)
 */
function cloneNode(node: MindmapNode): MindmapNode {
  const clone: MindmapNode = {
    id: `node-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
    text: node.text,
  };
  if (node.image) clone.image = node.image;
  if (node.children && node.children.length > 0) {
    clone.children = node.children.map(child => cloneNode(child));
  }
  return clone;
}

/**
 * Cut selected node to clipboard
 */
function cutNode() {
  if (!selectedNodeId || selectedNodeId === 'root' || !mindmapData) return;

  const node = findNode(mindmapData.root, selectedNodeId);
  if (node) {
    clipboardNode = cloneNode(node);
    clipboardIsCut = true;
    deleteSelectedNode();
  }
}

/**
 * Copy selected node to clipboard
 */
function copyNode() {
  if (!selectedNodeId || !mindmapData) return;

  const node = findNode(mindmapData.root, selectedNodeId);
  if (node) {
    clipboardNode = cloneNode(node);
    clipboardIsCut = false;
  }
}

/**
 * Paste node from clipboard as child of selected node
 */
function pasteNode() {
  if (!clipboardNode || !selectedNodeId || !mindmapData) return;

  const newNode = cloneNode(clipboardNode);
  
  const addToParent = (node: MindmapNode): boolean => {
    if (node.id === selectedNodeId) {
      if (!node.children) node.children = [];
      node.children.push(newNode);
      return true;
    }
    if (node.children) {
      for (const child of node.children) {
        if (addToParent(child)) return true;
      }
    }
    return false;
  };

  const historySnapshot = cloneMindmapHistoryDocument(mindmapData);
  if (!addToParent(mindmapData.root)) return;
  pushHistoryDocument(historySnapshot, `Pasted “${describeNodeTitle(newNode)}”`);
  selectedNodeId = newNode.id;
  saveAndRender();
}

/**
 * Find a node by ID
 */
function findNode(node: MindmapNode, id: string): MindmapNode | null {
  if (node.id === id) return node;
  if (node.children) {
    for (const child of node.children) {
      const found = findNode(child, id);
      if (found) return found;
    }
  }
  return null;
}

/**
 * Set up Ctrl+drag for link creation
 */
function setupLinkCreation() {
  // Track mouse position during link creation
  svg.on('mousemove.linkCreate', (event: MouseEvent) => {
    if (!isCreatingLink || !linkSourceNodeId) return;

    const [mouseX, mouseY] = d3.pointer(event, zoomContainer.node());

    const sourceNode = nodeMap.get(linkSourceNodeId);
    if (sourceNode) {
      linkPreviewLine
        .attr('x1', sourceNode.x)
        .attr('y1', sourceNode.y)
        .attr('x2', mouseX)
        .attr('y2', mouseY)
        .style('display', 'block');
    }

    // Highlight potential target nodes
    nodesGroup.selectAll('.node')
      .classed('link-target', (d: D3Node) => {
        if (d.data.id === linkSourceNodeId) return false;
        const dist = Math.sqrt(Math.pow(d.x - mouseX, 2) + Math.pow(d.y - mouseY, 2));
        return dist < 50;
      });
  });

  // Handle mouse up to complete link creation
  svg.on('mouseup.linkCreate', (event: MouseEvent) => {
    if (!isCreatingLink || !linkSourceNodeId) return;

    const [mouseX, mouseY] = d3.pointer(event, zoomContainer.node());

    // Find target node under mouse
    let targetNodeId: string | null = null;
    nodeMap.forEach((node, id) => {
      if (id === linkSourceNodeId) return;
      const dist = Math.sqrt(Math.pow(node.x - mouseX, 2) + Math.pow(node.y - mouseY, 2));
      if (dist < 50) {
        targetNodeId = id;
      }
    });

    if (targetNodeId) {
      createLink(linkSourceNodeId, targetNodeId);
    }

    // Reset link creation state
    cancelLinkCreation();
  });

  // Cancel on escape or right click
  document.addEventListener('keydown', (e: KeyboardEvent) => {
    if (e.key === 'Escape' && isCreatingLink) {
      cancelLinkCreation();
    }
  }, { signal: _abort?.signal });

  svg.on('contextmenu.linkCreate', () => {
    if (isCreatingLink) {
      cancelLinkCreation();
    }
  });
}

/**
 * Start creating a link from a node
 */
function startLinkCreation(nodeId: string) {
  isCreatingLink = true;
  linkSourceNodeId = nodeId;
  document.body.classList.add('creating-link');
}

/**
 * Cancel link creation
 */
function cancelLinkCreation() {
  isCreatingLink = false;
  linkSourceNodeId = null;
  linkPreviewLine?.style('display', 'none');
  nodesGroup.selectAll('.node').classed('link-target', false);
  document.body.classList.remove('creating-link');
}

/**
 * Create a cross-reference link between two nodes
 */
function createLink(fromId: string, toId: string, label?: string) {
  if (!mindmapData || fromId === toId) return;

  // Check if link already exists
  const existingLink = mindmapData.links?.find(
    l => (l.from === fromId && l.to === toId) || (l.from === toId && l.to === fromId),
  );
  if (existingLink) return;

  const historySnapshot = cloneMindmapHistoryDocument(mindmapData);
  if (!mindmapData.links) mindmapData.links = [];
  mindmapData.links.push({ from: fromId, to: toId, label });

  pushHistoryDocument(historySnapshot, 'Added cross-link');
  saveAndRender();
}

function normalizeNodeTree(node: any, isRoot = false): MindmapNode {
  const normalized: any = (node && typeof node === 'object') ? node : {};
  normalized.id = isRoot ? 'root' : (typeof normalized.id === 'string' && normalized.id.trim()
    ? normalized.id
    : `node-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`);
  normalized.text = typeof normalized.text === 'string'
    ? normalized.text
    : (normalized.text == null ? (isRoot ? 'Root' : 'New Node') : String(normalized.text));

  if (!Array.isArray(normalized.children)) normalized.children = [];
  normalized.children = normalized.children.map((child: any) => normalizeNodeTree(child, false));

  if (isRoot) {
    normalized.children.forEach((child: any, i: number) => {
      if (child.side !== 'left' && child.side !== 'right') {
        child.side = i % 2 === 0 ? 'right' : 'left';
      }
    });
  }

  return normalized as MindmapNode;
}

/**
 * Parse YAML content and render the mindmap
 */
function loadContent(content: string) {
  try {
    const parsed = jsyaml.load(content) as MindmapDocument | undefined;
    
    // Clear dimensions cache on load
    clearDimensionsCache();
    
    // Handle empty or invalid YAML
    if (!parsed || typeof parsed !== 'object') {
      mindmapData = {
        version: 1,
        layout: 'horizontal-tree',
        root: { id: 'root', text: 'Root', children: [] },
        links: []
      };
    } else {
      mindmapData = parsed;
      // Ensure required fields
      if (!mindmapData.version) mindmapData.version = 1;
      if (!mindmapData.layout) mindmapData.layout = 'horizontal-tree';
      if (!mindmapData.root) {
        mindmapData.root = { id: 'root', text: 'Root', children: [] };
      }
      if (!mindmapData.links) mindmapData.links = [];
    }

    // Normalize node/link structure so controls always have stable ids
    mindmapData.root = normalizeNodeTree(mindmapData.root, true);
    mindmapData.links = (mindmapData.links || []).filter((l: any) =>
      l && typeof l.from === 'string' && typeof l.to === 'string',
    );

    resetHistoryState();
    render();

    // Fit to view on first load
    setTimeout(fitToView, 100);
  } catch (error) {
    console.error('[mindmap] Failed to parse YAML:', error);
    mindmapData = {
      version: 1,
      layout: 'horizontal-tree',
      root: { id: 'root', text: 'Root', children: [] },
      links: [],
    };
    resetHistoryState();
    render();
  }
}

/**
 * Render the mindmap
 */
function render() {
  if (!mindmapData) return;

  // Build hierarchy
  const root = d3.hierarchy(mindmapData.root, (d: MindmapNode) =>
    d.collapsed ? [] : d.children,
  );

  // Apply layout
  let layoutFn: (root: any) => void;

  switch (mindmapData.layout) {
    case 'vertical-tree':
      layoutFn = applyVerticalTreeLayout;
      break;
    case 'radial':
      layoutFn = applyRadialLayout;
      break;
    case 'force-directed':
      layoutFn = applyForceLayout;
      break;
    case 'horizontal-tree':
    default:
      layoutFn = applyHorizontalTreeLayout;
      break;
  }

  layoutFn(root);

  // Build node map for quick lookup
  nodeMap.clear();
  root.each((node: D3Node) => {
    nodeMap.set(node.data.id, node);
  });

  // Render links
  renderLinks(root);

  // Render crosslinks
  renderCrosslinks();

  // Render nodes
  renderNodes(root);
}

/**
 * Apply horizontal tree layout with dynamic node sizing.
 *
 * NOTE: This must operate on real d3 hierarchy nodes. Creating plain objects
 * via `{ ...root }` strips hierarchy prototype methods (`eachBefore`, etc.)
 * and crashes d3 tree layout.
 */
function applyHorizontalTreeLayout(root: any) {
  // Pre-calculate all node dimensions
  root.each((node: D3Node) => {
    calculateNodeDimensions(node.data);
  });

  const treeLayout = d3.tree()
    .nodeSize([NODE_HEIGHT, LEVEL_SPACING])
    .separation((a: any, b: any) => {
      const aHeight = calculateNodeDimensions(a.data).height;
      const bHeight = calculateNodeDimensions(b.data).height;
      const maxHeight = Math.max(aHeight, bHeight);
      const baseSeparation = (maxHeight + BASE_SIBLING_SPACING) / NODE_HEIGHT;
      return a.parent === b.parent ? baseSeparation : baseSeparation * 1.2;
    });

  // Run layout on the original hierarchy node
  treeLayout(root);

  // Convert from vertical tree axes to horizontal axes and mirror left branches
  root.each((node: any) => {
    const temp = node.x;
    const depthAxis = node.y;

    // Determine side by the top-level branch under root
    let branch: any = node;
    while (branch.parent && branch.parent.parent) {
      branch = branch.parent;
    }
    const isLeft = branch.parent ? branch.data.side === 'left' : false;

    node.x = (isLeft ? -1 : 1) * depthAxis;
    node.y = temp;
  });

  // Keep root centered
  root.x = 0;
  root.y = 0;

  // Apply manual position overrides
  root.each((node: D3Node) => {
    if (node.data.position) {
      node.x = node.data.position.x;
      node.y = node.data.position.y;
    }
  });
}

/**
 * Apply vertical tree layout with dynamic node sizing
 */
function applyVerticalTreeLayout(root: any) {
  // Pre-calculate all node dimensions
  root.each((node: D3Node) => {
    calculateNodeDimensions(node.data);
  });
  
  const treeLayout = d3.tree()
    .nodeSize([NODE_MIN_WIDTH, VERTICAL_TREE_LEVEL_SPACING])
    .separation((a: any, b: any) => {
      // Get widths of both nodes (in vertical layout, width affects horizontal spacing)
      const aWidth = calculateNodeDimensions(a.data).width;
      const bWidth = calculateNodeDimensions(b.data).width;
      const maxWidth = Math.max(aWidth, bWidth);
      
      // Calculate separation based on actual node widths
      const baseSeparation = (maxWidth + BASE_SIBLING_SPACING) / NODE_MIN_WIDTH;
      return a.parent === b.parent ? baseSeparation : baseSeparation * 1.2;
    });

  treeLayout(root);

  root.each((node: D3Node) => {
    if (node.data.position) {
      node.x = node.data.position.x;
      node.y = node.data.position.y;
    }
  });
}

/**
 * Apply radial layout with dynamic sizing
 */
function applyRadialLayout(root: any) {
  // Pre-calculate all node dimensions
  root.each((node: D3Node) => {
    calculateNodeDimensions(node.data);
  });
  
  const radius = Math.min(800, 100 + root.descendants().length * 30);

  const treeLayout = d3.tree()
    .size([2 * Math.PI, radius])
    .separation((a: any, b: any) => {
      const aSize = calculateNodeDimensions(a.data);
      const bSize = calculateNodeDimensions(b.data);
      const maxSize = Math.max(aSize.width, aSize.height, bSize.width, bSize.height);
      const baseSep = (a.parent === b.parent ? 1 : 2) / Math.max(1, a.depth);
      return baseSep * (1 + maxSize / 200);
    });

  treeLayout(root);

  root.each((node: D3Node) => {
    if (node.data.position) {
      node.x = node.data.position.x;
      node.y = node.data.position.y;
    } else {
      // Convert polar to Cartesian
      const angle = node.x;
      const r = node.y;
      node.x = r * Math.cos(angle - Math.PI / 2);
      node.y = r * Math.sin(angle - Math.PI / 2);
    }
  });
}

/**
 * Apply force-directed layout with dynamic node sizing
 */
function applyForceLayout(root: any) {
  const nodes = root.descendants();
  const links = root.links();

  // Pre-calculate dimensions and initialize positions
  nodes.forEach((node: D3Node, i: number) => {
    calculateNodeDimensions(node.data);
    if (node.data.position) {
      node.x = node.data.position.x;
      node.y = node.data.position.y;
    } else {
      node.x = (i % 10) * 150;
      node.y = Math.floor(i / 10) * 100;
    }
  });

  // Simple force simulation with collision detection based on node size
  const simulation = d3.forceSimulation(nodes)
    .force('link', d3.forceLink(links).id((d: any) => d.data.id).distance(120))
    .force('charge', d3.forceManyBody().strength(-400))
    .force('center', d3.forceCenter(0, 0))
    .force('collision', d3.forceCollide().radius((d: any) => {
      const dims = calculateNodeDimensions(d.data);
      return Math.max(dims.width, dims.height) / 2 + 10;
    }))
    .stop();

  // Run simulation
  for (let i = 0; i < 120; i++) {
    simulation.tick();
  }
}

/**
 * Render parent-child links
 */
function renderLinks(root: any) {
  const links = root.links();

  const linkGenerator = mindmapData?.layout === 'horizontal-tree'
    ? d3.linkHorizontal().x((d: any) => d.x).y((d: any) => d.y)
    : d3.linkVertical().x((d: any) => d.x).y((d: any) => d.y);

  const linkSelection = linksGroup.selectAll('.link')
    .data(links, (d: any) => `${d.source.data.id}-${d.target.data.id}`);

  linkSelection.enter()
    .append('path')
    .attr('class', 'link')
    .merge(linkSelection)
    .transition()
    .duration(300)
    .attr('d', linkGenerator);

  linkSelection.exit().remove();
}

/**
 * Update links immediately without transition (for drag)
 */
function updateLinksImmediate() {
  if (!mindmapData) return;

  const linkGenerator = mindmapData.layout === 'horizontal-tree'
    ? d3.linkHorizontal().x((d: any) => d.x).y((d: any) => d.y)
    : d3.linkVertical().x((d: any) => d.x).y((d: any) => d.y);

  // Update tree links
  linksGroup.selectAll('.link')
    .attr('d', linkGenerator);

  // Update crosslinks
  crosslinksGroup.selectAll('.crosslink')
    .attr('d', (d: any) => {
      const sx = d.source.x;
      const sy = d.source.y;
      const tx = d.target.x;
      const ty = d.target.y;
      const mx = (sx + tx) / 2;
      const my = (sy + ty) / 2 - 30;
      return `M${sx},${sy} Q${mx},${my} ${tx},${ty}`;
    });

  // Update crosslink labels
  crosslinksGroup.selectAll('.crosslink-label')
    .attr('x', (d: any) => (d.source.x + d.target.x) / 2)
    .attr('y', (d: any) => (d.source.y + d.target.y) / 2 - 35);
}

/**
 * Render cross-reference links (splines)
 */
function renderCrosslinks() {
  if (!mindmapData?.links) return;

  const crosslinkData = mindmapData.links
    .map(link => ({
      ...link,
      source: nodeMap.get(link.from),
      target: nodeMap.get(link.to),
    }))
    .filter(link => link.source && link.target);

  // Draw paths
  const pathSelection = crosslinksGroup.selectAll('.crosslink')
    .data(crosslinkData, (d: any) => `${d.from}-${d.to}`);

  pathSelection.enter()
    .append('path')
    .attr('class', 'crosslink')
    .attr('marker-end', 'url(#arrow)')
    .merge(pathSelection)
    .transition()
    .duration(300)
    .attr('d', (d: any) => {
      const sx = d.source.x;
      const sy = d.source.y;
      const tx = d.target.x;
      const ty = d.target.y;
      const mx = (sx + tx) / 2;
      const my = (sy + ty) / 2 - 30;
      return `M${sx},${sy} Q${mx},${my} ${tx},${ty}`;
    });

  pathSelection.exit().remove();

  // Draw labels
  const labelSelection = crosslinksGroup.selectAll('.crosslink-label')
    .data(crosslinkData.filter(d => d.label), (d: any) => `${d.from}-${d.to}-label`);

  labelSelection.enter()
    .append('text')
    .attr('class', 'crosslink-label')
    .merge(labelSelection)
    .transition()
    .duration(300)
    .attr('x', (d: any) => (d.source.x + d.target.x) / 2)
    .attr('y', (d: any) => (d.source.y + d.target.y) / 2 - 20)
    .attr('text-anchor', 'middle')
    .text((d: any) => d.label);

  labelSelection.exit().remove();
}

/**
 * Render nodes
 */
function renderNodes(root: any) {
  const nodes = root.descendants();

  const nodeSelection = nodesGroup.selectAll('.node')
    .data(nodes, (d: any) => d.data.id);

  // Enter new nodes
  const nodeEnter = nodeSelection.enter()
    .append('g')
    .attr('class', 'node')
    .attr('transform', (d: any) => `translate(${d.x},${d.y})`)
    .on('click', (event: MouseEvent, d: any) => {
      event.stopPropagation();
      selectNode(d.data.id);
    })
    .on('dblclick', (event: MouseEvent, d: any) => {
      event.stopPropagation();
      startEditing(d.data.id);
    })
    .on('mousedown', (event: MouseEvent, d: any) => {
      // Ctrl+click to start link creation
      if (event.ctrlKey || event.metaKey) {
        event.stopPropagation();
        event.preventDefault();
        startLinkCreation(d.data.id);
      }
    })
    .call(d3.drag()
      .filter((event: any) => !event.ctrlKey && !event.metaKey) // Don't drag when Ctrl is held
      .on('start', onDragStart)
      .on('drag', onDrag)
      .on('end', onDragEnd),
    );

  nodeEnter.append('rect');
  nodeEnter.append('g').attr('class', 'node-content');

  // Update all nodes
  const nodeUpdate = nodeEnter.merge(nodeSelection);

  nodeUpdate
    .classed('selected', (d: any) => d.data.id === selectedNodeId)
    .classed('editing', (d: any) => d.data.id === editingNodeId)
    .transition()
    .duration(300)
    .attr('transform', (d: any) => `translate(${d.x},${d.y})`);

  // Update each node's content
  nodeUpdate.each(function(d: any) {
    const node = d3.select(this);
    const nodeData = d.data as MindmapNode;
    const hasImage = !!nodeData.image;

    // Calculate dimensions - support multi-line text and inline images
    const lines = nodeData.text.split('\n');
    const lineHeight = 18; // approximate line height
    
    // Check for inline images in markdown text
    const inlineImageMatches = nodeData.text.match(/!\[.*?\]\(.*?\)/g) || [];
    const hasInlineImages = inlineImageMatches.length > 0;
    const imageHeight = 52; // max-height 48px + small margin
    const imageWidth = 124; // max-width 120px + small margin
    
    // Calculate text dimensions excluding image markdown
    let textLineCount = 0;
    let maxTextWidth = 0;
    for (const line of lines) {
      const strippedLine = stripMarkdown(line).trim();
      if (strippedLine) {
        textLineCount++;
        maxTextWidth = Math.max(maxTextWidth, measureText(strippedLine));
      }
    }
    
    // Width: max of text width or image width (if has images)
    const textWidth = hasInlineImages 
      ? Math.max(maxTextWidth, imageWidth)
      : Math.max(maxTextWidth, 20); // min width for empty text
    
    const contentWidth = hasImage ? textWidth + NODE_IMAGE_SIZE + 8 : textWidth;
    const nodeWidth = Math.max(NODE_MIN_WIDTH, contentWidth + NODE_PADDING.x * 2);
    
    // Height: text lines + image heights
    const textHeight = textLineCount * lineHeight;
    const totalImageHeight = inlineImageMatches.length * imageHeight;
    const contentHeight = textHeight + totalImageHeight;
    
    const nodeHeight = Math.max(NODE_HEIGHT, contentHeight + NODE_PADDING.y * 2, 
      hasImage ? NODE_IMAGE_SIZE + NODE_PADDING.y * 2 : NODE_HEIGHT);

    // Update rectangle
    node.select('rect')
      .attr('width', nodeWidth)
      .attr('height', nodeHeight)
      .attr('x', -nodeWidth / 2)
      .attr('y', -nodeHeight / 2);

    // Update content group
    const contentGroup = node.select('.node-content');
    contentGroup.selectAll('*').remove();

    // Add image if present
    if (hasImage) {
      const imageUri = resolveImagePath(nodeData.image!);
      if (imageUri) {
        contentGroup.append('image')
          .attr('href', imageUri)
          .attr('x', -nodeWidth / 2 + NODE_PADDING.x)
          .attr('y', -NODE_IMAGE_SIZE / 2)
          .attr('width', NODE_IMAGE_SIZE)
          .attr('height', NODE_IMAGE_SIZE)
          .attr('preserveAspectRatio', 'xMidYMid meet');
      }
    }

    // Add text with markdown support using foreignObject
    const textX = hasImage ? -nodeWidth / 2 + NODE_PADDING.x + NODE_IMAGE_SIZE + 8 : 0;
    const textWidth2 = hasImage ? textWidth : nodeWidth - NODE_PADDING.x * 2;

    const fo = contentGroup.append('foreignObject')
      .attr('x', hasImage ? textX : -textWidth2 / 2)
      .attr('y', -nodeHeight / 2)
      .attr('width', textWidth2)
      .attr('height', nodeHeight);

    const div = fo.append('xhtml:div')
      .attr('class', 'node-text')
      .attr('data-node-id', nodeData.id)
      .style('width', '100%')
      .style('height', '100%')
      .style('display', 'flex')
      .style('flex-direction', hasInlineImages ? 'column' : 'row')
      .style('align-items', 'center')
      .style('justify-content', hasImage ? 'flex-start' : 'center')
      .style('overflow', 'hidden')
      .style('white-space', 'pre-wrap')
      .style('word-break', 'break-word')
      .style('text-align', 'center')
      .style('color', nodeData.id === selectedNodeId ? 'var(--node-selected-fg)' : 'var(--node-fg)')
      .style('font-size', '13px')
      .style('line-height', '18px')
      .style('outline', 'none')
      .style('cursor', 'default');

    // If editing this node, make it contenteditable
    if (nodeData.id === editingNodeId) {
      div
        .attr('contenteditable', 'true')
        .style('cursor', 'text')
        .style('overflow', 'visible')
        .style('white-space', 'pre-wrap')
        .style('min-width', '60px')
        .text(nodeData.text);
      
      // Focus and select after render
      setTimeout(() => {
        const el = div.node() as HTMLElement;
        if (el) {
          el.focus();
          // Select all text
          const range = document.createRange();
          range.selectNodeContents(el);
          const sel = window.getSelection();
          sel?.removeAllRanges();
          sel?.addRange(range);
        }
      }, 0);
    } else {
      // Render markdown for display (convert newlines to <br>)
      const htmlContent = renderMarkdown(nodeData.text);
      div.html(htmlContent);
    }
  });

  // Exit old nodes
  nodeSelection.exit().remove();
}

/**
 * Strip markdown for text measurement (removes images entirely)
 */
function stripMarkdown(text: string): string {
  return text
    .replace(/!\[.*?\]\(.*?\)/g, '') // remove images completely
    .replace(/\*\*(.*?)\*\*/g, '$1')  // bold
    .replace(/\*(.*?)\*/g, '$1')       // italic
    .replace(/__(.*?)__/g, '$1')       // bold
    .replace(/_(.*?)_/g, '$1')         // italic
    .replace(/`(.*?)`/g, '$1')         // code
    .replace(/\[(.*?)\]\(.*?\)/g, '$1'); // links
}

/**
 * Render markdown to HTML (simplified inline markdown)
 */
function renderMarkdown(text: string): string {
  // Simple markdown rendering (don't use marked for better control over images)
  let html = escapeHtml(text);
  html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/\*(.*?)\*/g, '<em>$1</em>');
  html = html.replace(/__(.*?)__/g, '<strong>$1</strong>');
  html = html.replace(/_(.*?)_/g, '<em>$1</em>');
  html = html.replace(/`(.*?)`/g, '<code>$1</code>');
  
  // Handle images: ![alt](path) - wrap in block div for proper sizing
  html = html.replace(/!\[(.*?)\]\((.*?)\)/g, (match, alt, src) => {
    const resolvedSrc = resolveImagePath(src);
    if (resolvedSrc) {
      return `<span class="inline-image-wrapper"><img src="${resolvedSrc}" alt="${alt}" class="inline-image" onload="window.dispatchEvent(new CustomEvent('imageLoaded'))" /></span>`;
    }
    // Request resolution for this image
    requestImageResolution(src);
    return `<span class="image-placeholder" data-src="${src}">[Image: ${alt || src}]</span>`;
  });
  
  // Handle links: [text](url)
  html = html.replace(/\[(.*?)\]\((.*?)\)/g, '<a href="$2">$1</a>');
  
  // Convert newlines to <br> for display
  html = html.replace(/\n/g, '<br>');

  return html;
}

/**
 * Request image path resolution from extension
 */
function requestImageResolution(imagePath: string) {
  if (!imageCache.has(imagePath) && !imagePath.startsWith('data:') && !imagePath.startsWith('http') && _resolveImage) {
    const resolved = _resolveImage(imagePath);
    if (resolved) { imageCache.set(imagePath, resolved); }
  }
}

/**
 * Escape HTML entities
 */
function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

/**
 * Resolve relative image path to webview URI
 */
function resolveImagePath(imagePath: string): string | null {
  if (!imagePath) return null;

  // If already a data URI or absolute URL, return as-is
  if (imagePath.startsWith('data:') || imagePath.startsWith('http://') || imagePath.startsWith('https://')) {
    return imagePath;
  }

  // Check cache
  if (imageCache.has(imagePath)) {
    return imageCache.get(imagePath)!;
  }

  // Use the pane-provided resolver
  if (_resolveImage) {
    const resolved = _resolveImage(imagePath);
    if (resolved) {
      imageCache.set(imagePath, resolved);
      return resolved;
    }
  }

  return null;
}

/**
 * Measure text width
 */
function measureText(text: string): number {
  const canvas = document.createElement('canvas');
  const context = canvas.getContext('2d');
  if (!context) return NODE_MIN_WIDTH;
  context.font = '13px system-ui, sans-serif';
  return context.measureText(text).width;
}

/**
 * Select a node
 */
function selectNode(nodeId: string) {
  selectedNodeId = nodeId;
  render();
}

/**
 * Start editing a node
 */
function startEditing(nodeId: string) {
  const node = nodeMap.get(nodeId);
  if (!node) return;

  editingNodeId = nodeId;
  selectedNodeId = nodeId;

  // Re-render to show contenteditable in the node
  render();
}

/**
 * Finish editing and save
 */
function finishEditing() {
  if (!editingNodeId || !mindmapData) return;

  // Find the contenteditable element in the node
  const editor = document.querySelector(`.node-text[data-node-id="${editingNodeId}"]`) as HTMLElement;

  if (editor) {
    const newText = editor.textContent?.trim() || 'Node';
    updateNodeText(editingNodeId, newText);
  }

  editingNodeId = null;
  render();
}

/**
 * Update node text in the data structure
 */
function updateNodeText(nodeId: string, text: string) {
  if (!mindmapData) return;

  const currentNode = findNode(mindmapData.root, nodeId);
  if (!currentNode || currentNode.text === text) return;

  recordHistorySnapshot(`Edited “${describeNodeTitle(currentNode)}”`);

  const updateRecursive = (node: MindmapNode): boolean => {
    if (node.id === nodeId) {
      node.text = text;
      return true;
    }
    if (node.children) {
      for (const child of node.children) {
        if (updateRecursive(child)) return true;
      }
    }
    return false;
  };

  updateRecursive(mindmapData.root);
  saveDocument();
  updateHistoryControls();
}

/**
 * Create a child node under the selected node
 */
function createChildNode() {
  if (!selectedNodeId || !mindmapData) return;

  const newId = `node-${Date.now()}`;
  const newNode: MindmapNode = {
    id: newId,
    text: 'New Node',
    children: [],
  };

  const addChild = (node: MindmapNode): boolean => {
    if (node.id === selectedNodeId) {
      if (!node.children) node.children = [];
      
      // For root node in horizontal tree, auto-assign side to balance
      if (node.id === 'root' && mindmapData?.layout === 'horizontal-tree') {
        const leftCount = node.children.filter(c => c.side === 'left').length;
        const rightCount = node.children.filter(c => c.side !== 'left').length;
        newNode.side = leftCount <= rightCount ? 'left' : 'right';
      }
      
      node.children.push(newNode);
      return true;
    }
    if (node.children) {
      for (const child of node.children) {
        if (addChild(child)) return true;
      }
    }
    return false;
  };

  const historySnapshot = cloneMindmapHistoryDocument(mindmapData);
  if (!addChild(mindmapData.root)) return;
  pushHistoryDocument(historySnapshot, 'Added child node');
  selectedNodeId = newId;
  saveAndRender();
}

/**
 * Create a sibling node next to the selected node
 */
function createSiblingNode() {
  if (!selectedNodeId || !mindmapData || selectedNodeId === 'root') return;

  const newId = `node-${Date.now()}`;
  const newNode: MindmapNode = {
    id: newId,
    text: 'New Node',
    children: [],
  };

  const addSibling = (node: MindmapNode, parent: MindmapNode | null): boolean => {
    if (node.id === selectedNodeId && parent?.children) {
      const index = parent.children.indexOf(node);
      // Inherit side from sibling for horizontal tree
      if (mindmapData?.layout === 'horizontal-tree' && node.side) {
        newNode.side = node.side;
      }
      parent.children.splice(index + 1, 0, newNode);
      return true;
    }
    if (node.children) {
      for (const child of node.children) {
        if (addSibling(child, node)) return true;
      }
    }
    return false;
  };

  const historySnapshot = cloneMindmapHistoryDocument(mindmapData);
  if (!addSibling(mindmapData.root, null)) return;
  pushHistoryDocument(historySnapshot, 'Added sibling node');
  selectedNodeId = newId;
  saveAndRender();
}

/**
 * Move selected node up one level (reparent to grandparent)
 */
function moveNodeUp() {
  if (!selectedNodeId || !mindmapData || selectedNodeId === 'root') return;

  const moveUp = (node: MindmapNode, parent: MindmapNode | null, grandparent: MindmapNode | null): boolean => {
    if (node.id === selectedNodeId && parent && grandparent?.children) {
      // Remove from parent
      const parentIndex = parent.children?.indexOf(node) ?? -1;
      if (parentIndex >= 0) {
        parent.children?.splice(parentIndex, 1);
      }

      // Add after parent in grandparent's children
      const gpIndex = grandparent.children.indexOf(parent);
      grandparent.children.splice(gpIndex + 1, 0, node);
      return true;
    }
    if (node.children) {
      for (const child of node.children) {
        if (moveUp(child, node, parent)) return true;
      }
    }
    return false;
  };

  const historySnapshot = cloneMindmapHistoryDocument(mindmapData);
  if (!moveUp(mindmapData.root, null, null)) return;
  pushHistoryDocument(historySnapshot, 'Promoted node');
  saveAndRender();
}

/**
 * Delete the selected node
 */
function deleteSelectedNode() {
  if (!selectedNodeId || !mindmapData || selectedNodeId === 'root') return;

  const deleteNode = (node: MindmapNode): boolean => {
    if (node.children) {
      const index = node.children.findIndex(c => c.id === selectedNodeId);
      if (index >= 0) {
        node.children.splice(index, 1);
        return true;
      }
      for (const child of node.children) {
        if (deleteNode(child)) return true;
      }
    }
    return false;
  };

  const nodeToDelete = findNode(mindmapData.root, selectedNodeId);
  if (!nodeToDelete) return;
  const historySnapshot = cloneMindmapHistoryDocument(mindmapData);
  if (!deleteNode(mindmapData.root)) return;
  pushHistoryDocument(historySnapshot, `Deleted “${describeNodeTitle(nodeToDelete)}”`);

  // Also remove any crosslinks referencing this node
  if (mindmapData.links) {
    mindmapData.links = mindmapData.links.filter(
      l => l.from !== selectedNodeId && l.to !== selectedNodeId,
    );
  }

  selectedNodeId = null;
  saveAndRender();
}

/**
 * Navigate between nodes with arrow keys
 */
function navigateNode(direction: 'up' | 'down' | 'left' | 'right') {
  if (!mindmapData) return;

  const currentNode = selectedNodeId ? nodeMap.get(selectedNodeId) : null;

  if (!currentNode) {
    // Select root if nothing selected
    selectedNodeId = mindmapData.root.id;
    render();
    return;
  }

  let targetNode: D3Node | undefined;
  const nodes = Array.from(nodeMap.values());

  const isHorizontal = mindmapData.layout === 'horizontal-tree';

  switch (direction) {
    case 'left':
      if (isHorizontal) {
        // Go to parent
        targetNode = currentNode.parent as D3Node | undefined;
      } else {
        // Find nearest node to the left
        targetNode = findNearestNode(currentNode, nodes, -1, 0);
      }
      break;

    case 'right':
      if (isHorizontal) {
        // Go to first child
        targetNode = currentNode.children?.[0] as D3Node | undefined;
      } else {
        targetNode = findNearestNode(currentNode, nodes, 1, 0);
      }
      break;

    case 'up':
      if (isHorizontal) {
        // Go to previous sibling
        targetNode = findSibling(currentNode, -1);
      } else {
        // Go to parent
        targetNode = currentNode.parent as D3Node | undefined;
      }
      break;

    case 'down':
      if (isHorizontal) {
        // Go to next sibling
        targetNode = findSibling(currentNode, 1);
      } else {
        // Go to first child
        targetNode = currentNode.children?.[0] as D3Node | undefined;
      }
      break;
  }

  if (targetNode) {
    selectedNodeId = targetNode.data.id;
    render();
  }
}

/**
 * Find sibling node
 */
function findSibling(node: D3Node, offset: number): D3Node | undefined {
  if (!node.parent) return undefined;
  const siblings = node.parent.children as D3Node[] | undefined;
  if (!siblings) return undefined;
  const index = siblings.indexOf(node);
  return siblings[index + offset];
}

/**
 * Find nearest node in a direction
 */
function findNearestNode(current: D3Node, nodes: D3Node[], dx: number, dy: number): D3Node | undefined {
  let nearest: D3Node | undefined;
  let minDist = Infinity;

  for (const node of nodes) {
    if (node === current) continue;

    const deltaX = node.x - current.x;
    const deltaY = node.y - current.y;

    // Check if in the right direction
    if (dx !== 0 && Math.sign(deltaX) !== Math.sign(dx)) continue;
    if (dy !== 0 && Math.sign(deltaY) !== Math.sign(dy)) continue;

    const dist = Math.sqrt(deltaX * deltaX + deltaY * deltaY);
    if (dist < minDist) {
      minDist = dist;
      nearest = node;
    }
  }

  return nearest;
}

// Drag state
let dragStartX = 0;
let dragStartY = 0;
let dropTarget: D3Node | null = null;

/**
 * Handle drag start
 */
function onDragStart(event: D3DragEvent, d: D3Node) {
  dragStartX = d.x;
  dragStartY = d.y;
  d3.select(event.sourceEvent.target.closest('.node')).raise();
}

/**
 * Handle drag
 */
function onDrag(event: D3DragEvent, d: D3Node) {
  d.x = event.x;
  d.y = event.y;
  
  const nodeEl = d3.select(event.sourceEvent.target.closest('.node'));
  nodeEl.attr('transform', `translate(${d.x},${d.y})`);

  // Update links in real-time (no transition)
  updateLinksImmediate();
  
  // Visual feedback for side flip in horizontal tree
  if (mindmapData?.layout === 'horizontal-tree' && d.parent && d.parent.data.id === 'root') {
    const wasOnLeft = d.data.side === 'left';
    const nowOnLeft = d.x < 0;
    nodeEl.classed('will-flip-side', wasOnLeft !== nowOnLeft);
  }

  // Check for drop targets
  const nodes = Array.from(nodeMap.values());
  dropTarget = null;

  for (const node of nodes) {
    if (node === d || node.data.id === d.data.id) continue;

    const dist = Math.sqrt(Math.pow(node.x - d.x, 2) + Math.pow(node.y - d.y, 2));
    if (dist < 50) {
      dropTarget = node;
      break;
    }
  }

  // Highlight drop target
  nodesGroup.selectAll('.node')
    .classed('drag-over', (n: D3Node) => n === dropTarget);
}

/**
 * Handle drag end
 */
function onDragEnd(_event: D3DragEvent, d: D3Node) {
  nodesGroup.selectAll('.node').classed('drag-over', false).classed('will-flip-side', false);

  if (dropTarget && mindmapData && d.data.id !== 'root') {
    // Re-parent node
    reparentNode(d.data.id, dropTarget.data.id);
  } else if (Math.abs(d.x - dragStartX) > 5 || Math.abs(d.y - dragStartY) > 5) {
    // Check if node crossed to the other side (for horizontal tree bidirectional layout)
    if (mindmapData?.layout === 'horizontal-tree' && d.parent && d.parent.data.id === 'root') {
      const wasOnLeft = d.data.side === 'left';
      const nowOnLeft = d.x < 0;
      
      if (wasOnLeft !== nowOnLeft) {
        // Node crossed sides - update side property and clear manual position
        const historySnapshot = cloneMindmapHistoryDocument(mindmapData);
        d.data.side = nowOnLeft ? 'left' : 'right';
        delete d.data.position;
        pushHistoryDocument(historySnapshot, `Moved “${describeNodeTitle(d.data)}” to ${nowOnLeft ? 'left' : 'right'} branch`);
        saveAndRender();
        dropTarget = null;
        return;
      }
    }
    
    // Save manual position
    const historySnapshot = cloneMindmapHistoryDocument(mindmapData);
    d.data.position = { x: d.x, y: d.y };
    pushHistoryDocument(historySnapshot, `Moved “${describeNodeTitle(d.data)}”`);
    saveAndRender();
  } else {
    render();
  }

  dropTarget = null;
}

/**
 * Re-parent a node to a new parent
 */
function reparentNode(nodeId: string, newParentId: string) {
  if (!mindmapData || nodeId === 'root' || nodeId === newParentId) return;

  // Check if newParent is a descendant of node (would create cycle)
  const isDescendant = (node: MindmapNode, targetId: string): boolean => {
    if (node.id === targetId) return true;
    return node.children?.some(c => isDescendant(c, targetId)) ?? false;
  };

  const findNode = (node: MindmapNode): MindmapNode | undefined => {
    if (node.id === nodeId) return node;
    for (const child of node.children ?? []) {
      const found = findNode(child);
      if (found) return found;
    }
    return undefined;
  };

  const nodeToMove = findNode(mindmapData.root);
  if (!nodeToMove || isDescendant(nodeToMove, newParentId)) return;

  // Remove from current parent
  const removeFromParent = (node: MindmapNode): boolean => {
    if (node.children) {
      const index = node.children.findIndex(c => c.id === nodeId);
      if (index >= 0) {
        node.children.splice(index, 1);
        return true;
      }
      for (const child of node.children) {
        if (removeFromParent(child)) return true;
      }
    }
    return false;
  };

  const historySnapshot = cloneMindmapHistoryDocument(mindmapData);
  if (!removeFromParent(mindmapData.root)) return;

  // Clear manual position since it's being reparented
  delete nodeToMove.position;

  // Add to new parent
  const addToParent = (node: MindmapNode): boolean => {
    if (node.id === newParentId) {
      if (!node.children) node.children = [];
      node.children.push(nodeToMove);
      return true;
    }
    return node.children?.some(c => addToParent(c)) ?? false;
  };

  if (!addToParent(mindmapData.root)) return;
  pushHistoryDocument(historySnapshot, `Moved “${describeNodeTitle(nodeToMove)}” under “${describeNodeTitle(findNode(mindmapData.root, newParentId) || undefined)}”`);
  saveAndRender();
}

/**
 * Fit the view to show all nodes
 */
function fitToView() {
  if (!mindmapData || nodeMap.size === 0) return;

  const nodes = Array.from(nodeMap.values());
  let minX = Infinity, maxX = -Infinity;
  let minY = Infinity, maxY = -Infinity;

  for (const node of nodes) {
    minX = Math.min(minX, node.x - NODE_MIN_WIDTH);
    maxX = Math.max(maxX, node.x + NODE_MIN_WIDTH);
    minY = Math.min(minY, node.y - NODE_HEIGHT);
    maxY = Math.max(maxY, node.y + NODE_HEIGHT);
  }

  const svgNode = svg.node();
  if (!svgNode) return;
  const svgRect = svgNode.getBoundingClientRect();
  if (!svgRect.width || !svgRect.height) return; // not visible yet
  const width = maxX - minX;
  const height = maxY - minY;
  const scale = Math.min(
    0.9 * svgRect.width / width,
    0.9 * svgRect.height / height,
    1.5,
  );
  const centerX = (minX + maxX) / 2;
  const centerY = (minY + maxY) / 2;

  svg.transition()
    .duration(500)
    .call(
      zoom.transform,
      d3.zoomIdentity
        .translate(svgRect.width / 2, svgRect.height / 2)
        .scale(scale)
        .translate(-centerX, -centerY),
    );
}

/**
 * Save document and re-render
 */
function saveAndRender() {
  // Clear dimensions cache since content may have changed
  clearDimensionsCache();
  saveDocument();
  render();
}

/**
 * Save document back to extension
 */
function saveDocument() {
  if (!mindmapData) return;

  try {
    const yaml = jsyaml.dump(mindmapData, {
      indent: 2,
      lineWidth: -1,
      noRefs: true,
    });
    _onEdit?.(yaml);
  } catch (error) {
    console.error('[mindmap] Failed to save:', error);
  }
}

// ── Event listeners registered with AbortController ─────────────
// (called from init(), cleaned up via _abort.abort() in destroy)

function setupGlobalListeners() {
  const signal = _abort?.signal;

  // Click on background to deselect and focus container for keyboard input
  document.addEventListener('click', (e) => {
    const container = document.getElementById('mindmap-container');
    if (container?.contains(e.target as Node)) {
      container.focus();
    }
    if (e.target === svg?.node() || (e.target as Element)?.id === 'mindmap-container') {
      finishEditing();
      selectedNodeId = null;
      render();
    }
  }, { signal });

  // Re-render when inline images load to adjust node sizes
  let imageLoadTimeout: number | null = null;
  window.addEventListener('imageLoaded', () => {
    if (imageLoadTimeout) clearTimeout(imageLoadTimeout);
    imageLoadTimeout = window.setTimeout(() => { render(); imageLoadTimeout = null; }, 100);
  }, { signal } as any);

  // Handle link clicks inside the mindmap only — open in external browser
  document.addEventListener('click', (e) => {
    const container = document.getElementById('mindmap-container');
    const target = e.target as HTMLElement | null;
    const anchor = target?.closest?.('a') as HTMLAnchorElement | null;
    if (!container || !anchor || !container.contains(anchor)) return;
    const href = anchor.getAttribute('href');
    if (!href) return;
    e.preventDefault();
    e.stopPropagation();
    window.open(href, '_blank', 'noopener,noreferrer');
  }, { signal });
}

// ── Public API ──────────────────────────────────────────────────

(window as any).__mindmapEditor = {
  /** Mount the mindmap editor. Container must already have an <svg id="mindmap-svg">. */
  mount(opts: MindmapMountOpts) {
    _onEdit = opts.onEdit;
    _resolveImage = opts.resolveImagePath;
    _abort = new AbortController();

    init();               // sets up D3, toolbar, keyboard, context menu
    setupGlobalListeners();

    // Load initial content
    loadContent(opts.content);

    // Apply theme
    const container = document.getElementById('mindmap-container');
    container?.classList.toggle('light', !opts.isDark);
  },

  /** Push new content from the host (external file reload). */
  update(content: string) {
    loadContent(content);
  },

  /** Set light/dark theme. */
  setTheme(isDark: boolean) {
    const container = document.getElementById('mindmap-container');
    container?.classList.toggle('light', !isDark);
  },

  /** Tear down everything — abort listeners, clear D3 state. */
  destroy() {
    _abort?.abort();
    _abort = null;

    // Remove D3 event handlers
    svg?.on('.zoom', null);
    svg?.on('contextmenu', null);
    svg?.on('mousemove.linkCreate', null);
    svg?.on('mouseup.linkCreate', null);
    svg?.on('contextmenu.linkCreate', null);

    // Clear state
    svg = null as any;
    zoomContainer = null;
    linksGroup = null;
    crosslinksGroup = null;
    nodesGroup = null;
    zoom = null;
    mindmapData = null;
    selectedNodeId = null;
    editingNodeId = null;
    nodeMap.clear();
    imageCache.clear();
    isCreatingLink = false;
    linkSourceNodeId = null;
    linkPreviewLine = null;
    clipboardNode = null;
    undoStack = [];
    redoStack = [];
    _onEdit = null;
    _resolveImage = null;
  },
};
