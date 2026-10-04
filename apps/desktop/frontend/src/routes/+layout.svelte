<script lang="ts">
	import '@fontsource/geist-sans/400.css';
	import '@fontsource/geist-sans/500.css';
	import '@fontsource/geist-sans/600.css';
	import '@fontsource/geist-sans/700.css';
	import '../app.css';
	import { FloatingScrollbar } from '@neoworks-dev/ui';
	import LibrarySidebar from '#lib/components/LibrarySidebar.svelte';
	import PlayerBar from '#lib/components/PlayerBar.svelte';
	import QueuePanel from '#lib/components/QueuePanel.svelte';
	import TopBar from '#lib/components/TopBar.svelte';
	import { panels } from '#lib/layout/panels.svelte.ts';
	import { library } from '#lib/library/library.svelte.ts';

	let { children } = $props();

	library.load();
</script>

<svelte:head>
	<title>Otter Music</title>
</svelte:head>

<div class="flex h-screen flex-col gap-2 overflow-hidden bg-canvas p-2 text-default">
	<TopBar />
	<div class="flex min-h-0 flex-1 gap-2">
		<LibrarySidebar />
		<FloatingScrollbar class="min-w-0 flex-1 rounded-lg bg-surface">
			<main class="px-3 py-4">
				{@render children()}
			</main>
		</FloatingScrollbar>
		{#if panels.queueOpen}
			<QueuePanel />
		{/if}
	</div>
	<PlayerBar />
</div>
