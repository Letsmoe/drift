<script lang="ts">
	import type { Item, ItemKind } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import TrackContextMenu from '#lib/components/TrackContextMenu.svelte';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { itemHref, trackHref } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let { item, context = 'home' }: { item: Item; context?: string } = $props();

	const playing = $derived(isTrackPlaying());
	const contextMenu = new PopupMenu();
	// Songs and episodes open their album or show; the play button plays them in place.
	const pageHref = $derived(hrefOf(item));
	const subtitle = $derived(subtitleOf(item));

	// Song cards already carry "Artist • 6.2M plays" in YouTube's subtitle; episodes don't.
	function subtitleOf(shown: Item): string {
		const playCount = shown.track?.playCount;
		if (!playCount || shown.subtitle.includes(playCount)) {
			return shown.subtitle;
		}
		return [shown.subtitle, playCount].filter((part) => part !== '').join(' • ');
	}

	function hrefOf(shown: Item): string | undefined {
		if (shown.track) {
			return trackHref(shown.track);
		}
		return itemHref(shown.kind, shown.id);
	}

	function isTrackPlaying(): boolean {
		if (!item.track) {
			return false;
		}
		return player.isPlaying(item.track);
	}

	// Only songs and episodes have a track menu; other cards keep the browser default.
	function openContextMenu(event: MouseEvent) {
		if (item.track) {
			contextMenu.openAtPointer(event);
		}
	}

	function playTrack() {
		if (item.track) {
			player.playOrToggle(item.track, [item.track], 0, context, item.kind as ItemKind);
		}
	}
</script>

<article
	class="group/card duration-fast relative flex w-56 shrink-0 flex-col gap-2 rounded-lg p-3 transition-colors hover:bg-hover"
	oncontextmenu={openContextMenu}
>
	<div class="relative">
		<Cover
			thumbnails={item.thumbnails}
			size={200}
			class={[
				'w-full shadow-md',
				item.kind === 'artist' && 'rounded-full',
				item.kind !== 'artist' && 'rounded-md'
			]}
		/>
		{#if item.track}
			<PlayOverlayButton
				{playing}
				label="Play {item.title}"
				onclick={playTrack}
				class={[
					'absolute right-2 bottom-2 z-10',
					!playing &&
						'translate-y-2 opacity-0 group-hover/card:translate-y-0 group-hover/card:opacity-100'
				]}
			/>
		{/if}
	</div>
	<div class="min-w-0">
		<h3 class="truncate text-base font-medium">
			{#if pageHref}
				<!-- Stretched over the card: the card opens the page, the play button sits above it. -->
				<a href={pageHref} class="after:absolute after:inset-0">{item.title}</a>
			{:else}
				{item.title}
			{/if}
		</h3>
		<p class="line-clamp-2 text-sm text-dim">{subtitle}</p>
	</div>
</article>

{#if item.track}
	<TrackContextMenu popup={contextMenu} track={item.track} kind={item.kind as ItemKind} />
{/if}
