<script lang="ts">
	import { Presence, Text, View } from '@neoworks/svelte-gpui/ui';
	import { fade, fly } from 'svelte/transition';
	import { pages } from 'virtual:pages';
	import ArtBackdrop from '#lib/components/ArtBackdrop.svelte';
	import FullPlayer from '#lib/components/FullPlayer.svelte';
	import LibrarySidebar from '#lib/components/LibrarySidebar.svelte';
	import PlayerBar from '#lib/components/PlayerBar.svelte';
	import QueuePanel from '#lib/components/QueuePanel.svelte';
	import TopBar from '#lib/components/TopBar.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { panels } from '#lib/layout/panels.svelte.ts';
	import { textFocus } from '#lib/layout/textFocus.svelte.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { nowPlayingPalette } from '#lib/palette/coverPalette.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { router } from '#lib/router.svelte.ts';
	import { colors, glassBackground, radii } from '#lib/theme.ts';

	const pageEnterDuration = 180;
	const pageEnterOffset = 10;
	const pageFrameStyle = { flex: 1, minHeight: 0 };

	library.load();
	downloads.start();

	const Page = $derived(pages[router.page]);
	const showQueue = $derived(panels.queueOpen && player.queue.entries.length > 0);

	$effect(() => {
		if (player.track === null) {
			nowPlayingPalette.load([]);
			return;
		}
		nowPlayingPalette.load(player.track.thumbnails);
	});

	// Space plays and pauses, unless a text field is being typed in.
	function handleKeydown(event: KeyboardEvent): void {
		if (event.key !== ' ' || textFocus.typing) {
			return;
		}
		event.preventDefault();
		player.toggle();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<View style={{ flex: 1, background: colors.canvas, color: colors.text, padding: 8, gap: 8 }}>
	<ArtBackdrop />
	<TopBar />
	<View style={{ flex: 1, minHeight: 0, direction: 'row', gap: 8 }}>
		<LibrarySidebar />
		<View
			style={{
				flex: 1,
				minWidth: 0,
				radius: radii.extraLarge,
				background: glassBackground,
				overflow: 'hidden'
			}}
		>
			<!-- A page fills this box and scrolls itself (PageList), so it needs a bounded parent. -->
			<!--
				A new path fades and slides its page in; a new query on the same path only rebuilds the page. There is
				no outro: the old page would stay laid out beside the new one until it finished.
			-->
			{#key router.path}
				<Presence
					in={[fly, { y: pageEnterOffset, duration: pageEnterDuration }]}
					style={pageFrameStyle}
				>
					{#key router.href}
						{#if Page}
							<Page params={router.params} query={router.query} />
						{:else}
							<View style={{ gap: 4, padding: 16 }}>
								<Text style={{ fontSize: 18, fontWeight: 'semibold' }}>{router.page}</Text>
								<Text style={{ color: colors.muted }}>
									This page is not written yet ({router.href}).
								</Text>
							</View>
						{/if}
					{/key}
				</Presence>
			{/key}
		</View>
		{#if showQueue}
			<QueuePanel />
		{/if}
	</View>
	<PlayerBar />
	{#if panels.fullPlayerOpen}
		<Presence
			in={[fly, { y: 48, duration: 320 }]}
			out={[fade, { duration: 160 }]}
			style={{ position: 'absolute', inset: 0 }}
		>
			<FullPlayer />
		</Presence>
	{/if}
</View>
