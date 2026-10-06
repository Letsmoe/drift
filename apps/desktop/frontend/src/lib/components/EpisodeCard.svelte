<script lang="ts">
	import type { Item, Track } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import TrackContextMenu from '#lib/components/TrackContextMenu.svelte';
	import { formatDuration, playCountOf, trackHref } from '#lib/format.ts';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let {
		item,
		track,
		tracks = [track],
		index = 0,
		context = 'home'
	}: { item: Item; track: Track; tracks?: Track[]; index?: number; context?: string } = $props();

	const contextMenu = new PopupMenu();
	const playing = $derived(player.isPlaying(track));
	const current = $derived(player.isCurrent(track));
	const details = $derived(
		[publishedOf(item), formatDuration(track.durationSeconds), playCountOf(track)]
			.filter((part) => part !== '')
			.join(' • ')
	);

	function publishedOf(shown: Item): string {
		if (shown.published) {
			return shown.published;
		}
		return '';
	}

	function play(): void {
		player.playOrToggle(track, tracks, index, context, 'episode');
	}
</script>

<article
	aria-label={track.title}
	oncontextmenu={(event) => contextMenu.openAtPointer(event)}
	class={[
		'group/episode duration-fast relative flex gap-3 rounded-lg p-2 transition-colors',
		current && 'bg-hover ring-1 ring-action/40',
		!current && 'bg-raised hover:bg-hover'
	]}
>
	<div class="relative w-36 shrink-0 self-start">
		<Cover thumbnails={track.thumbnails} size={144} aspect="video" class="w-full rounded-md" />
		<PlayOverlayButton
			{playing}
			size="sm"
			label="Play {track.title}"
			onclick={play}
			class={[
				'absolute right-1.5 bottom-1.5 z-10',
				!playing && 'opacity-0 group-hover/episode:opacity-100'
			]}
		/>
	</div>
	<div class="flex min-w-0 flex-1 flex-col gap-0.5">
		<h3 class={['line-clamp-2 text-sm leading-snug font-semibold', current && 'text-action']}>
			<!-- Stretched over the card: the card opens the show, the play button sits above it. -->
			<a href={trackHref(track)} class="after:absolute after:inset-0">{track.title}</a>
		</h3>
		<p class="truncate text-xs text-muted">{item.subtitle}</p>
		{#if item.description}
			<p class="line-clamp-2 text-xs text-dim">{item.description}</p>
		{/if}
		<p class="mt-auto truncate pt-1 text-xs text-faint">{details}</p>
	</div>
</article>

<TrackContextMenu popup={contextMenu} {track} kind="episode" />
