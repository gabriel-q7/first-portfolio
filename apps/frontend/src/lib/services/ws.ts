import type { WsClientMessage, WsServerMessage } from '$lib/types/terminal';

const WS_URL = import.meta.env.VITE_WS_URL ?? 'ws://localhost:8080/ws/terminal';

const RECONNECT_DELAY_MS = 2000;
const MAX_RECONNECT_ATTEMPTS = 10;
const MAX_DELAY_MULTIPLIER = 5;

type MessageHandler = (msg: WsServerMessage) => void;
type StatusHandler = (status: 'connecting' | 'connected' | 'disconnected') => void;

export class TerminalSocket {
  private ws: WebSocket | null = null;
  private onMessage: MessageHandler;
  private onStatus: StatusHandler;
  private reconnectAttempts = 0;
  private destroyed = false;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(onMessage: MessageHandler, onStatus: StatusHandler) {
    this.onMessage = onMessage;
    this.onStatus = onStatus;
  }

  connect(): void {
    if (this.destroyed) return;
    this.onStatus('connecting');
    try {
      this.ws = new WebSocket(WS_URL);
    } catch {
      this.scheduleReconnect();
      return;
    }

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.onStatus('connected');
    };

    this.ws.onmessage = (event: MessageEvent<string>) => {
      try {
        const msg: WsServerMessage = JSON.parse(event.data);
        this.onMessage(msg);
      } catch {
        // Ignore malformed messages
      }
    };

    this.ws.onclose = () => {
      this.onStatus('disconnected');
      if (!this.destroyed) {
        this.scheduleReconnect();
      }
    };

    this.ws.onerror = () => {
      // onclose fires after onerror; reconnect is handled there
    };
  }

  send(payload: string): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const msg: WsClientMessage = { type: 'command', payload };
      this.ws.send(JSON.stringify(msg));
    }
  }

  destroy(): void {
    this.destroyed = true;
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.ws?.close();
    this.ws = null;
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  private scheduleReconnect(): void {
    if (this.destroyed) return;
    if (this.reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) return;
    this.reconnectAttempts++;
    const delay = RECONNECT_DELAY_MS * Math.min(this.reconnectAttempts, MAX_DELAY_MULTIPLIER);
    this.reconnectTimer = setTimeout(() => {
      if (!this.destroyed) this.connect();
    }, delay);
  }
}
