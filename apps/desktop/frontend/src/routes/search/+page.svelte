<script lang="ts">
	import { page } from '$app/state';
	import { LoadingSpinner } from '@neoworks-dev/ui';
	import TrackRow from '#lib/components/TrackRow.svelte';
	import { SongSearch } from '#lib/search/songSearch.svelte.ts';

	const search = new SongSearch();
	const query = $derived(page.url.searchParams.get('q'));

	$effect(() => {
		if (query !== null) {
			search.run(query);
		}
	});
</script>

<h1 class="px-3 pb-4 text-2xl font-bold">Songs</h1>

{#if search.status === 'loading'}
	<div class="flex justify-center py-12"><LoadingSpinner /></div>
{:else if search.status === 'failed'}
	<p class="py-12 text-center text-sm text-red">{search.errorMessage}</p>
{:else if search.status === 'done' && search.tracks.length === 0}
	<p class="py-12 text-center text-sm text-dim">No songs found.</p>
{:else if search.status === 'done'}
	<ul class="flex flex-col gap-1">
		{#each search.tracks as track (track.videoId)}
			<li><TrackRow {track} /></li>
		{/each}
	</ul>
{/if}
