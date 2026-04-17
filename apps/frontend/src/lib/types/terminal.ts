export type CommandName = 'help' | 'about' | 'projects' | 'contact' | 'clear' | 'unknown';

export interface TerminalLine {
  id: string;
  type: 'input' | 'output' | 'error' | 'system' | 'boot';
  content: string;
  timestamp?: Date;
}

export interface Command {
  name: string;
  description: string;
  handler: () => Omit<TerminalLine, 'id' | 'timestamp'>[];
}

export interface Project {
  name: string;
  description: string;
  tech: string[];
  url?: string;
}

export interface ApiResponse<T> {
  data: T | null;
  error: string | null;
  loading: boolean;
}

// WebSocket protocol types shared with the backend.
export type WsMessageType = 'command' | 'output' | 'done' | 'error' | 'status';

export interface WsClientMessage {
  type: 'command';
  payload: string;
}

export interface WsServerMessage {
  type: WsMessageType;
  payload: string;
}
