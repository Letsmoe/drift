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
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { panels } from '#lib/layout/panels.svelte.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { nowPlayingPalette } from '#lib/palette/coverPalette.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let { children } = $props();

	library.load();
	downloads.start();

	$effect(() => {
		if (player.track === null) {
			nowPlayingPalette.load([]);
			return;
		}
		nowPlayingPalette.load(player.track.thumbnails);
	});
</script>

<svelte:head>
	<title>Otter Music</title>
</svelte:head>

<div
	class="art-shell flex h-screen flex-col gap-2 overflow-hidden bg-canvas p-2 text-default"
	style={nowPlayingPalette.style()}
>
	<TopBar />
	<div class="flex min-h-0 flex-1 gap-2">
		<LibrarySidebar />
		<FloatingScrollbar class="art-wash min-w-0 flex-1 rounded-2xl">
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
