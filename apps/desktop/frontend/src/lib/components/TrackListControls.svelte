<script lang="ts">
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import { trackSorts, type TrackListFilter } from '#lib/pages/trackListFilter.svelte.ts';

	let { filter }: { filter: TrackListFilter } = $props();
</script>

<div class="flex items-center justify-between gap-3 pb-2">
	<label
		class="flex h-9 w-64 items-center gap-2 rounded-full bg-raised px-3 text-dim focus-within:text-default"
	>
		<MagnifyingGlassIcon size={16} class="shrink-0" />
		<input
			type="search"
			bind:value={filter.query}
			placeholder="Search in playlist"
			class="min-w-0 flex-1 bg-transparent text-sm text-default outline-none placeholder:text-dim"
		/>
	</label>
	<label class="flex items-center gap-2 text-sm text-dim">
		Sort by
		<select
			bind:value={filter.sort}
			class="h-9 rounded-full bg-raised px-3 text-sm text-default outline-none hover:bg-hover"
		>
			{#each trackSorts as option (option.value)}
				<option value={option.value}>{option.label}</option>
			{/each}
		</select>
	</label>
</div>
