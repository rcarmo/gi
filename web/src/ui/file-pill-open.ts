export type FilePillOpenContext = {
  path: string;
  mode: "edit";
};

export type FilePillOpenResult =
  | { kind: "ignore" }
  | { kind: "open"; path: string }
  | { kind: "toast"; title: string; detail: string; level: "warning" };

export function resolveFilePillOpenAction(
  rawPath: unknown,
  options: {
    editorOpen: boolean;
    autoOpenEditor?: boolean;
    resolvePane: (context: FilePillOpenContext) => unknown;
  }
): FilePillOpenResult {
  if (typeof rawPath !== "string") return { kind: "ignore" };

  const path = rawPath.trim();
  if (!path) {
    return {
      kind: "toast",
      title: "No file selected",
      detail: "Use a valid file path from a file pill.",
      level: "warning",
    };
  }

  if (!options.editorOpen && !options.autoOpenEditor) {
    return {
      kind: "toast",
      title: "Editor pane is not open",
      detail: "Open the editor pane to open files from pills.",
      level: "warning",
    };
  }

  const protocolMatch = /^[a-z][a-z0-9+.-]*:/i.test(path);
  if (protocolMatch) {
    return {
      kind: "toast",
      title: "Cannot open external path from file pill",
      detail: "Use an in-workspace file path.",
      level: "warning",
    };
  }

  try {
    const editorExt = options.resolvePane({ path, mode: "edit" });
    if (!editorExt) {
      return {
        kind: "toast",
        title: "No editor available",
        detail: `No editor can open: ${path}`,
        level: "warning",
      };
    }
  } catch {
    return {
      kind: "toast",
      title: "No editor available",
      detail: `No editor can open: ${path}`,
      level: "warning",
    };
  }

  return { kind: "open", path };
}
