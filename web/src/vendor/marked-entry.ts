import {
  Marked,
  Lexer,
  Parser,
  Renderer,
  TextRenderer,
  Tokenizer,
  getDefaults,
  lexer,
  marked,
  options,
  parse,
  parseInline,
  parser,
  setOptions,
  use,
  walkTokens,
} from "marked";

const markedApi = Object.assign(marked, {
  Marked,
  Lexer,
  Parser,
  Renderer,
  TextRenderer,
  Tokenizer,
  getDefaults,
  lexer,
  marked,
  options,
  parse,
  parseInline,
  parser,
  setOptions,
  use,
  walkTokens,
});

(globalThis as Record<string, unknown>).marked = markedApi;
