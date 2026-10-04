<script lang="ts">
	import PauseIcon from 'phosphor-svelte/lib/PauseIcon';
	import PlayIcon from 'phosphor-svelte/lib/PlayIcon';
	import type { ClassValue } from 'svelte/elements';

	let {
		playing,
		label,
		size = 'md',
		class: className = '',
		onclick
	}: {
		playing: boolean;
		label: string;
		size?: 'sm' | 'md' | 'lg';
		class?: ClassValue;
		onclick: (event: MouseEvent) => void;
	} = $props();

	const sizeClasses = { sm: 'size-8', md: 'size-12', lg: 'size-14' };
	const iconSizes = { sm: 14, md: 20, lg: 24 };
</script>

<button
	type="button"
	aria-label={label}
	{onclick}
	class={[
		'duration-fast flex items-center justify-center rounded-full bg-action text-action-fg shadow-md transition-[transform,opacity] hover:scale-105',
		sizeClasses[size],
		className
	]}
>
	{#if playing}
		<PauseIcon size={iconSizes[size]} weight="fill" />
	{:else}
		<PlayIcon size={iconSizes[size]} weight="fill" />
	{/if}
</button>
