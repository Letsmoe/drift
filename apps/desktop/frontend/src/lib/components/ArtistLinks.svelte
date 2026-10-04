<script lang="ts">
	import type { Artist } from '#lib/api/music.ts';
	import { itemHref } from '#lib/format.ts';

	let { artists }: { artists: Artist[] } = $props();
</script>

{#each artists as artist, index (index)}
	{#if index > 0}<span>, </span>{/if}
	{#if artist.id}
		<a href={itemHref('artist', artist.id)} class="hover:text-default hover:underline">
			{artist.name}
		</a>
	{:else}
		<span>{artist.name}</span>
	{/if}
{/each}
