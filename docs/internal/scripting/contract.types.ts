export type ScriptToolInput = {
  script?: string;
  path?: string;
  engine?: "js" | "joker";
  session_id?: string;
};

export type ScriptToolOutput = {
  result: string;
  error?: string;
};

export type HTTPResponse = {
  status_code: number;
  status: string;
  headers: Record<string, string[]>;
  body: string;
  url: string;
};

export type SessionInfo = {
  session: Record<string, unknown>;
  config: Record<string, unknown>;
  message_count: number;
  turn_count: number;
  messages: Array<Record<string, unknown>>;
  turns: Array<Record<string, unknown>>;
};

export type FileEntry = {
  name: string;
  isDir: boolean;
  size?: number;
};

export type EventHookSpec = {
  name: string;
  source?: string;
  filter?: Record<string, unknown>;
  arguments?: Record<string, unknown>;
};

export type RawSocketSpec = {
  protocol?: "tcp" | "udp";
  address: string;
  timeout_ms?: number;
  local_addr?: string;
};

export type RawSocketPayload = {
  socket_id: string;
  data?: string;
  max_bytes?: number;
  timeout_ms?: number;
};

export type WebSocketSpec = {
  url: string;
  headers?: Record<string, string[]>;
  subprotocol?: string;
  timeout_ms?: number;
};

export type HTTPRequestSpec = {
  method?: string;
  url: string;
  headers?: Record<string, string[]>;
  body?: string;
  timeout_ms?: number;
  retry?: number;
  allow_redirects?: boolean;
  skip_tls?: boolean;
};

export type GiJsBridge = {
  // Session/config
  sessionId: string;
  config: Record<string, unknown>;
  runtimeConfig: Record<string, unknown>;
  sessionState: Record<string, unknown>;
  getSessionState(): Record<string, unknown>;
  setSessionState(patch: Record<string, unknown>): void;

  getSessionInfo(): SessionInfo;
  getRuntimeConfig(): Record<string, unknown>;
  listTurns(limit?: number): Array<Record<string, unknown>>;
  listMessages(limit?: number): Array<Record<string, unknown>>;

  // Files
  readFile(path: string): string;
  writeFile(path: string, content: string): void;
  listDir(path: string): FileEntry[];

  // Events
  registerEventHook(spec: EventHookSpec): void;
  emitEvent(name: string, payload?: Record<string, unknown>): void;
  clearEventHooks(): void;

  // Raw sockets
  net: {
    openRawSocket(spec: RawSocketSpec): string;
    writeRawSocket(payload: RawSocketPayload): number;
    readRawSocket(payload: RawSocketPayload): string;
    closeRawSocket(socketId: string): void;
  };

  // Web sockets
  websocket: {
    open(spec: WebSocketSpec): string;
    write(socketId: string, payload: string): void;
    read(socketId: string, timeout_ms?: number): string;
    close(socketId: string): void;
  };

  // HTTP
  http: {
    request(req: HTTPRequestSpec): HTTPResponse;
  };

  // Logging
  log(level: string, message: string): void;
};

export type GiJokerBridge = {
  // Session/state/config helpers
  "gi-get-session-state"(): Record<string, unknown>;
  "gi-set-session-state!"(patch: Record<string, unknown>): void;
  "gi-get-session-info"(): SessionInfo;
  "gi-get-runtime-config"(): Record<string, unknown>;
  "gi-list-turns"(): Array<Record<string, unknown>>;
  "gi-list-messages"(spec?: { limit?: number }): Array<Record<string, unknown>>;

  // Events
  "gi-register-event-hook"(spec: EventHookSpec): void;
  "gi-emit-event"(name: string, payload: Record<string, unknown>): void;
  "gi-clear-event-hooks"(): void;

  // Raw sockets
  "gi-open-raw-socket"(spec: RawSocketSpec): string;
  "gi-write-raw-socket"(payload: RawSocketPayload): number;
  "gi-read-raw-socket"(payload: RawSocketPayload): string;
  "gi-close-raw-socket"(socketId: string): void;

  // Websocket
  "gi-open-websocket"(spec: WebSocketSpec): string;
  "gi-write-websocket"(socketId: string, payload: string): void;
  "gi-read-websocket"(socketId: string, timeoutMs: number): string;
  "gi-close-websocket"(socketId: string): void;

  // HTTP
  "gi-http-request"(req: HTTPRequestSpec): HTTPResponse;
};
