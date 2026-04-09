<script lang="ts">
  let { visible = true }: { visible?: boolean } = $props();
</script>

{#if visible}
  <span class="cursor inline-block w-2 h-4 bg-current align-middle" aria-hidden="true"></span>
{/if}

<style>
  .cursor {
    animation: blink 1.1s step-end infinite;
  }

  @keyframes blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0; }
  }
</style>
