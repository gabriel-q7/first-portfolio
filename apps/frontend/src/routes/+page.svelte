<script lang="ts">
  import TerminalWindow from '$lib/components/TerminalWindow.svelte';
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
  <TerminalWindow />
</main>
