<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import '../app.css';
	import '@xterm/xterm/css/xterm.css';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}
