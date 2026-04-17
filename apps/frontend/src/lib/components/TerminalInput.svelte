<script lang="ts">
  import Cursor from './Cursor.svelte';

  let {
    disabled = false,
    onsubmit
  }: {
    disabled?: boolean;
    onsubmit?: (value: string) => void;
  } = $props();

  let value = $state('');
  let inputEl: HTMLInputElement;

  // Command history
  let history: string[] = $state([]);
  let historyIndex = $state(-1);
  let savedInput = $state('');

  export function focus() {
    inputEl?.focus();
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      const trimmed = value.trim();
      if (trimmed) {
        history = [...history, trimmed];
        historyIndex = -1;
        savedInput = '';
      }
      onsubmit?.(value);
      value = '';
      return;
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault();
      if (history.length === 0) return;
      if (historyIndex === -1) {
        savedInput = value;
        historyIndex = history.length - 1;
      } else if (historyIndex > 0) {
        historyIndex--;
      }
      value = history[historyIndex];
      return;
    }

    if (event.key === 'ArrowDown') {
      event.preventDefault();
      if (historyIndex === -1) return;
      if (historyIndex < history.length - 1) {
        historyIndex++;
        value = history[historyIndex];
      } else {
        historyIndex = -1;
        value = savedInput;
      }
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div
  class="flex items-center gap-2 px-4 py-3 border-t border-[var(--color-terminal-border)]"
  onclick={() => inputEl?.focus()}
>
  <span class="text-green-400 select-none shrink-0">guest@portfolio:~$</span>
  <div class="relative flex-1 flex items-center">
    <input
      bind:this={inputEl}
      bind:value
      onkeydown={handleKeydown}
      {disabled}
      type="text"
      autocomplete="off"
      autocorrect="off"
      autocapitalize="off"
      spellcheck="false"
      aria-label="Terminal input"
      class="
        w-full bg-transparent outline-none border-none caret-transparent
        text-green-400 font-mono text-sm
      "
    />
    {#if !disabled}
      <Cursor />
    {/if}
  </div>
</div>
