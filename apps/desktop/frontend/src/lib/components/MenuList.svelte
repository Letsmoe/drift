<script lang="ts">
	import CaretRightIcon from 'phosphor-svelte/lib/CaretRightIcon';
	import type { Attachment } from 'svelte/attachments';
	import MenuList from '#lib/components/MenuList.svelte';
	import type { MenuEntry } from '#lib/menu/menuEntry.ts';

	let {
		entries,
		onclose,
		autofocus = false,
		onleave
	}: {
		entries: MenuEntry[];
		onclose: () => void;
		autofocus?: boolean;
		/** Called for ArrowLeft in a submenu so the parent can close it. */
		onleave?: () => void;
	} = $props();

	let listElement = $state<HTMLElement>();
	let openSubmenuIndex = $state(-1);
	let flipSubmenu = $state(false);

	const focusFirstItem: Attachment<HTMLElement> = (element) => {
		if (autofocus) {
			itemButtons(element)[0]?.focus();
		}
	};

	function itemButtons(list: HTMLElement): HTMLElement[] {
		return [...list.querySelectorAll<HTMLElement>(':scope > div > [role=menuitem]')];
	}

	function select(entry: MenuEntry) {
		onclose();
		entry.onselect?.();
	}

	function openSubmenu(index: number, trigger: HTMLElement) {
		const bounds = trigger.getBoundingClientRect();
		const submenuWidth = 260;
		flipSubmenu = bounds.right + submenuWidth > window.innerWidth;
		openSubmenuIndex = index;
	}

	function activate(index: number, event: MouseEvent) {
		const entry = entries[index];
		if (entry.children) {
			openSubmenu(index, event.currentTarget as HTMLElement);
			return;
		}
		select(entry);
	}

	function hover(index: number, event: MouseEvent) {
		if (entries[index].children) {
			openSubmenu(index, event.currentTarget as HTMLElement);
			return;
		}
		openSubmenuIndex = -1;
	}

	function moveFocus(step: number) {
		if (!listElement) {
			return;
		}
		const buttons = itemButtons(listElement);
		const currentIndex = buttons.indexOf(document.activeElement as HTMLElement);
		const nextIndex = (currentIndex + step + buttons.length) % buttons.length;
		buttons[nextIndex]?.focus();
	}

	function enterSubmenu(index: number, event: KeyboardEvent) {
		if (!entries[index].children) {
			return;
		}
		openSubmenu(index, event.currentTarget as HTMLElement);
		const submenu = (event.currentTarget as HTMLElement).parentElement?.querySelector(
			'[role=menu]'
		);
		queueMicrotask(() =>
			(submenu?.querySelector('[role=menuitem]') as HTMLElement | null)?.focus()
		);
	}

	function leaveSubmenu() {
		openSubmenuIndex = -1;
	}

	function handleKeydown(index: number, event: KeyboardEvent) {
		if (event.key === 'ArrowDown') {
			event.preventDefault();
			event.stopPropagation();
			moveFocus(1);
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			event.stopPropagation();
			moveFocus(-1);
		} else if (event.key === 'ArrowRight') {
			event.stopPropagation();
			enterSubmenu(index, event);
		} else if (event.key === 'ArrowLeft' && onleave) {
			event.stopPropagation();
			onleave();
		}
	}
</script>

<div
	bind:this={listElement}
	{@attach focusFirstItem}
	role="menu"
	class={[
		'min-w-52 rounded-lg border border-line bg-elevated p-1 text-sm shadow-lg',
		onleave && 'max-h-80 overflow-y-auto'
	]}
>
	{#each entries as entry, index (index)}
		<div class="relative">
			{#if entry.separatorBefore}
				<div class="my-1 h-px bg-line"></div>
			{/if}
			<button
				type="button"
				role="menuitem"
				onclick={(event) => activate(index, event)}
				onmouseenter={(event) => hover(index, event)}
				onkeydown={(event) => handleKeydown(index, event)}
				class="duration-fast flex w-full items-center gap-3 rounded-md px-3 py-2 text-left text-default transition-colors hover:bg-hover focus-visible:bg-hover focus-visible:outline-none"
			>
				{#if entry.icon}
					<entry.icon size={16} />
				{/if}
				<span class="flex-1 truncate">{entry.label}</span>
				{#if entry.children}
					<CaretRightIcon size={14} class="text-dim" />
				{/if}
			</button>
			{#if entry.children && openSubmenuIndex === index}
				<div
					class={[
						'absolute top-0 z-10',
						flipSubmenu && 'right-full pr-1',
						!flipSubmenu && 'left-full pl-1'
					]}
				>
					<MenuList entries={entry.children} {onclose} onleave={leaveSubmenu} />
				</div>
			{/if}
		</div>
	{/each}
</div>
