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
  handler: () => TerminalLine[];
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
