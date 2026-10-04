<script lang="ts">
	import type { Item, ItemKind } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { itemHref } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let { item, context }: { item: Item; context: string } = $props();

	const playing = $derived(isTrackPlaying());
	const pageHref = $derived(hrefOf(item));

	function hrefOf(shown: Item): string | undefined {
		if (shown.track) {
			return undefined;
		}
		return itemHref(shown.kind, shown.id);
	}

	function isTrackPlaying(): boolean {
		if (!item.track) {
			return false;
		}
		return player.isPlaying(item.track);
	}

	function playTrack() {
		if (item.track) {
			player.playOrToggle(item.track, [item.track], 0, context, item.kind as ItemKind);
		}
	}
</script>

{#snippet card()}
	<Cover
		thumbnails={item.thumbnails}
		size={92}
		class={[
			'size-24 shadow-md',
			item.kind === 'artist' && 'rounded-full',
			item.kind !== 'artist' && 'rounded-md'
		]}
	/>
	<div class="min-w-0">
		<h3 class="truncate text-3xl font-bold">{item.title}</h3>
		<p class="truncate text-sm text-dim">{item.subtitle}</p>
	</div>
	{#if item.track}
		<PlayOverlayButton
			{playing}
			label="Play {item.title}"
			onclick={playTrack}
			class={[
				'absolute right-5 bottom-5',
				!playing &&
					'translate-y-2 opacity-0 group-hover/card:translate-y-0 group-hover/card:opacity-100'
			]}
		/>
	{/if}
{/snippet}

{#if pageHref}
	<a
		href={pageHref}
		class="group/card duration-fast relative flex flex-col gap-5 rounded-lg bg-raised p-5 transition-colors hover:bg-hover"
		>{@render card()}</a
	>
{:else}
	<div
		class="group/card duration-fast relative flex flex-col gap-5 rounded-lg bg-raised p-5 transition-colors hover:bg-hover"
	>
		{@render card()}
	</div>
{/if}
