<script lang="ts">
  import { tick } from 'svelte';
  import type { TerminalLine } from '$lib/types/terminal';
  import TerminalLineComponent from './TerminalLine.svelte';

  let { lines }: { lines: TerminalLine[] } = $props();

  let container: HTMLDivElement;

  $effect(() => {
    // Re-run whenever lines changes to auto-scroll
    lines;
    tick().then(() => {
      if (container) {
        container.scrollTop = container.scrollHeight;
      }
    });
  });
</script>

<div
  bind:this={container}
  class="flex-1 overflow-y-auto px-4 py-2 space-y-0.5"
  role="list"
  aria-label="Terminal output"
  aria-live="polite"
  aria-atomic="false"
>
  {#each lines as line (line.id)}
    <TerminalLineComponent {line} />
  {/each}
</div>
