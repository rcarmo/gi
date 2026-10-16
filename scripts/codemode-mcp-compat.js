// Gi compatibility helpers retained from pi-codemode 1.0.1 (MIT).
const MCP_TYPESCRIPT_PREAMBLE = `type Role = "user" | "assistant";
type MetaObject = Record<string, unknown>;
type Annotations = {
  audience?: Role[];
  priority?: number;
  lastModified?: string;
};
type Icon = {
  src: string;
  mimeType?: string;
  sizes?: string[];
  theme?: "light" | "dark";
};
type TextResourceContents = {
  uri: string;
  mimeType?: string;
  _meta?: MetaObject;
  text: string;
};
type BlobResourceContents = {
  uri: string;
  mimeType?: string;
  _meta?: MetaObject;
  blob: string;
};
type TextContent = {
  type: "text";
  text: string;
  annotations?: Annotations;
  _meta?: MetaObject;
};
type ImageContent = {
  type: "image";
  data: string;
  mimeType: string;
  annotations?: Annotations;
  _meta?: MetaObject;
};
type AudioContent = {
  type: "audio";
  data: string;
  mimeType: string;
  annotations?: Annotations;
  _meta?: MetaObject;
};
type ResourceLink = {
  icons?: Icon[];
  name: string;
  title?: string;
  uri: string;
  description?: string;
  mimeType?: string;
  annotations?: Annotations;
  size?: number;
  _meta?: MetaObject;
  type: "resource_link";
};
type EmbeddedResource = {
  type: "resource";
  resource: TextResourceContents | BlobResourceContents;
  annotations?: Annotations;
  _meta?: MetaObject;
};
type ContentBlock =
  | TextContent
  | ImageContent
  | AudioContent
  | ResourceLink
  | EmbeddedResource;
type CallToolResult<TStructured = { [key: string]: unknown }> = {
  _meta?: MetaObject;
  content: ContentBlock[];
  isError?: boolean;
  structuredContent?: TStructured;
  [key: string]: unknown;
};`;
/**
 * Render TypeScript declarations for the script-visible API. Tools become members of
 * `declare const tools`, globals become `declare function` statements, and `ns.member` globals
 * members of `declare const ns`. Descriptions become doc comments; schemas become types.
 */
function renderDeclarations(options) {
    const sections = [];
    const tools = options.tools ?? [];
    if (tools.length > 0) {
        const members = tools.map((tool) => `${docComment(tool.description, INDENT)}${INDENT}${renderToolSignature(tool)}`);
        sections.push(`declare const tools: {\n${members.join("\n")}\n};`);
    }
    const namespaces = new Map();
    for (const global of options.globals ?? []) {
        const dot = global.name.indexOf(".");
        if (dot === -1) {
            sections.push(renderGlobal(`declare function ${global.name}`, global, ""));
            continue;
        }
        const namespace = global.name.slice(0, dot);
        const members = namespaces.get(namespace) ?? [];
        if (members.length === 0)
            namespaces.set(namespace, members);
        members.push(renderGlobal(global.name.slice(dot + 1), global, INDENT));
    }
    for (const [namespace, members] of namespaces) {
        sections.push(`declare const ${namespace}: {\n${members.join("\n")}\n};`);
    }
    return sections.join("\n\n");
}
/**
 * One tool as a member of the `tools` object: `name(args: T): Promise<R>;` with the
 * name as the identifier scripts use. Input types longer than `inputMaxChars` render as `unknown`.
 * Tools whose output schema is an MCP `CallToolResult` render as `Promise<CallToolResult<T>>`,
 * which needs {@link MCP_TYPESCRIPT_PREAMBLE}.
 */
function renderToolSignature(tool, options = {}) {
    const input = tool.inputSchema === undefined
        ? "unknown"
        : schemaToType(tool.inputSchema, { maxChars: options.inputMaxChars ?? DEFAULT_INPUT_SCHEMA_MAX_CHARS });
    return `${toCodemodeIdentifier(tool.name)}(args: ${input}): Promise<${renderToolOutputType(tool.outputSchema)}>;`;
}
/**
 * A tool's sample: the description followed by the tool's declaration. Used for tool
 * listings and `ALL_TOOLS` entries.
 */
function renderToolSample(tool, options = {}) {
    const declaration = `declare const tools: { ${renderToolSignature(tool, options)} };`;
    return `${tool.description?.trim() ?? ""}\n\ncodemode tool declaration:\n\`\`\`ts\n${declaration}\n\`\`\``;
}
/**
 * The `structuredContent` schema of an MCP `CallToolResult` output schema (detected by a
 * `content` array of objects, boolean `isError`, and object `_meta`), `true` when it declares none, or
 * `undefined` when the schema is not a `CallToolResult`.
 */
function mcpStructuredContentSchema(schema) {
    if (!isObject(schema) || !isObject(schema.properties))
        return undefined;
    const { content, isError, _meta, structuredContent } = schema.properties;
    if (!isObject(content) || content.type !== "array" || !isObject(content.items) || content.items.type !== "object") {
        return undefined;
    }
    if (!isObject(isError) || isError.type !== "boolean" || !isObject(_meta) || _meta.type !== "object")
        return undefined;
    return isObject(structuredContent) || typeof structuredContent === "boolean" ? structuredContent : true;
}
