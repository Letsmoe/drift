<script lang="ts">
	import type { Track } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import TrackContextMenu from '#lib/components/TrackContextMenu.svelte';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { formatArtists, trackHref } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { whenVisible } from '#lib/whenVisible.ts';

	let { track, onplay, onvisible }: { track: Track; onplay: () => void; onvisible?: () => void } =
		$props();

	const contextMenu = new PopupMenu();
	const playing = $derived(player.isPlaying(track));
	const current = $derived(player.isCurrent(track));
	const pageHref = $derived(trackHref(track));

	function reportVisible(): void {
		onvisible?.();
	}
</script>

<div
	role="group"
	aria-label={track.title}
	class="group/card duration-fast relative flex w-40 shrink-0 flex-col gap-2 rounded-lg p-3 transition-colors hover:bg-raised"
	{@attach whenVisible(reportVisible)}
	oncontextmenu={(event) => contextMenu.openAtPointer(event)}
>
	<div class="relative">
		<Cover thumbnails={track.thumbnails} size={136} class="w-full rounded-md" />
		<PlayOverlayButton
			{playing}
			size="md"
			label="Play {track.title}"
			onclick={onplay}
			class={[
				'absolute right-2 bottom-2 z-10',
				!playing &&
					'translate-y-1 opacity-0 group-hover/card:translate-y-0 group-hover/card:opacity-100'
			]}
		/>
	</div>
	<div class="min-w-0">
		<p class="truncate text-sm font-semibold" class:text-action={current}>
			{#if pageHref}
				<!-- Stretched over the card: the card opens the page, the play button sits above it. -->
				<a href={pageHref} class="after:absolute after:inset-0">{track.title}</a>
			{:else}
				{track.title}
			{/if}
		</p>
		<p class="truncate text-xs text-dim">{formatArtists(track.artists)}</p>
	</div>
</div>

<TrackContextMenu popup={contextMenu} {track} />
