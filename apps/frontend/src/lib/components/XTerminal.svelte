<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { WebLinksAddon } from '@xterm/addon-web-links';
  import { WebSocketService } from '$lib/services/websocket';
  import type { ServerMessage } from '$lib/types/protocol';

  /** WebSocket URL — falls back to ws://localhost:8080/ws */
  const wsUrl: string = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';

  let container: HTMLDivElement;
  let term: InstanceType<typeof Terminal>;
  let fitAddon: InstanceType<typeof FitAddon>;
  let wsService: WebSocketService;

  // Command history for up/down arrow navigation.
  let history: string[] = [];
  let historyIndex = -1;

  // Current input buffer.
  let inputBuffer = '';
  // Whether a command is currently executing.
  let busy = false;
  // Current active request ID.
  let currentRequestId = '';

  const PROMPT = '\r\n\x1b[32mguest@portfolio\x1b[0m:\x1b[36m~\x1b[0m$ ';
  const PROMPT_BUSY = '\x1b[33m...\x1b[0m ';

  function generateId(): string {
    return Math.random().toString(36).slice(2, 10);
  }

  function writePrompt() {
    term.write(PROMPT);
  }

  function handleServerMessage(msg: ServerMessage) {
    switch (msg.type) {
      case 'status':
        term.writeln('');
        term.writeln('\x1b[32m  ' + (msg.content ?? '') + '\x1b[0m');
        writePrompt();
        break;

      case 'output':
        if (msg.content === '__CLEAR__') {
          term.clear();
          writePrompt();
          return;
        }
        term.writeln(msg.content ?? '');
        break;

      case 'error':
        term.writeln('\x1b[31m' + (msg.content ?? 'error') + '\x1b[0m');
        break;

      case 'done':
        busy = false;
        writePrompt();
        break;
    }
  }

  function sendCommand(input: string) {
    const trimmed = input.trim();
    if (!trimmed) {
      writePrompt();
      return;
    }

    // Add to history.
    if (history[history.length - 1] !== trimmed) {
      history = [...history, trimmed];
    }
    historyIndex = -1;

    busy = true;
    currentRequestId = generateId();

    wsService.send({
      type: 'command',
      request_id: currentRequestId,
      command: trimmed,
      timestamp: new Date().toISOString()
    });
  }

  function handleData(data: string) {
    for (const char of data) {
      const code = char.charCodeAt(0);

      // Ctrl+C — cancel running command or clear input.
      if (code === 3) {
        if (busy) {
          wsService.send({
            type: 'cancel',
            request_id: currentRequestId,
            timestamp: new Date().toISOString()
          });
          term.writeln('^C');
          busy = false;
          writePrompt();
        } else {
          term.write('^C');
          inputBuffer = '';
          writePrompt();
        }
        return;
      }

      // Enter.
      if (code === 13) {
        term.writeln('');
        if (busy) return;
        const cmd = inputBuffer;
        inputBuffer = '';
        sendCommand(cmd);
        return;
      }

      // Backspace.
      if (code === 127) {
        if (inputBuffer.length > 0) {
          inputBuffer = inputBuffer.slice(0, -1);
          term.write('\b \b');
        }
        return;
      }

      // Arrow up — previous history.
      if (data === '\x1b[A') {
        if (history.length === 0) return;
        if (historyIndex === -1) historyIndex = history.length - 1;
        else if (historyIndex > 0) historyIndex--;
        replaceInput(history[historyIndex]);
        return;
      }

      // Arrow down — next history.
      if (data === '\x1b[B') {
        if (historyIndex === -1) return;
        if (historyIndex < history.length - 1) {
          historyIndex++;
          replaceInput(history[historyIndex]);
        } else {
          historyIndex = -1;
          replaceInput('');
        }
        return;
      }

      // Ignore other escape sequences and control chars.
      if (code < 32) return;

      // Printable character.
      if (!busy) {
        inputBuffer += char;
        term.write(char);
      }
    }
  }

  function replaceInput(newInput: string) {
    // Erase current input on the line.
    term.write('\r' + PROMPT.replace('\r\n', '') + '\x1b[K');
    inputBuffer = newInput;
    term.write(newInput);
  }

  onMount(() => {
    term = new Terminal({
      cursorBlink: true,
      fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', 'Courier New', monospace",
      fontSize: 14,
      lineHeight: 1.6,
      theme: {
        background: '#0a0e0a',
        foreground: '#d1fae5',
        cursor: '#4ade80',
        cursorAccent: '#0a0e0a',
        selectionBackground: '#166534',
        black: '#0a0e0a',
        red: '#f87171',
        green: '#4ade80',
        yellow: '#fbbf24',
        blue: '#60a5fa',
        magenta: '#c084fc',
        cyan: '#67e8f9',
        white: '#d1fae5',
        brightBlack: '#374151',
        brightGreen: '#22c55e',
        brightYellow: '#fde68a',
        brightCyan: '#a5f3fc',
        brightWhite: '#f0fdf4'
      }
    });

    fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.loadAddon(new WebLinksAddon());
    term.open(container);
    fitAddon.fit();

    term.onData(handleData);

    // Restore history from sessionStorage.
    try {
      const stored = sessionStorage.getItem('terminal_history');
      if (stored) history = JSON.parse(stored) as string[];
    } catch {
      // Ignore storage errors.
    }

    // Connect to WebSocket.
    wsService = new WebSocketService({
      url: wsUrl,
      onMessage: handleServerMessage,
      onOpen: () => {
        // Status message handled via handleServerMessage.
      },
      onClose: () => {
        if (!busy) {
          term.writeln('\r\n\x1b[33m  [disconnected — reconnecting...]\x1b[0m');
        }
      },
      onError: () => {
        term.writeln('\r\n\x1b[31m  [connection error]\x1b[0m');
      }
    });
    wsService.connect();

    // Fit on resize.
    const ro = new ResizeObserver(() => fitAddon.fit());
    ro.observe(container);

    return () => ro.disconnect();
  });

  onDestroy(() => {
    // Save history to sessionStorage.
    try {
      sessionStorage.setItem('terminal_history', JSON.stringify(history));
    } catch {
      // Ignore storage errors.
    }
    wsService?.close();
    term?.dispose();
  });
</script>

<div
  bind:this={container}
  class="w-full h-full"
  style="background: #0a0e0a;"
></div>

<style>
  :global(.xterm) {
    height: 100%;
  }
  :global(.xterm-viewport) {
    scrollbar-width: thin;
    scrollbar-color: #166534 #0a0e0a;
  }
</style>
