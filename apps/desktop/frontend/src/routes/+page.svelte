<script lang="ts">
	import { LoadingSpinner } from '@neoworks-dev/ui';
	import ShelfRow from '#lib/components/ShelfRow.svelte';
	import { HomeFeed } from '#lib/home/homeFeed.svelte.ts';
	import { whenVisible } from '#lib/whenVisible.ts';

	const feed = new HomeFeed();
	feed.load();
</script>

{#if feed.status === 'loading'}
	<div class="flex justify-center py-24"><LoadingSpinner /></div>
{:else if feed.status === 'failed'}
	<p class="py-24 text-center text-sm text-red">{feed.errorMessage}</p>
{:else}
	<div class="flex flex-col gap-8">
		{#each feed.shelves as shelf, index (index)}
			<ShelfRow {shelf} />
		{/each}
	</div>
	{#if feed.hasMore}
		<!-- Re-created per page so it fires again when it is still in view after loading. -->
		{#key feed.shelves.length}
			<div class="flex justify-center py-8" {@attach whenVisible(() => feed.loadMore())}>
				<LoadingSpinner />
			</div>
		{/key}
	{/if}
{/if}
