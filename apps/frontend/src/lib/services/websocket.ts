import type { ClientMessage, ServerMessage } from '$lib/types/protocol';

export type MessageHandler = (msg: ServerMessage) => void;

export interface WebSocketServiceOptions {
  url: string;
  onMessage: MessageHandler;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: (event: Event) => void;
  /** Maximum number of reconnect attempts (default: 10). */
  maxRetries?: number;
}

/**
 * WebSocketService manages a WebSocket connection with automatic reconnection
 * using exponential backoff.
 */
export class WebSocketService {
  private ws: WebSocket | null = null;
  private retryCount = 0;
  private retryTimeout: ReturnType<typeof setTimeout> | null = null;
  private closed = false;

  private readonly url: string;
  private readonly onMessage: MessageHandler;
  private readonly onOpen?: () => void;
  private readonly onClose?: () => void;
  private readonly onError?: (event: Event) => void;
  private readonly maxRetries: number;

  constructor(opts: WebSocketServiceOptions) {
    this.url = opts.url;
    this.onMessage = opts.onMessage;
    this.onOpen = opts.onOpen;
    this.onClose = opts.onClose;
    this.onError = opts.onError;
    this.maxRetries = opts.maxRetries ?? 10;
  }

  /** Connect (or reconnect) to the WebSocket server. */
  connect(): void {
    if (this.closed) return;

    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      this.retryCount = 0;
      this.onOpen?.();
    };

    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const msg: ServerMessage = JSON.parse(event.data as string);
        this.onMessage(msg);
      } catch {
        // Ignore malformed messages.
      }
    };

    this.ws.onerror = (event: Event) => {
      this.onError?.(event);
    };

    this.ws.onclose = () => {
      this.onClose?.();
      if (!this.closed && this.retryCount < this.maxRetries) {
        const delay = Math.min(1000 * 2 ** this.retryCount, 30_000);
        this.retryCount++;
        this.retryTimeout = setTimeout(() => this.connect(), delay);
      }
    };
  }

  /** Send a command message to the server. */
  send(msg: ClientMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  /** Close the connection permanently (no reconnect). */
  close(): void {
    this.closed = true;
    if (this.retryTimeout !== null) {
      clearTimeout(this.retryTimeout);
      this.retryTimeout = null;
    }
    this.ws?.close();
  }

  /** True if the underlying WebSocket is currently open. */
  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}
