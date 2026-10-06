<script lang="ts">
	import type { Artist } from '#lib/api/music.ts';
	import { createItem } from '#lib/api/music.ts';
	import { artistInfo } from '#lib/artist/artistInfo.svelte.ts';
	import Cover from '#lib/components/Cover.svelte';
	import FollowButton from '#lib/components/FollowButton.svelte';
	import { itemHref } from '#lib/format.ts';

	let { artist }: { artist: Artist } = $props();

	const artistPage = $derived(artistInfo.pageOf(artist.id));

	$effect(() => {
		artistInfo.load(artist.id);
	});
</script>

{#if artistPage}
	<section class="flex flex-col gap-3 rounded-xl bg-raised/60 p-3">
		<h3 class="text-sm font-semibold text-muted">About the artist</h3>
		<div class="flex items-center gap-3">
			<Cover thumbnails={artistPage.thumbnails} size={56} class="size-14 shrink-0 rounded-full" />
			<div class="flex min-w-0 flex-1 flex-col items-start gap-1.5">
				<a href={itemHref('artist', artist.id)} class="truncate font-semibold hover:underline">
					{artistPage.name}
				</a>
				<FollowButton
					item={createItem('artist', artist.id, artistPage.name, '', artistPage.thumbnails)}
				/>
			</div>
		</div>
		{#if artistPage.description}
			<p class="line-clamp-4 text-xs text-muted">{artistPage.description}</p>
		{/if}
	</section>
{/if}
