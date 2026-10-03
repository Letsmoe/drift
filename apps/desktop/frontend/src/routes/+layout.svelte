<script lang="ts">
	import '@fontsource/geist-sans/400.css';
	import '@fontsource/geist-sans/500.css';
	import '@fontsource/geist-sans/600.css';
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<title>Otter Music</title>
</svelte:head>

<div class="flex h-screen flex-col overflow-hidden bg-canvas text-default">
	{@render children()}
</div>
