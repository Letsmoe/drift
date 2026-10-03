<script lang="ts">
	import { LoadingSpinner, SectionHeader } from '@neoworks-dev/ui';
	import SearchField from '#lib/components/SearchField.svelte';
	import TrackRow from '#lib/components/TrackRow.svelte';
	import { SongSearch } from '#lib/search/songSearch.svelte.ts';

	const search = new SongSearch();
</script>

<header class="border-b border-line-faint px-6 py-4">
	<SearchField
		bind:value={search.query}
		placeholder="Search songs"
		onsubmit={() => search.submit()}
	/>
</header>

<main class="flex-1 overflow-y-auto px-6 py-4">
	{#if search.status === 'loading'}
		<div class="flex justify-center py-12"><LoadingSpinner /></div>
	{:else if search.status === 'failed'}
		<p class="py-12 text-center text-sm text-red">{search.errorMessage}</p>
	{:else if search.status === 'done' && search.tracks.length === 0}
		<p class="py-12 text-center text-sm text-dim">No songs found.</p>
	{:else if search.status === 'done'}
		<SectionHeader title="Songs" />
		<ul class="flex flex-col gap-1">
			{#each search.tracks as track (track.videoId)}
				<li><TrackRow {track} /></li>
			{/each}
		</ul>
	{:else}
		<p class="py-12 text-center text-sm text-dim">Search YouTube Music to get started.</p>
	{/if}
</main>
