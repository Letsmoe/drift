<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { Episode } from '#lib/api/music.ts';
	import Cover from '#lib/components/Cover.svelte';
	import Label from '#lib/components/Label.svelte';
	import PlayOverlayButton from '#lib/components/PlayOverlayButton.svelte';
	import { formatDuration, playCountOf } from '#lib/format.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	const coverSide = 96;

	let { episode, onplay }: { episode: Episode; onplay: () => void } = $props();

	const track = $derived(episode.track);
	const current = $derived(player.isCurrent(track));
	const playing = $derived(player.isPlaying(track));
	const details = $derived(
		[episode.published, formatDuration(track.durationSeconds), playCountOf(track)]
			.filter((part) => part !== '')
			.join(' • ')
	);
</script>

<View
	style={{
		direction: 'row',
		width: 'full',
		gap: 16,
		padding: 16,
		radius: radii.large,
		transition: hoverTransition
	}}
	hover={{ background: colors.hover }}
>
	<Cover
		thumbnails={track.thumbnails}
		size={coverSide}
		style={{ size: coverSide, shrink: 0, radius: radii.medium }}
	/>
	<View style={{ flex: 1, minWidth: 0, gap: 4, overflow: 'hidden' }}>
		<Label size="body" weight="bold" accented={current} truncate>{track.title}</Label>
		<Label size="small" tone="dim" lines={2}>{episode.description}</Label>
		<View style={{ direction: 'row', align: 'center', gap: 12, paddingTop: 8 }}>
			<PlayOverlayButton size="sm" {playing} onpress={onplay} />
			<Label size="small" tone="dim" truncate style={{ flex: 1, minWidth: 0 }}>{details}</Label>
		</View>
	</View>
</View>
