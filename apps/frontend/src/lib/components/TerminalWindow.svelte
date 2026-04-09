<script lang="ts">
  import { onMount } from 'svelte';
  import { terminalHistory, isBooting, isReady } from '$lib/stores/terminal';
  import { runBootSequence } from '$lib/utils/boot';
  import { executeCommand } from '$lib/commands/index';
  import TerminalHistory from './TerminalHistory.svelte';
  import TerminalInput from './TerminalInput.svelte';

  let inputComponent: TerminalInput;

  onMount(async () => {
    await runBootSequence(
      (line) => terminalHistory.addLine(line),
      () => {
        isBooting.set(false);
        isReady.set(true);
        setTimeout(() => inputComponent?.focus(), 50);
      }
    );
  });

  function handleCommand(input: string) {
    const results = executeCommand(input);

    const clearIndex = results.findIndex((r) => r.content === '__CLEAR__');
    if (clearIndex !== -1) {
      terminalHistory.clear();
      return;
    }

    terminalHistory.addLines(results);
    setTimeout(() => inputComponent?.focus(), 50);
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
  </div>

  <!-- Output history -->
  <TerminalHistory lines={$terminalHistory} />

  <!-- Input line -->
  <TerminalInput
    bind:this={inputComponent}
    disabled={$isBooting}
    onsubmit={handleCommand}
  />
</div>
