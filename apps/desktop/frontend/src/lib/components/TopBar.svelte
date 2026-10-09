<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import CaretLeftIcon from 'phosphor-svelte/lib/CaretLeftIcon';
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';
	import ClockCounterClockwiseIcon from 'phosphor-svelte/lib/ClockCounterClockwiseIcon';
	import HouseIcon from 'phosphor-svelte/lib/HouseIcon';
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import SearchField from '#lib/components/SearchField.svelte';
	import TogetherMenu from '#lib/components/TogetherMenu.svelte';
	import { NavigationHistory } from '#lib/layout/navigationHistory.svelte.ts';
	import { SearchSuggestions } from '#lib/search/suggestions.svelte.ts';

	// MouseEvent.button numbers of the side buttons on most mice.
	const mouseBackButton = 3;
	const mouseForwardButton = 4;

	const navigationHistory = new NavigationHistory();

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

	function navigateWithMouseButtons(event: MouseEvent) {
		if (event.button === mouseBackButton) {
			event.preventDefault();
			navigationHistory.back();
		} else if (event.button === mouseForwardButton) {
			event.preventDefault();
			navigationHistory.forward();
		}
	}

	function navigateWithKeys(event: KeyboardEvent) {
		if (!event.altKey) {
			return;
		}
		if (event.key === 'ArrowLeft') {
			event.preventDefault();
			navigationHistory.back();
		} else if (event.key === 'ArrowRight') {
			event.preventDefault();
			navigationHistory.forward();
		}
	}
</script>

<svelte:window onmouseup={navigateWithMouseButtons} onkeydown={navigateWithKeys} />

{#snippet historyButton(label: string, enabled: boolean, onclick: () => void, forward: boolean)}
	<button
		type="button"
		aria-label={label}
		title={label}
		disabled={!enabled}
		{onclick}
		class="duration-fast glass flex size-9 items-center justify-center rounded-full text-default transition-colors hover:bg-hover disabled:cursor-default disabled:text-faint disabled:hover:bg-transparent"
	>
		{#if forward}
			<CaretRightIcon size={18} weight="bold" />
		{:else}
			<CaretLeftIcon size={18} weight="bold" />
		{/if}
	</button>
{/snippet}

<header class="relative flex h-12 shrink-0 items-center justify-center gap-2 px-2">
	<nav aria-label="History" class="absolute left-2 flex items-center gap-2">
		{@render historyButton(
			'Back',
			navigationHistory.canGoBack,
			() => navigationHistory.back(),
			false
		)}
		{@render historyButton(
			'Forward',
			navigationHistory.canGoForward,
			() => navigationHistory.forward(),
			true
		)}
	</nav>
	<a
		href="/"
		aria-label="Home"
		aria-current={onHome ? 'page' : undefined}
		class="duration-fast glass flex size-12 shrink-0 items-center justify-center rounded-full text-dim transition-colors hover:bg-hover hover:text-default aria-[current=page]:text-default"
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
	<div class="absolute right-2 flex items-center gap-2">
		<TogetherMenu />
	</div>
</header>
