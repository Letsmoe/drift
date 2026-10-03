<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import HouseIcon from 'phosphor-svelte/lib/HouseIcon';
	import SearchField from '#lib/components/SearchField.svelte';

	let query = $state(initialQuery());

	const onHome = $derived(page.url.pathname === '/');

	function initialQuery(): string {
		const urlQuery = page.url.searchParams.get('q');
		if (urlQuery === null) {
			return '';
		}
		return urlQuery;
	}

	function search() {
		const trimmedQuery = query.trim();
		if (trimmedQuery === '') {
			return;
		}
		goto(`/search?q=${encodeURIComponent(trimmedQuery)}`);
	}
</script>

<header class="flex h-12 shrink-0 items-center justify-center gap-2 px-2">
	<a
		href="/"
		aria-label="Home"
		aria-current={onHome ? 'page' : undefined}
		class="duration-fast flex size-12 shrink-0 items-center justify-center rounded-full bg-surface text-dim transition-colors hover:bg-hover hover:text-default aria-[current=page]:text-default"
	>
		<HouseIcon size={22} weight={onHome ? 'fill' : 'regular'} />
	</a>
	<div class="w-full max-w-md">
		<SearchField bind:value={query} placeholder="What do you want to play?" onsubmit={search} />
	</div>
</header>
