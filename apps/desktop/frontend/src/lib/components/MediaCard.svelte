<script lang="ts">
	import type { Item, ItemKind } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { itemHref } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';

	let { item, context = 'home' }: { item: Item; context?: string } = $props();

	const playing = $derived(isTrackPlaying());
	// Songs and episodes play in place; every other kind opens its page.
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
	<div class="relative">
		<Cover
			thumbnails={item.thumbnails}
			size={152}
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
					'absolute right-2 bottom-2',
					!playing &&
						'translate-y-2 opacity-0 group-hover/card:translate-y-0 group-hover/card:opacity-100'
				]}
			/>
		{/if}
	</div>
	<div class="min-w-0">
		<h3 class="truncate text-base font-medium">{item.title}</h3>
		<p class="line-clamp-2 text-sm text-dim">{item.subtitle}</p>
	</div>
{/snippet}

{#if pageHref}
	<a
		href={pageHref}
		class="group/card duration-fast flex w-44 shrink-0 flex-col gap-2 rounded-lg p-3 transition-colors hover:bg-hover"
		>{@render card()}</a
	>
{:else}
	<article
		class="group/card duration-fast flex w-44 shrink-0 flex-col gap-2 rounded-lg p-3 transition-colors hover:bg-hover"
	>
		{@render card()}
	</article>
{/if}
