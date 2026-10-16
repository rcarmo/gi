// Refreshes internal/codemode/vendor from the installed Pi packages:
//   quickjs.wasm  - quickjs-wasi (MIT, Vercel), QuickJS-ng compiled to WASI
//   prelude.js    - Pi's codemode prelude (MIT, @earendil-works/pi-codemode)
// Run: bun scripts/vendor-codemode.mjs
import { copyFileSync, readFileSync, writeFileSync } from 'node:fs';
import { homedir } from 'node:os';

const nm = process.env.PI_NODE_MODULES || `${homedir()}/.bun/install/global/node_modules`;
const out = 'internal/codemode/vendor';
const piPkg = JSON.parse(readFileSync(`${nm}/@earendil-works/pi-codemode/package.json`, 'utf8'));
const qjsPkg = JSON.parse(readFileSync(`${nm}/quickjs-wasi/package.json`, 'utf8'));
const { PRELUDE_SOURCE, MAX_STORE_VALUE_CHARS, MAX_STORE_TOTAL_CHARS } = await import(`${nm}/@earendil-works/pi-codemode/dist/runtime/prelude-source.js`);

copyFileSync(`${nm}/quickjs-wasi/quickjs.wasm`, `${out}/quickjs.wasm`);
copyFileSync(`${nm}/quickjs-wasi/LICENSE`, `${out}/LICENSE.quickjs-wasi`);
writeFileSync(`${out}/prelude.js`, `// Vendored from @earendil-works/pi-codemode ${piPkg.version} (MIT) by scripts/vendor-codemode.mjs; do not edit.\n${PRELUDE_SOURCE}\n`);
writeFileSync(`${out}/VERSIONS`, `pi-codemode ${piPkg.version} (${piPkg.license})\nquickjs-wasi ${qjsPkg.version} (${qjsPkg.license})\nMAX_STORE_VALUE_CHARS ${MAX_STORE_VALUE_CHARS}\nMAX_STORE_TOTAL_CHARS ${MAX_STORE_TOTAL_CHARS}\n`);
writeFileSync(`${out}/LICENSE.pi-codemode`, `MIT License\n\nCopyright (c) the pi authors (https://github.com/earendil-works/pi)\n\nPermission is hereby granted, free of charge, to any person obtaining a copy\nof this software and associated documentation files (the "Software"), to deal\nin the Software without restriction, including without limitation the rights\nto use, copy, modify, merge, publish, distribute, sublicense, and/or sell\ncopies of the Software, and to permit persons to whom the Software is\nfurnished to do so, subject to the following conditions:\n\nThe above copyright notice and this permission notice shall be included in all\ncopies or substantial portions of the Software.\n\nTHE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR\nIMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,\nFITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE\nAUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER\nLIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,\nOUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE\nSOFTWARE.\n`);
// declarations.js + identifier.js as a plain script exposing globalThis.__decl,
// evaluated inside QuickJS to render model-facing TypeScript declarations.
const strip = (src) => src
  .replace(/^import .*$/gm, '')
  .replace(/^export \{[^}]*\};?$/gm, '')
  .replace(/^export (function|const|class) /gm, '$1 ')
  .replace(/^\/\/# sourceMappingURL=.*$/gm, '');
const identifierSrc = strip(readFileSync(`${nm}/@earendil-works/pi-codemode/dist/identifier.js`, 'utf8'));
const mcpCompat = readFileSync(new URL('./codemode-mcp-compat.js', import.meta.url), 'utf8');
let declarationsSrc = strip(readFileSync(`${nm}/@earendil-works/pi-codemode/dist/declarations.js`, 'utf8'));
const outputAnchor = 'function renderToolOutputType(schema) {';
if (declarationsSrc.split(outputAnchor).length !== 2) throw new Error('declaration output anchor changed');
declarationsSrc = declarationsSrc.replace(outputAnchor, outputAnchor + `
    const structured = mcpStructuredContentSchema(schema);
    if (structured !== undefined) {
        const type = structured === true ? "unknown" : schemaToType(structured);
        return type === "unknown" ? "CallToolResult" : \`CallToolResult<\${type}>\`;
    }`);
writeFileSync(`${out}/declarations.js`, `// Vendored from @earendil-works/pi-codemode ${piPkg.version} (MIT): identifier.js + declarations.js\n// as a plain script, by scripts/vendor-codemode.mjs; do not edit.\n${identifierSrc}\n${mcpCompat}\n${declarationsSrc}\nglobalThis.__decl = { renderToolSample, renderDeclarations, renderToolOutputType, mcpStructuredContentSchema, MCP_TYPESCRIPT_PREAMBLE, toCodemodeIdentifier };\n`);
// Model-facing codemode tool texts, from Pi's codemode extension itself: the
// base description (intro and globals, without the models API that gi does
// not have), the code parameter text, and the system-prompt contribution.
const piAgent = `${nm}/@earendil-works/pi-coding-agent`;
const agentVersion = JSON.parse(readFileSync(`${piAgent}/package.json`, 'utf8')).version;
const tool = await import(`${piAgent}/dist/extensions/codemode/tool.js`);
writeFileSync(`${out}/codemode-texts.json`, JSON.stringify({
  baseDescription: tool.createCodemodeDescription([], { models: false }),
  // With the models API; Pi's docs path becomes a placeholder gi fills in.
  baseDescriptionModels: tool.createCodemodeDescription([], { models: true }).split(tool.CODEMODE_DOCS_PATH).join('{{CODEMODE_DOCS}}'),
  codeDescription: tool.codemodeSchema.properties.code.description,
  promptSnippet: tool.codemodeToolSystemPromptContribution.snippet,
  promptGuidelines: tool.codemodeToolSystemPromptContribution.guidelines,
}, null, 2) + '\n');
writeFileSync(`${out}/VERSIONS`, readFileSync(`${out}/VERSIONS`, 'utf8') + `pi-coding-agent ${agentVersion} (codemode texts)\n`);
console.log(`vendored pi-codemode ${piPkg.version}, quickjs-wasi ${qjsPkg.version}`);
