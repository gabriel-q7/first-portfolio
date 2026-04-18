<script lang="ts">
  import XTerminal from '$lib/components/XTerminal.svelte';
</script>

<svelte:head>
  <title>Gabriel — Portfolio</title>
  <meta name="description" content="Software engineer portfolio — terminal edition" />
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="" />
  <link
    href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&display=swap"
    rel="stylesheet"
  />
</svelte:head>

<main
  class="flex items-center justify-center w-screen h-screen bg-[var(--color-terminal-bg)] px-4"
>
  <div
    class="w-full max-w-5xl h-[90vh] flex flex-col rounded-lg overflow-hidden"
    style="border: 1px solid var(--color-terminal-border); box-shadow: 0 0 8px var(--color-terminal-green-muted), inset 0 0 8px rgba(74,222,128,0.03);"
  >
    <!-- Title bar -->
    <div
      class="flex items-center gap-2 px-4 py-2 flex-shrink-0"
      style="background: var(--color-terminal-surface); border-bottom: 1px solid var(--color-terminal-border);"
    >
      <span class="w-3 h-3 rounded-full bg-red-500 opacity-70"></span>
      <span class="w-3 h-3 rounded-full bg-yellow-500 opacity-70"></span>
      <span class="w-3 h-3 rounded-full bg-green-500 opacity-70"></span>
      <span class="flex-1 text-center text-xs text-emerald-700 tracking-widest select-none">
        portfolio — bash
      </span>
    </div>
    <!-- Terminal -->
    <div class="flex-1 min-h-0 p-2" style="background: #0a0e0a;">
      <XTerminal />
    </div>
  </div>
</main>
