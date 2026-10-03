<script lang="ts">
	import '@fontsource/geist-sans/400.css';
	import '@fontsource/geist-sans/500.css';
	import '@fontsource/geist-sans/600.css';
	import '../app.css';
	import PlayerBar from '#lib/components/PlayerBar.svelte';

	let { children } = $props();
</script>

<svelte:head>
	<title>Otter Music</title>
</svelte:head>

<div class="flex h-screen flex-col overflow-hidden bg-canvas text-default">
	<div class="flex min-h-0 flex-1 flex-col">
		{@render children()}
	</div>
	<PlayerBar />
</div>
