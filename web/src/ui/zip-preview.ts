const EOCD_SIGNATURE = 0x06054b50;
const CENTRAL_DIRECTORY_SIGNATURE = 0x02014b50;
const ZIP64_EOCD_LOCATOR_SIGNATURE = 0x07064b50;
const MAX_EOCD_SEARCH_BYTES = 22 + 0xffff;
const UTF8_FLAG = 0x0800;

const utf8Decoder = new TextDecoder("utf-8", { fatal: false });

export type ZipPreviewEntry = {
  path: string;
  isDirectory: boolean;
  compressedSize: number;
  uncompressedSize: number;
  compressionMethod: number;
  comment: string;
};

export type ZipPreviewSummary = {
  fileCount: number;
  directoryCount: number;
  totalEntries: number;
  compressedBytes: number;
  uncompressedBytes: number;
};

export type ZipPreviewData = {
  entries: ZipPreviewEntry[];
  summary: ZipPreviewSummary;
};

function readUint16(bytes: Uint8Array, offset: number): number {
  return bytes[offset] | (bytes[offset + 1] << 8);
}

function readUint32(bytes: Uint8Array, offset: number): number {
  return (
    bytes[offset]
    | (bytes[offset + 1] << 8)
    | (bytes[offset + 2] << 16)
    | (bytes[offset + 3] << 24)
  ) >>> 0;
}

function decodeBytes(bytes: Uint8Array, offset: number, length: number): string {
  return utf8Decoder.decode(bytes.subarray(offset, offset + length));
}

function findEndOfCentralDirectory(bytes: Uint8Array): number {
  const start = Math.max(0, bytes.length - MAX_EOCD_SEARCH_BYTES);
  for (let offset = bytes.length - 22; offset >= start; offset -= 1) {
    if (readUint32(bytes, offset) === EOCD_SIGNATURE) return offset;
  }
  return -1;
}

function detectZip64(bytes: Uint8Array, eocdOffset: number): boolean {
  const start = Math.max(0, eocdOffset - 20);
  for (let offset = start; offset <= eocdOffset - 4; offset += 1) {
    if (readUint32(bytes, offset) === ZIP64_EOCD_LOCATOR_SIGNATURE) return true;
  }
  return false;
}

function inferDirectoryNames(entries: ZipPreviewEntry[]): Set<string> {
  const names = new Set<string>();
  for (const entry of entries) {
    const normalized = entry.path.replace(/\/+/g, "/");
    if (!normalized) continue;
    if (entry.isDirectory) {
      names.add(normalized.endsWith("/") ? normalized.slice(0, -1) : normalized);
      continue;
    }
    const parts = normalized.split("/").filter(Boolean);
    if (parts.length <= 1) continue;
    let prefix = "";
    for (let index = 0; index < parts.length - 1; index += 1) {
      prefix = prefix ? `${prefix}/${parts[index]}` : parts[index];
      names.add(prefix);
    }
  }
  return names;
}

function normalizePath(value: string): string {
  return String(value || "").replace(/\\/g, "/").trim();
}

export function getCompressionMethodLabel(method: number): string {
  switch (Number(method)) {
    case 0:
      return "Stored";
    case 8:
      return "Deflate";
    case 9:
      return "Deflate64";
    case 12:
      return "BZIP2";
    case 14:
      return "LZMA";
    case 93:
      return "Zstandard";
    default:
      return `Method ${method}`;
  }
}

export function parseZipPreview(bytesLike: Uint8Array | ArrayBuffer | ArrayBufferView): ZipPreviewData {
  const bytes = bytesLike instanceof Uint8Array
    ? bytesLike
    : bytesLike instanceof ArrayBuffer
      ? new Uint8Array(bytesLike)
      : new Uint8Array(bytesLike.buffer, bytesLike.byteOffset, bytesLike.byteLength);

  if (bytes.length < 22) {
    throw new Error("ZIP archive is too small to contain a valid directory.");
  }

  const eocdOffset = findEndOfCentralDirectory(bytes);
  if (eocdOffset < 0) {
    throw new Error("ZIP archive directory could not be found.");
  }

  if (detectZip64(bytes, eocdOffset)) {
    throw new Error("ZIP64 archives are not previewable yet.");
  }

  const totalEntries = readUint16(bytes, eocdOffset + 10);
  const centralDirectorySize = readUint32(bytes, eocdOffset + 12);
  const centralDirectoryOffset = readUint32(bytes, eocdOffset + 16);

  if (centralDirectoryOffset + centralDirectorySize > bytes.length) {
    throw new Error("ZIP archive directory is truncated.");
  }

  const entries: ZipPreviewEntry[] = [];
  let cursor = centralDirectoryOffset;
  const endOffset = centralDirectoryOffset + centralDirectorySize;

  while (cursor < endOffset) {
    if (cursor + 46 > bytes.length) {
      throw new Error("ZIP archive directory entry is truncated.");
    }
    if (readUint32(bytes, cursor) !== CENTRAL_DIRECTORY_SIGNATURE) {
      throw new Error("ZIP archive directory contains an unexpected record.");
    }

    const generalPurposeFlags = readUint16(bytes, cursor + 8);
    const compressionMethod = readUint16(bytes, cursor + 10);
    const compressedSize = readUint32(bytes, cursor + 20);
    const uncompressedSize = readUint32(bytes, cursor + 24);
    const fileNameLength = readUint16(bytes, cursor + 28);
    const extraLength = readUint16(bytes, cursor + 30);
    const commentLength = readUint16(bytes, cursor + 32);
    const nameOffset = cursor + 46;
    const commentOffset = nameOffset + fileNameLength + extraLength;
    const nextOffset = commentOffset + commentLength;

    if (nextOffset > bytes.length) {
      throw new Error("ZIP archive directory entry metadata is truncated.");
    }

    const supportsUtf8 = (generalPurposeFlags & UTF8_FLAG) === UTF8_FLAG;
    const path = normalizePath(decodeBytes(bytes, nameOffset, fileNameLength));
    const comment = decodeBytes(bytes, commentOffset, commentLength);
    const isDirectory = path.endsWith("/");

    if (path) {
      entries.push({
        path,
        isDirectory,
        compressedSize,
        uncompressedSize,
        compressionMethod,
        comment: supportsUtf8 ? comment : comment,
      });
    }

    cursor = nextOffset;
  }

  entries.sort((left, right) => {
    if (left.isDirectory !== right.isDirectory) return left.isDirectory ? -1 : 1;
    return left.path.localeCompare(right.path);
  });

  const files = entries.filter((entry) => !entry.isDirectory);
  const directories = inferDirectoryNames(entries);

  return {
    entries,
    summary: {
      fileCount: files.length,
      directoryCount: directories.size,
      totalEntries: entries.length,
      compressedBytes: files.reduce((sum, entry) => sum + entry.compressedSize, 0),
      uncompressedBytes: files.reduce((sum, entry) => sum + entry.uncompressedSize, 0),
    },
  };
}

export function formatCompressionRatio(summary: ZipPreviewSummary | null | undefined): string | null {
  if (!summary) return null;
  if (summary.uncompressedBytes <= 0) return null;
  const saved = 1 - (summary.compressedBytes / summary.uncompressedBytes);
  if (!Number.isFinite(saved)) return null;
  return `${Math.round(saved * 100)}% smaller`;
}
