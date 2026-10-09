<script lang="ts">
	import '@fontsource/geist-sans/400.css';
	import '@fontsource/geist-sans/500.css';
	import '@fontsource/geist-sans/600.css';
	import '@fontsource/geist-sans/700.css';
	import '../app.css';
	import { FloatingScrollbar } from '@neoworks-dev/ui';
	import ArtBackdrop from '#lib/components/ArtBackdrop.svelte';
	import FullPlayer from '#lib/components/FullPlayer.svelte';
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

	// Present only while a cover tints the shell; app.css switches the accent on it.
	function artAttribute(): string | undefined {
		if (nowPlayingPalette.hasColors()) {
			return '';
		}
		return undefined;
	}

	$effect(() => {
		if (player.track === null) {
			nowPlayingPalette.load([]);
			return;
		}
		nowPlayingPalette.load(player.track.thumbnails);
	});
</script>

<svelte:head>
	<title>Drift</title>
</svelte:head>

<div
	class="art-shell relative isolate flex h-screen flex-col gap-2 overflow-hidden bg-canvas p-2 text-default"
	style={nowPlayingPalette.style()}
	data-art={artAttribute()}
>
	<ArtBackdrop class="-z-10" />
	<TopBar />
	<div class="flex min-h-0 flex-1 gap-2">
		<LibrarySidebar />
		<FloatingScrollbar class="glass min-w-0 flex-1 rounded-2xl">
			<main class="px-3 py-4">
				{@render children()}
			</main>
		</FloatingScrollbar>
		{#if panels.queueOpen && player.queue.entries.length > 0}
			<QueuePanel />
		{/if}
	</div>
	<PlayerBar />
	{#if panels.fullPlayerOpen}
		<FullPlayer />
	{/if}
</div>
