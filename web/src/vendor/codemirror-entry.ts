import { EditorState } from '@codemirror/state';
import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter, highlightSpecialChars, scrollPastEnd, showPanel, Decoration, ViewPlugin, WidgetType, drawSelection } from '@codemirror/view';
import { history, defaultKeymap, historyKeymap, indentWithTab } from '@codemirror/commands';
import { defaultHighlightStyle, StreamLanguage, HighlightStyle, syntaxHighlighting, syntaxTree, indentOnInput, indentUnit } from '@codemirror/language';

export { EditorState, Compartment, RangeSetBuilder, RangeSet, Prec } from '@codemirror/state';
export { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter, highlightSpecialChars, scrollPastEnd, showPanel, Decoration, ViewPlugin, WidgetType } from '@codemirror/view';
export { javascript, javascriptLanguage, jsxLanguage, tsxLanguage, typescriptLanguage } from '@codemirror/lang-javascript';
export { python, pythonLanguage } from '@codemirror/lang-python';
export { markdown, markdownLanguage } from '@codemirror/lang-markdown';
export { go, goLanguage } from '@codemirror/lang-go';
export { json, jsonLanguage } from '@codemirror/lang-json';
export { css, cssLanguage } from '@codemirror/lang-css';
export { html, htmlLanguage } from '@codemirror/lang-html';
export { yaml, yamlLanguage } from '@codemirror/lang-yaml';
export { sql, StandardSQL } from '@codemirror/lang-sql';
export { xml, xmlLanguage } from '@codemirror/lang-xml';
export { StreamLanguage, HighlightStyle, syntaxHighlighting, syntaxTree, indentOnInput, indentUnit } from '@codemirror/language';
export { tags, classHighlighter, highlightTree } from '@lezer/highlight';
export { shell } from '@codemirror/legacy-modes/mode/shell';
export { dockerFile } from '@codemirror/legacy-modes/mode/dockerfile';
export { powerShell } from '@codemirror/legacy-modes/mode/powershell';
export { ruby } from '@codemirror/legacy-modes/mode/ruby';
export { rust } from '@codemirror/legacy-modes/mode/rust';
export { swift } from '@codemirror/legacy-modes/mode/swift';
export { toml } from '@codemirror/legacy-modes/mode/toml';
export { indentWithTab } from '@codemirror/commands';
export { search, openSearchPanel, closeSearchPanel, searchKeymap, highlightSelectionMatches } from '@codemirror/search';
export { autocompletion, completionKeymap, closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete';
export { lintGutter } from '@codemirror/lint';
export { vim } from '@replit/codemirror-vim';
export { indentationMarkers } from '@replit/codemirror-indentation-markers';
export { githubLight, githubDark } from '@uiw/codemirror-theme-github';
export { MergeView } from '@codemirror/merge';

export const minimalSetup = [
  highlightSpecialChars(),
  history(),
  drawSelection(),
  syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
  keymap.of([...defaultKeymap, ...historyKeymap]),
];
