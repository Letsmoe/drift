<script lang="ts">
	import type { Track } from '#lib/api/music.ts';
	import ArtistLinks from '#lib/components/ArtistLinks.svelte';
	import { itemHref } from '#lib/format.ts';

	let { track }: { track: Track } = $props();

	const albumId = $derived(track.album?.id);
</script>

{#snippet title()}
	{#if albumId}
		<a href={itemHref('album', albumId)} class="hover:underline">{track.title}</a>
	{:else}
		{track.title}
	{/if}
{/snippet}

<p class="truncate text-base font-medium">{@render title()}</p>
<p class="truncate text-sm text-dim"><ArtistLinks artists={track.artists} /></p>
