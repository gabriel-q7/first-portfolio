import { writable } from 'svelte/store';
import type { TerminalLine } from '$lib/types/terminal';

function generateId(): string {
  return Math.random().toString(36).slice(2, 9);
}

function createTerminalStore() {
  const { subscribe, update, set } = writable<TerminalLine[]>([]);

  return {
    subscribe,
    addLine(line: Omit<TerminalLine, 'id' | 'timestamp'>): void {
      update((lines) => [
        ...lines,
        { ...line, id: generateId(), timestamp: new Date() }
      ]);
    },
    addLines(newLines: Omit<TerminalLine, 'id' | 'timestamp'>[]): void {
      update((lines) => [
        ...lines,
        ...newLines.map((l) => ({ ...l, id: generateId(), timestamp: new Date() }))
      ]);
    },
    clear(): void {
      set([]);
    }
  };
}

export const terminalHistory = createTerminalStore();
export const isBooting = writable(true);
export const isReady = writable(false);
