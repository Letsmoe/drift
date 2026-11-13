<script lang="ts">
	import { Button, Icon, View } from '@neoworks/svelte-gpui/ui';
	import type { Snippet } from 'svelte';
	import type { Thumbnail } from '#lib/api/music.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Cover from '#lib/components/Cover.svelte';
	import Label from '#lib/components/Label.svelte';
	import { speakerHighFill } from '#lib/icons.ts';
	import { goto, router } from '#lib/router.svelte.ts';
	import { isContextCurrent, isContextPlaying } from '#lib/player/contextPlayback.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	const coverSize = 48;

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

	const active = $derived(router.href === href || router.path === href);
	const current = $derived(isContextCurrent(context));
	const coverRadius = $derived(pickCoverRadius());

	function pickCoverRadius(): number {
		if (round) {
			return radii.full;
		}
		return radii.medium;
	}

	function pickBackground(): string {
		if (active) {
			return colors.raised;
		}
		return 'transparent';
	}
</script>

<Button
	style={{
		direction: 'row',
		align: 'center',
		gap: 12,
		padding: 8,
		radius: radii.medium,
		background: pickBackground(),
		transition: hoverTransition
	}}
	hover={{ background: colors.hover }}
	onpress={() => goto(href)}
>
	{#if cover}
		{@render cover()}
	{:else}
		<View style={{ size: coverSize, shrink: 0, radius: coverRadius, overflow: 'hidden' }}>
			<Cover {thumbnails} size={coverSize} style={{ size: coverSize }} />
		</View>
	{/if}
	<View style={{ flex: 1, minWidth: 0 }}>
		<Label size="medium" weight="medium" tone="default" accented={current} truncate>{title}</Label>
		<Label size="small" tone="dim" truncate>{subtitle}</Label>
	</View>
	{#if isContextPlaying(context)}
		<Icon svg={speakerHighFill} size={16} color={appearance.accent} style={{ shrink: 0 }} />
	{/if}
</Button>
