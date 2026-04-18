export type MessageType = 'command' | 'output' | 'done' | 'error' | 'status' | 'cancel';

/** Message sent from client to server. */
export interface ClientMessage {
  type: MessageType;
  request_id: string;
  command?: string;
  timestamp: string;
}

/** Message sent from server to client. */
export interface ServerMessage {
  type: MessageType;
  request_id?: string;
  content?: string;
  timestamp: string;
}
