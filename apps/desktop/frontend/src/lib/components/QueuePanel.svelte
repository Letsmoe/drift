<script lang="ts">
	import { Button, Icon, Scroll, View } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Cover from '#lib/components/Cover.svelte';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import QueueResizeHandle from '#lib/components/QueueResizeHandle.svelte';
	import UpcomingQueue from '#lib/components/UpcomingQueue.svelte';
	import { artTint } from '#lib/layout/artTint.ts';
	import { broom, sparkle } from '#lib/layout/chromeIcons.ts';
	import { panels } from '#lib/layout/panels.svelte.ts';
	import TrackCreditsLine from '#lib/layout/TrackCreditsLine.svelte';
	import { player } from '#lib/player/player.svelte.ts';
	import { splitQueue } from '#lib/queue/queueSections.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	const heroHeight = 224;

	const sections = $derived(splitQueue(player.queue));

	function autoplayButtonLabel(): string {
		if (player.queue.autoplay) {
			return 'Autoplay on';
		}
		return 'Autoplay off';
	}

	function autoplayColor(): string {
		if (player.queue.autoplay) {
			return appearance.accent;
		}
		return appearance.dim;
	}
</script>

<View
	style={{
		width: panels.queueWidth,
		shrink: 0,
		minHeight: 0,
		radius: radii.extraLarge,
		background: artTint(0, 0.32),
		transition: hoverTransition
	}}
>
	<QueueResizeHandle />
	<View
		style={{
			direction: 'row',
			align: 'center',
			justify: 'between',
			height: 48,
			shrink: 0,
			gap: 8,
			paddingX: 16
		}}
	>
		<Label size="medium" weight="semibold" tone="default">Now playing</Label>
		<View style={{ direction: 'row', align: 'center', gap: 4 }}>
			<Button
				style={{
					direction: 'row',
					align: 'center',
					gap: 4,
					paddingX: 8,
					paddingY: 4,
					radius: radii.full,
					color: autoplayColor(),
					transition: hoverTransition
				}}
				hover={{ background: colors.hover }}
				onpress={() => player.toggleAutoplay()}
			>
				<Icon svg={sparkle} size={14} />
				<Label size="tiny">{autoplayButtonLabel()}</Label>
			</Button>
			<IconAction
				icon={broom}
				iconSize={16}
				boxSize={28}
				hoverBackground={colors.hover}
				disabled={player.queue.entries.length === 0}
				onpress={() => player.clearQueue()}
			/>
		</View>
	</View>

	<Scroll style={{ flex: 1, minHeight: 0 }}>
		<View style={{ gap: 24, paddingX: 16, paddingBottom: 16 }}>
			{#if sections.nowPlaying}
				<Button
					style={{
						height: heroHeight,
						shrink: 0,
						radius: radii.extraLarge,
						overflow: 'hidden',
						shadow: { y: 8, blur: 24, color: 'rgba(0, 0, 0, 0.4)' }
					}}
					onpress={() => panels.toggleFullPlayer()}
				>
					<Cover thumbnails={sections.nowPlaying.track.thumbnails} size={panels.queueWidth} fill />
					<View
						style={{
							position: 'absolute',
							left: 0,
							right: 0,
							bottom: 0,
							paddingX: 16,
							paddingTop: 24,
							paddingBottom: 16,
							background: 'rgba(0, 0, 0, 0.55)'
						}}
					>
						<TrackCreditsLine
							track={sections.nowPlaying.track}
							titleSize="xlarge"
							artistSize="body"
						/>
					</View>
				</Button>
			{/if}

			<View style={{ gap: 16, marginX: -6 }}>
				<UpcomingQueue />
			</View>
		</View>
	</Scroll>
</View>
