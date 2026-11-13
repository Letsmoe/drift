<!--
  The rows of a menu on a floating card, with submenus that open to the side. Arrow keys move the highlight,
  Right enters a submenu, Left leaves it and Enter chooses. The host's own Menu has no submenus, which the
  track menu needs, so this one is written here.
-->
<script lang="ts">
	import { Icon, Overlay, Scroll, View, type KeyEvent } from '@neoworks/svelte-gpui/ui';
	import Label from '#lib/components/Label.svelte';
	import MenuList from '#lib/components/MenuList.svelte';
	import { caretRight } from '#lib/icons.ts';
	import type { MenuEntry } from '#lib/menu/menuEntry.ts';
	import { colors, floatingPanel, hoverTransition, radii } from '#lib/theme.ts';

	const submenuMaxHeight = 320;

	let {
		entries,
		onclose,
		autofocus = false,
		onleave
	}: {
		entries: MenuEntry[];
		onclose: () => void;
		autofocus?: boolean;
		/** Called for Left in a submenu so the parent can close it. */
		onleave?: () => void;
	} = $props();

	let root: ReturnType<typeof View> | undefined = $state();
	let activeIndex = $state(-1);
	let openSubmenuIndex = $state(-1);

	$effect(() => {
		if (autofocus) {
			root?.focus();
		}
	});

	function select(entry: MenuEntry): void {
		onclose();
		entry.onselect?.();
	}

	function activate(index: number): void {
		const entry = entries[index];
		if (entry === undefined) {
			return;
		}
		if (entry.children) {
			openSubmenuIndex = index;
			return;
		}
		select(entry);
	}

	function hover(index: number, hovered: boolean): void {
		if (!hovered) {
			return;
		}
		activeIndex = index;
		if (entries[index].children) {
			openSubmenuIndex = index;
			return;
		}
		openSubmenuIndex = -1;
	}

	function moveActive(step: number): void {
		openSubmenuIndex = -1;
		activeIndex = (activeIndex + step + entries.length) % entries.length;
	}

	function leaveSubmenu(): void {
		openSubmenuIndex = -1;
		root?.focus();
	}

	function handleKey(event: KeyEvent): void {
		if (event.key === 'ArrowDown') {
			event.preventDefault();
			moveActive(1);
			return;
		}
		if (event.key === 'ArrowUp') {
			event.preventDefault();
			moveActive(-1);
			return;
		}
		if (event.key === 'ArrowLeft' && onleave) {
			event.preventDefault();
			event.stopPropagation();
			onleave();
			return;
		}
		if (event.key === 'ArrowRight' || event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			activate(activeIndex);
		}
	}

	function rowBackground(index: number): string {
		if (activeIndex === index) {
			return colors.hover;
		}
		return 'transparent';
	}
</script>

{#snippet rows()}
	{#each entries as entry, index (index)}
		{#if entry.separatorBefore}
			<View style={{ height: 1, marginY: 4, background: colors.line, shrink: 0 }} />
		{/if}
		<View
			onpress={() => activate(index)}
			onhover={(hovered) => hover(index, hovered)}
			style={{
				direction: 'row',
				align: 'center',
				gap: 12,
				paddingX: 12,
				paddingY: 8,
				radius: radii.medium,
				shrink: 0,
				cursor: 'pointer',
				transition: hoverTransition,
				background: rowBackground(index)
			}}
		>
			{#if entry.icon}
				<Icon svg={entry.icon} size={16} color={colors.text} />
			{/if}
			<Label size="small" tone="default" style={{ flex: 1, nowrap: true }}>{entry.label}</Label>
			{#if entry.children}
				<Icon svg={caretRight} size={14} color={colors.dim} />
			{/if}
			{#if entry.children && openSubmenuIndex === index}
				<Overlay side="right" align="start" offset={4}>
					<MenuList entries={entry.children} {onclose} onleave={leaveSubmenu} autofocus />
				</Overlay>
			{/if}
		</View>
	{/each}
{/snippet}

<View bind:this={root} focusable onkey={handleKey} style={{ ...floatingPanel, minWidth: 182 }}>
	{#if onleave}
		<Scroll style={{ maxHeight: submenuMaxHeight }}>
			{@render rows()}
		</Scroll>
	{:else}
		{@render rows()}
	{/if}
</View>
