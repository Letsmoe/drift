<script lang="ts">
	import type { Component } from 'svelte';

	let {
		icon: Icon,
		label,
		active,
		onclick
	}: {
		icon: Component<{ size?: number; weight?: 'regular' | 'bold' }>;
		label: string;
		active: boolean;
		onclick: () => void;
	} = $props();

	const weight = $derived(iconWeight());

	function iconWeight(): 'regular' | 'bold' {
		if (active) {
			return 'bold';
		}
		return 'regular';
	}
</script>

<button
	type="button"
	aria-label={label}
	aria-pressed={active}
	title={label}
	{onclick}
	class={[
		'duration-fast relative flex size-8 cursor-pointer items-center justify-center rounded-full transition-colors hover:bg-hover',
		active && 'text-action',
		!active && 'text-muted hover:text-default'
	]}
>
	<Icon size={18} {weight} />
	{#if active}
		<span class="absolute bottom-0.5 size-1 rounded-full bg-action"></span>
	{/if}
</button>
