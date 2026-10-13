export interface SSEEvent {
  data: string;
  event: string;
  id?: string;
}

export interface SSEOptions {
  url: string;
  params?: Record<string, string | number | boolean | undefined>;
  token?: string;
  getToken?: () => string | Promise<string>;
  headers?: HeadersInit;
  lastEventId?: string;
  reconnectDelay?: number;
  maxReconnectAttempts?: number;
  onOpen?: (response: Response) => void;
  onMessage: (event: SSEEvent) => void;
  onError?: (error: Error) => void;
  onClose?: () => void;
}

export interface SSEConnection {
  close: () => void;
}

function buildURL(url: string, params?: SSEOptions['params']): string {
  const host = import.meta.env.VITE_IS_REQUEST_PROXY !== 'true' ? '' : import.meta.env.VITE_API_URL;
  const baseURL = `${host}${import.meta.env.VITE_API_URL_PREFIX}`;
  const target = `${baseURL}${url}`;
  if (!params) return target;
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined) query.set(key, String(value));
  });
  const separator = target.includes('?') ? '&' : '?';
  const text = query.toString();
  return text ? `${target}${separator}${text}` : target;
}

function parseFrame(frame: string): SSEEvent | undefined {
  const fields: Record<string, string[]> = {};
  frame.replace(/\r/g, '').split('\n').forEach((line) => {
    if (!line || line.startsWith(':')) return;
    const separator = line.indexOf(':');
    const key = separator < 0 ? line : line.slice(0, separator);
    const value = separator < 0 ? '' : line.slice(separator + 1).replace(/^ /, '');
    (fields[key] ??= []).push(value);
  });
  if (!fields.data?.length) return undefined;
  return { data: fields.data.join('\n'), event: fields.event?.at(-1) || 'message', id: fields.id?.at(-1) };
}

/** Fetch-based SSE client supports bearer authorization and resumable streams. */
export function openSSE(options: SSEOptions): SSEConnection {
  const controller = new AbortController();
  const reconnectDelay = options.reconnectDelay ?? 1000;
  const maxReconnectAttempts = options.maxReconnectAttempts ?? Infinity;
  let lastEventId = options.lastEventId;
  let reconnectAttempts = 0;

  const connect = async () => {
    while (!controller.signal.aborted) {
      try {
        const token = options.token ?? (await options.getToken?.());
        const headers = new Headers(options.headers);
        headers.set('Accept', 'text/event-stream');
        if (token) headers.set('Authorization', `Bearer ${token}`);
        if (lastEventId) headers.set('Last-Event-ID', lastEventId);

        const response = await fetch(buildURL(options.url, options.params), { headers, signal: controller.signal });
        if (!response.ok || !response.body) throw new Error(`SSE connection failed (${response.status})`);
        options.onOpen?.(response);
        reconnectAttempts = 0;

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';
        while (!controller.signal.aborted) {
          const { done, value } = await reader.read();
          if (done) break;
          buffer += decoder.decode(value, { stream: true });
          const frames = buffer.split(/\r?\n\r?\n/);
          buffer = frames.pop() ?? '';
          frames.forEach((frame) => {
            const event = parseFrame(frame);
            if (!event) return;
            if (event.id) lastEventId = event.id;
            options.onMessage(event);
          });
        }
      } catch (error) {
        if (controller.signal.aborted) break;
        options.onError?.(error instanceof Error ? error : new Error('SSE connection failed'));
      }

      if (controller.signal.aborted || reconnectAttempts >= maxReconnectAttempts) break;
      reconnectAttempts += 1;
      await new Promise((resolve) => window.setTimeout(resolve, reconnectDelay));
    }
    options.onClose?.();
  };

  void connect();
  return { close: () => controller.abort() };
}
