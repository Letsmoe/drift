<script lang="ts">
	import { Button, Icon, Overlay, View, type KeyEvent } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import NeoworksMenu from '#lib/components/NeoworksMenu.svelte';
	import SearchField from '#lib/components/SearchField.svelte';
	import {
		caretLeftBold,
		caretRightBold,
		clockCounterClockwise,
		gearSix,
		house,
		houseFill,
		magnifyingGlass
	} from '#lib/layout/chromeIcons.ts';
	import TogetherMenu from '#lib/components/TogetherMenu.svelte';
	import { back, forward, goto, router } from '#lib/router.svelte.ts';
	import { SearchSuggestions } from '#lib/search/suggestions.svelte.ts';
	import { colors, floatingPanel, glassBackground, radii } from '#lib/theme.ts';

	const suggestions = new SearchSuggestions();
	let query = $state(initialQuery());

	const onHome = $derived(router.page === 'Home');
	const searchQuery = $derived(currentSearchQuery());
	const suggestionsVisible = $derived(suggestions.isOpen && suggestions.items.length > 0);

	// Keep the field in step with the results page when a chip or a link changes the query.
	$effect(() => {
		if (router.page === 'Search') {
			query = searchQuery;
		}
	});

	function currentSearchQuery(): string {
		const searchParameter = router.query.q;
		if (searchParameter === undefined) {
			return '';
		}
		return searchParameter;
	}

	function initialQuery(): string {
		if (router.page !== 'Search') {
			return '';
		}
		return currentSearchQuery();
	}

	function submitText(text: string): void {
		const trimmedText = text.trim();
		suggestions.close();
		if (trimmedText === '') {
			return;
		}
		query = trimmedText;
		goto(`/search?q=${encodeURIComponent(trimmedText)}`);
	}

	function submit(): void {
		const highlighted = suggestions.activeText();
		if (highlighted !== '') {
			submitText(highlighted);
			return;
		}
		submitText(query);
	}

	function handleKey(event: KeyEvent): void {
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

	// Alt with the arrow keys walks the history, as in a browser.
	function handleWindowKey(event: KeyboardEvent): void {
		if (!event.altKey) {
			return;
		}
		if (event.key === 'ArrowLeft') {
			event.preventDefault();
			back();
		} else if (event.key === 'ArrowRight') {
			event.preventDefault();
			forward();
		}
	}

	function suggestionBackground(index: number): string {
		if (index === suggestions.activeIndex) {
			return colors.hover;
		}
		return 'transparent';
	}

	function historyColor(enabled: boolean): string {
		if (enabled) {
			return colors.text;
		}
		return appearance.faint;
	}

	function homeColor(): string {
		if (onHome) {
			return colors.text;
		}
		return appearance.dim;
	}

	function homeIcon(): string {
		if (onHome) {
			return houseFill;
		}
		return house;
	}
</script>

<svelte:window onkeydown={handleWindowKey} />

<View
	style={{
		direction: 'row',
		height: 48,
		shrink: 0,
		align: 'center',
		justify: 'center',
		gap: 8,
		paddingX: 8
	}}
>
	<View
		style={{
			position: 'absolute',
			left: 8,
			top: 6,
			direction: 'row',
			align: 'center',
			gap: 8
		}}
	>
		<IconAction
			icon={caretLeftBold}
			iconSize={18}
			boxSize={36}
			color={historyColor(router.canGoBack)}
			hoverColor={colors.text}
			hoverBackground={colors.hover}
			disabled={!router.canGoBack}
			style={{ background: glassBackground }}
			onpress={back}
		/>
		<IconAction
			icon={caretRightBold}
			iconSize={18}
			boxSize={36}
			color={historyColor(router.canGoForward)}
			hoverColor={colors.text}
			hoverBackground={colors.hover}
			disabled={!router.canGoForward}
			style={{ background: glassBackground }}
			onpress={forward}
		/>
	</View>

	<IconAction
		icon={homeIcon()}
		iconSize={22}
		boxSize={48}
		color={homeColor()}
		hoverColor={colors.text}
		hoverBackground={colors.hover}
		style={{ background: glassBackground }}
		onpress={() => goto('/')}
	/>
	<View style={{ flex: 1, maxWidth: 448 }}>
		<SearchField
			bind:value={query}
			placeholder="What do you want to play?"
			onsubmit={submit}
			onfocus={() => suggestions.update(query)}
			oninput={() => suggestions.update(query)}
			onblur={() => suggestions.close()}
			onkey={handleKey}
		>
			{#if suggestionsVisible}
				<Overlay side="bottom" align="start" offset={8} matchWidth keepFocus style={floatingPanel}>
					{#each suggestions.items as suggestion, index (suggestion)}
						<Button
							style={{
								direction: 'row',
								align: 'center',
								gap: 12,
								paddingX: 12,
								paddingY: 8,
								radius: radii.medium,
								background: suggestionBackground(index)
							}}
							hover={{ background: colors.hover }}
							onpress={() => submitText(suggestion)}
						>
							<Icon
								svg={suggestions.showsRecent ? clockCounterClockwise : magnifyingGlass}
								size={16}
								color={appearance.dim}
								style={{ shrink: 0 }}
							/>
							<View style={{ flex: 1, minWidth: 0 }}>
								<Label tone="default" truncate>{suggestion}</Label>
							</View>
						</Button>
					{/each}
				</Overlay>
			{/if}
		</SearchField>
	</View>

	<View
		style={{
			position: 'absolute',
			right: 8,
			top: 6,
			direction: 'row',
			align: 'center',
			gap: 8
		}}
	>
		<TogetherMenu />
		<IconAction
			icon={gearSix}
			iconSize={20}
			boxSize={36}
			color={appearance.dim}
			hoverColor={colors.text}
			hoverBackground={colors.hover}
			style={{ background: glassBackground }}
			onpress={() => goto('/settings')}
		/>
		<NeoworksMenu />
	</View>
</View>
