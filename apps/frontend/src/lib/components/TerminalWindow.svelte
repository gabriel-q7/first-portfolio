<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { terminalHistory, isBooting, isReady } from '$lib/stores/terminal';
  import { runBootSequence } from '$lib/utils/boot';
  import { TerminalSocket } from '$lib/services/ws';
  import type { WsServerMessage } from '$lib/types/terminal';
  import TerminalHistory from './TerminalHistory.svelte';
  import TerminalInput from './TerminalInput.svelte';

  let inputComponent: TerminalInput;
  let wsStatus: 'connecting' | 'connected' | 'disconnected' = $state('connecting');
  let socket: TerminalSocket;

  // Lines that are being streamed for the current command (appended incrementally).
  // We track a pending line id so we can update it in-place as chunks arrive.
  let awaitingDone = $state(false);

  function handleWsMessage(msg: WsServerMessage) {
    switch (msg.type) {
      case 'output':
        terminalHistory.addLine({ content: msg.payload, type: 'output' });
        break;
      case 'error':
        terminalHistory.addLine({ content: msg.payload, type: 'error' });
        break;
      case 'status':
        // status messages from the server (e.g. "connected", "Thinking…")
        if (msg.payload !== 'connected') {
          terminalHistory.addLine({ content: msg.payload, type: 'system' });
        }
        break;
      case 'done':
        awaitingDone = false;
        setTimeout(() => inputComponent?.focus(), 50);
        break;
    }
  }

  function handleWsStatus(status: 'connecting' | 'connected' | 'disconnected') {
    wsStatus = status;
  }

  onMount(async () => {
    // Start WebSocket before the boot sequence so it's ready when input opens.
    socket = new TerminalSocket(handleWsMessage, handleWsStatus);
    socket.connect();

    await runBootSequence(
      (line) => terminalHistory.addLine(line),
      () => {
        isBooting.set(false);
        isReady.set(true);
        setTimeout(() => inputComponent?.focus(), 50);
      }
    );
  });

  onDestroy(() => {
    socket?.destroy();
  });

  function handleCommand(input: string) {
    const trimmed = input.trim();

    // Echo the input line
    terminalHistory.addLine({ content: `guest@portfolio:~$ ${input}`, type: 'input' });

    if (!trimmed) return;

    // Handle clear locally — no round trip needed
    if (trimmed.toLowerCase() === 'clear') {
      terminalHistory.clear();
      return;
    }

    // All other commands go to the backend via WebSocket
    if (!socket.isConnected) {
      terminalHistory.addLine({
        content: `  Not connected to backend (status: ${wsStatus}). Retrying…`,
        type: 'error'
      });
      return;
    }

    awaitingDone = true;
    socket.send(trimmed);
  }
</script>

<div
  class="
    flex flex-col w-full max-w-3xl h-[85vh] max-h-[700px]
    bg-[var(--color-terminal-surface)] rounded-lg overflow-hidden
    border border-[var(--color-terminal-border)] border-glow
    relative
  "
  role="region"
  aria-label="Terminal"
>
  <!-- Scanline overlay -->
  <div class="scanlines absolute inset-0 z-10 pointer-events-none rounded-lg opacity-30"></div>

  <!-- Title bar -->
  <div class="flex items-center gap-2 px-4 py-3 border-b border-[var(--color-terminal-border)] shrink-0">
    <span class="w-3 h-3 rounded-full bg-red-500 opacity-80"></span>
    <span class="w-3 h-3 rounded-full bg-yellow-500 opacity-80"></span>
    <span class="w-3 h-3 rounded-full bg-green-500 opacity-80"></span>
    <span class="flex-1 text-center text-xs text-emerald-700 tracking-widest select-none">
      portfolio — bash
    </span>
    <!-- Connection indicator -->
    <span
      class="text-xs select-none {wsStatus === 'connected'
        ? 'text-green-500'
        : wsStatus === 'connecting'
          ? 'text-yellow-500'
          : 'text-red-500'}"
      title="WebSocket: {wsStatus}"
    >
      {wsStatus === 'connected' ? '●' : wsStatus === 'connecting' ? '○' : '✕'}
    </span>
  </div>

  <!-- Output history -->
  <TerminalHistory lines={$terminalHistory} />

  <!-- Busy indicator while a command is running -->
  {#if awaitingDone}
    <div class="px-4 py-1 text-xs text-emerald-700 select-none shrink-0">
      ▸ processing…
    </div>
  {/if}

  <!-- Input line -->
  <TerminalInput
    bind:this={inputComponent}
    disabled={$isBooting || awaitingDone}
    onsubmit={handleCommand}
  />
</div>
