export interface BtwSessionLike {
  question?: string;
  answer?: string;
  thinking?: string;
  error?: string | null;
  model?: string | null;
  status?: string;
}

interface RefBox<T> {
  current: T;
}

type StateSetter<T> = (next: T | ((prev: T) => T)) => void;

type ToastFn = (title: string, detail?: string | null, kind?: string, durationMs?: number) => void;

/** Abort any active BTW stream and close the BTW panel state. */
export function closeBtwPanelSession(options: {
  btwAbortRef: RefBox<AbortController | null>;
  setBtwSession: StateSetter<BtwSessionLike | null>;
}): void {
  const { btwAbortRef, setBtwSession } = options;
  if (btwAbortRef.current) {
    btwAbortRef.current.abort();
    btwAbortRef.current = null;
  }
  setBtwSession(null);
}

export interface RunBtwPromptSessionOptions {
  question: unknown;
  currentChatJid: string;
  streamSidePrompt: (prompt: string, options: Record<string, unknown>) => Promise<any>;
  resolveBtwChatJid: (chatJid: string) => string;
  showIntentToast: ToastFn;
  btwAbortRef: RefBox<AbortController | null>;
  setBtwSession: StateSetter<BtwSessionLike | null>;
}

/** Execute BTW side-prompt streaming flow and update session state progressively. */
export async function runBtwPromptSession(options: RunBtwPromptSessionOptions): Promise<boolean> {
  const {
    question,
    currentChatJid,
    streamSidePrompt,
    resolveBtwChatJid,
    showIntentToast,
    btwAbortRef,
    setBtwSession,
  } = options;

  const trimmed = String(question || '').trim();
  if (!trimmed) {
    showIntentToast('BTW needs a question', 'Usage: /btw <question>', 'warning');
    return true;
  }

  if (btwAbortRef.current) {
    btwAbortRef.current.abort();
  }

  const controller = new AbortController();
  btwAbortRef.current = controller;

  setBtwSession({
    question: trimmed,
    answer: '',
    thinking: '',
    error: null,
    model: null,
    status: 'running',
  });

  try {
    const finalResult = await streamSidePrompt(trimmed, {
      signal: controller.signal,
      chatJid: resolveBtwChatJid(currentChatJid),
      systemPrompt: 'Answer the user briefly and directly. This is a side conversation that should not affect the main chat until explicitly injected.',
      onEvent: (eventType: string) => {
        if (eventType === 'side_prompt_start') {
          setBtwSession((prev) => (prev ? { ...prev, status: 'running' } : prev));
        }
      },
      onThinkingDelta: (delta: string) => {
        setBtwSession((prev) => (prev ? { ...prev, thinking: `${prev.thinking || ''}${delta || ''}` } : prev));
      },
      onTextDelta: (delta: string) => {
        setBtwSession((prev) => (prev ? { ...prev, answer: `${prev.answer || ''}${delta || ''}` } : prev));
      },
    });

    if (btwAbortRef.current !== controller) {
      return true;
    }

    setBtwSession((prev) => (prev
      ? {
        ...prev,
        answer: finalResult?.result || prev.answer || '',
        thinking: finalResult?.thinking || prev.thinking || '',
        model: finalResult?.model || null,
        status: 'success',
        error: null,
      }
      : prev));
  } catch (error: any) {
    if (controller.signal.aborted) return true;
    setBtwSession((prev) => (prev
      ? {
        ...prev,
        status: 'error',
        error: error?.payload?.error || error?.message || 'BTW request failed.',
      }
      : prev));
  } finally {
    if (btwAbortRef.current === controller) {
      btwAbortRef.current = null;
    }
  }

  return true;
}

export interface HandleBtwInterceptOptions {
  content: unknown;
  parseBtwCommand: (content: unknown) => { type: string; question?: string } | null;
  closeBtwPanel: () => void;
  runBtwPrompt: (question: unknown) => Promise<boolean>;
  showIntentToast: ToastFn;
}

/** Parse `/btw`-style commands and run the corresponding panel action. */
export async function handleBtwInterceptCommand(options: HandleBtwInterceptOptions): Promise<boolean> {
  const {
    content,
    parseBtwCommand,
    closeBtwPanel,
    runBtwPrompt,
    showIntentToast,
  } = options;

  const parsed = parseBtwCommand(content);
  if (!parsed) return false;

  if (parsed.type === 'help') {
    showIntentToast('BTW usage', 'Use /btw <question> to open a side conversation.', 'info', 4000);
    return true;
  }

  if (parsed.type === 'clear') {
    closeBtwPanel();
    showIntentToast('BTW cleared', 'Closed the side conversation panel.', 'info');
    return true;
  }

  if (parsed.type === 'ask') {
    await runBtwPrompt(parsed.question);
    return true;
  }

  return false;
}

export interface InjectBtwSessionOptions {
  btwSession: BtwSessionLike | null;
  buildBtwInjectionText: (session: BtwSessionLike | null) => string;
  isComposeBoxAgentActive: boolean;
  currentChatJid: string;
  sendAgentMessage: (
    agentId: string,
    content: string,
    threadId: string | null,
    attachments: any[],
    queueMode: string | null,
    chatJid: string,
  ) => Promise<any>;
  handleMessageResponse: (response: any) => void;
  showIntentToast: ToastFn;
}

/** Inject BTW side-session summary into the main chat (or queue when agent is busy). */
export async function injectBtwSession(options: InjectBtwSessionOptions): Promise<boolean> {
  const {
    btwSession,
    buildBtwInjectionText,
    isComposeBoxAgentActive,
    currentChatJid,
    sendAgentMessage,
    handleMessageResponse,
    showIntentToast,
  } = options;

  const content = buildBtwInjectionText(btwSession);
  if (!content) return false;

  try {
    const response = await sendAgentMessage('default', content, null, [], isComposeBoxAgentActive ? 'queue' : null, currentChatJid);
    handleMessageResponse(response);
    showIntentToast(
      response?.queued === 'followup' ? 'BTW queued' : 'BTW injected',
      response?.queued === 'followup'
        ? 'The BTW summary was queued as a follow-up because the agent is busy.'
        : 'The BTW summary was sent to the main chat.',
      'info',
      3500,
    );
    return true;
  } catch (error: any) {
    showIntentToast('BTW inject failed', error?.message || 'Could not inject BTW answer into chat.', 'warning');
    return false;
  }
}
