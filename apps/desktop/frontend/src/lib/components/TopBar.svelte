<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ClockCounterClockwiseIcon from 'phosphor-svelte/lib/ClockCounterClockwiseIcon';
	import HouseIcon from 'phosphor-svelte/lib/HouseIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import SearchField from '#lib/components/SearchField.svelte';
	import { SearchSuggestions } from '#lib/search/suggestions.svelte.ts';

	const suggestions = new SearchSuggestions();
	let query = $state(initialQuery());

	const onHome = $derived(page.url.pathname === '/');
	const urlQuery = $derived(currentUrlQuery());

	// Keep the field in step with the results page when a chip or a link changes the query.
	$effect(() => {
		if (page.url.pathname === '/search') {
			query = urlQuery;
		}
	});

	function currentUrlQuery(): string {
		const searchParameter = page.url.searchParams.get('q');
		if (searchParameter === null) {
			return '';
		}
		return searchParameter;
	}

	function initialQuery(): string {
		if (page.url.pathname !== '/search') {
			return '';
		}
		return currentUrlQuery();
	}

	function submitText(text: string) {
		const trimmedText = text.trim();
		suggestions.close();
		if (trimmedText === '') {
			return;
		}
		query = trimmedText;
		goto(`/search?q=${encodeURIComponent(trimmedText)}`);
	}

	function submit() {
		const highlighted = suggestions.activeText();
		if (highlighted !== '') {
			submitText(highlighted);
			return;
		}
		submitText(query);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'ArrowDown') {
			event.preventDefault();
			suggestions.move(1);
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			suggestions.move(-1);
		} else if (event.key === 'Escape') {
			suggestions.close();
		}
	}

	function closeWhenFocusLeaves(event: FocusEvent) {
		const container = event.currentTarget as HTMLElement;
		if (!container.contains(event.relatedTarget as Node | null)) {
			suggestions.close();
		}
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
	<div class="relative w-full max-w-md" onfocusout={closeWhenFocusLeaves}>
		<SearchField
			bind:value={query}
			placeholder="What do you want to play?"
			expanded={suggestions.isOpen && suggestions.items.length > 0}
			onsubmit={submit}
			onfocus={() => suggestions.update(query)}
			oninput={() => suggestions.update(query)}
			onkeydown={handleKeydown}
		/>
		{#if suggestions.isOpen && suggestions.items.length > 0}
			<ul
				id="search-suggestions"
				role="listbox"
				class="absolute top-full z-40 mt-2 w-full overflow-hidden rounded-lg border border-line bg-elevated p-1 shadow-lg"
				onpointerdown={(event) => event.preventDefault()}
			>
				{#each suggestions.items as suggestion, index (suggestion)}
					<li role="option" aria-selected={index === suggestions.activeIndex}>
						<button
							type="button"
							onclick={() => submitText(suggestion)}
							class={[
								'flex w-full items-center gap-3 rounded-md px-3 py-2 text-left text-sm',
								index === suggestions.activeIndex && 'bg-hover',
								'hover:bg-hover'
							]}
						>
							{#if suggestions.showsRecent}
								<ClockCounterClockwiseIcon size={16} class="shrink-0 text-dim" />
							{:else}
								<MagnifyingGlassIcon size={16} class="shrink-0 text-dim" />
							{/if}
							<span class="truncate">{suggestion}</span>
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</header>
