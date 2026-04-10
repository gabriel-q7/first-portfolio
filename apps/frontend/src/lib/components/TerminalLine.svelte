<script lang="ts">
  import type { TerminalLine } from '$lib/types/terminal';

  let { line }: { line: TerminalLine } = $props();

  const colorMap: Record<TerminalLine['type'], string> = {
    input: 'text-green-400',
    output: 'text-emerald-200',
    error: 'text-red-400',
    system: 'text-green-500',
    boot: 'text-emerald-600'
  };
</script>

<div
  class="terminal-line whitespace-pre-wrap break-words {colorMap[line.type]}"
  role="listitem"
>
  {line.content}
</div>
