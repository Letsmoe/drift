<script lang="ts">
	import { page } from '$app/state';
	import SpeakerHighIcon from 'phosphor-svelte/lib/SpeakerHighIcon';
	import type { Snippet } from 'svelte';
	import type { Thumbnail } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import { isContextCurrent, isContextPlaying } from '#lib/player/contextPlayback.ts';

	let {
		href,
		title,
		subtitle,
		context,
		thumbnails = [],
		round = false,
		cover = undefined
	}: {
		href: string;
		title: string;
		subtitle: string;
		context: string;
		thumbnails?: Thumbnail[];
		round?: boolean;
		/** Replaces the cover image, e.g. the Liked Songs tile. */
		cover?: Snippet;
	} = $props();

	const active = $derived(page.url.pathname === href);
	const current = $derived(isContextCurrent(context));
</script>

<a
	{href}
	aria-current={active ? 'page' : undefined}
	class="duration-fast flex items-center gap-3 rounded-md p-2 transition-colors hover:bg-hover aria-[current=page]:bg-raised"
>
	{#if cover}
		{@render cover()}
	{:else}
		<Cover
			{thumbnails}
			size={48}
			class={['size-12 shrink-0', round && 'rounded-full', !round && 'rounded-md']}
		/>
	{/if}
	<div class="min-w-0 flex-1">
		<p class={['truncate text-base font-medium', current && 'text-green']}>{title}</p>
		<p class="truncate text-sm text-dim">{subtitle}</p>
	</div>
	{#if isContextPlaying(context)}
		<SpeakerHighIcon size={16} weight="fill" class="shrink-0 text-green" />
	{/if}
</a>
