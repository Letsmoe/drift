<script lang="ts">
	import DotsThreeIcon from 'phosphor-svelte/lib/DotsThreeIcon';
	import HeartIcon from 'phosphor-svelte/lib/HeartIcon';
	import PauseIcon from 'phosphor-svelte/lib/PauseIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import SpeakerHighIcon from 'phosphor-svelte/lib/SpeakerHighIcon';
	import type { ItemKind, Track } from '#lib/api/music.ts';
	import ArtistLinks from '#lib/components/ArtistLinks.svelte';
	import Cover from '#lib/components/Cover.svelte';
	import DownloadButton from '#lib/components/DownloadButton.svelte';
	import DownloadedIndicator from '#lib/components/DownloadedIndicator.svelte';
	import Menu from '#lib/components/Menu.svelte';
	import { formatDuration, itemHref } from '#lib/format.ts';
	import { library } from '#lib/library/library.svelte.ts';
	import { PopupMenu } from '#lib/menu/popupMenu.svelte.ts';
	import { buildTrackMenuEntries } from '#lib/menu/trackMenuEntries.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let {
		track,
		index,
		gridClass,
		kind = 'song',
		showCover,
		showAlbum,
		dateAdded = undefined,
		onplay,
		onremove = undefined
	}: {
		track: Track;
		index: number;
		gridClass: string;
		kind?: ItemKind;
		showCover: boolean;
		showAlbum: boolean;
		/** Set to show the "Date added" column. */
		dateAdded?: string;
		onplay: () => void;
		onremove?: () => void;
	} = $props();

	const popup = new PopupMenu();
	const current = $derived(player.isCurrent(track));
	const playing = $derived(player.isPlaying(track));
	const liked = $derived(library.isLiked(track.videoId));

	const likeLabel = $derived(describeLikeAction(liked));

	function describeLikeAction(isLiked: boolean): string {
		if (isLiked) {
			return 'Remove from Liked Songs';
		}
		return 'Save to Liked Songs';
	}

	function heartWeightFor(isLiked: boolean): 'fill' | 'regular' {
		if (isLiked) {
			return 'fill';
		}
		return 'regular';
	}

	function openContextMenu(event: MouseEvent) {
		event.preventDefault();
		popup.openAt(event.clientX, event.clientY);
	}

	function openMoreMenu(event: MouseEvent) {
		event.stopPropagation();
		popup.openBelow(event.currentTarget as HTMLElement);
	}

	function entries() {
		return buildTrackMenuEntries(track, { kind, onRemoveFromPlaylist: onremove });
	}
</script>

<div
	role="row"
	tabindex="-1"
	oncontextmenu={openContextMenu}
	ondblclick={onplay}
	class={[
		'group/row duration-fast grid h-14 items-center gap-4 rounded-md px-4 transition-colors focus-within:bg-hover hover:bg-hover',
		popup.isOpen && 'bg-hover',
		gridClass
	]}
>
	<div class="flex w-4 items-center justify-end text-sm text-dim tabular-nums">
		<span
			class={['group-focus-within/row:hidden group-hover/row:hidden', current && 'text-action']}
		>
			{#if playing}
				<SpeakerHighIcon size={16} weight="fill" />
			{:else}
				{index + 1}
			{/if}
		</span>
		<button
			type="button"
			aria-label="Play {track.title}"
			onclick={onplay}
			class="hidden text-default group-focus-within/row:block group-hover/row:block"
		>
			{#if playing}
				<PauseIcon size={16} weight="fill" />
			{:else}
				<PlayIcon size={16} weight="fill" />
			{/if}
		</button>
	</div>

	<div class="flex min-w-0 items-center gap-3">
		{#if showCover}
			<Cover thumbnails={track.thumbnails} size={40} class="size-10 shrink-0 rounded-sm" />
		{/if}
		<div class="min-w-0">
			<div class="flex min-w-0 items-center gap-1.5">
				<p class={['truncate text-base', current && 'text-action']}>{track.title}</p>
				<DownloadedIndicator videoId={track.videoId} />
			</div>
			<p class="truncate text-sm text-dim"><ArtistLinks artists={track.artists} /></p>
		</div>
	</div>

	{#if showAlbum}
		<p class="truncate text-sm text-dim">
			{#if track.album && track.album.id}
				<a href={itemHref('album', track.album.id)} class="hover:text-default hover:underline">
					{track.album.name}
				</a>
			{/if}
		</p>
	{/if}

	{#if dateAdded !== undefined}
		<p class="truncate text-sm text-dim">{dateAdded}</p>
	{/if}

	<div class="flex items-center justify-end gap-3 text-dim">
		<DownloadButton {track} {kind} />
		<button
			type="button"
			aria-label={likeLabel}
			aria-pressed={liked}
			onclick={() => library.toggleLike(track, kind)}
			class={[
				'group-hover/row:opacity-100 hover:text-default focus-visible:opacity-100',
				!liked && 'opacity-0',
				liked && 'text-action hover:text-action'
			]}
		>
			<HeartIcon size={18} weight={heartWeightFor(liked)} />
		</button>
		<span class="w-10 text-right text-sm tabular-nums">{formatDuration(track.durationSeconds)}</span
		>
		<button
			type="button"
			aria-label="More options for {track.title}"
			aria-haspopup="menu"
			onclick={openMoreMenu}
			class="opacity-0 group-focus-within/row:opacity-100 group-hover/row:opacity-100 hover:text-default"
		>
			<DotsThreeIcon size={22} weight="bold" />
		</button>
	</div>
</div>

{#if popup.position}
	<Menu
		entries={entries()}
		x={popup.position.x}
		y={popup.position.y}
		onclose={() => popup.close()}
	/>
{/if}
