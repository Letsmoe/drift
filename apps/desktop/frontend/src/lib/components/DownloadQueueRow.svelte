<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { DownloadItem, DownloadStatus } from '#lib/api/downloads.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Cover from '#lib/components/Cover.svelte';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { downloadFraction, formatBytes } from '#lib/downloads/downloadProgress.ts';
	import { formatArtists } from '#lib/format.ts';
	import { arrowClockwise, x } from '#lib/icons.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	let { item }: { item: DownloadItem } = $props();

	const videoId = $derived(item.track.videoId);
	// The live status from events is fresher than the item loaded with the list.
	const status = $derived(currentStatus());
	const failed = $derived(status.state === 'failed');
	const percent = $derived(Math.round(downloadFraction(status) * 100));

	function currentStatus(): DownloadStatus {
		const live = downloads.statusOf(videoId);
		if (live) {
			return live;
		}
		return item.status;
	}

	function stateTone(): 'danger' | 'dim' {
		if (failed) {
			return 'danger';
		}
		return 'dim';
	}

	function describeState(): string {
		switch (status.state) {
			case 'downloading':
				return `${formatBytes(status.bytesDone)} of ${formatBytes(status.bytesTotal)}`;
			case 'paused':
				return 'Paused';
			case 'failed':
				return status.error;
			default:
				return 'Waiting';
		}
	}
</script>

<View
	style={{
		direction: 'row',
		align: 'center',
		gap: 12,
		paddingX: 16,
		paddingY: 8,
		radius: radii.medium,
		transition: hoverTransition
	}}
	hover={{ background: colors.hover }}
>
	<Cover
		thumbnails={item.track.thumbnails}
		size={40}
		style={{ size: 40, shrink: 0, radius: radii.small }}
	/>
	<View style={{ flex: 1, minWidth: 0 }}>
		<Label size="body" truncate>{item.track.title}</Label>
		<Label size="small" tone="dim" truncate>{formatArtists(item.track.artists)}</Label>
	</View>
	<View style={{ width: 224, gap: 4 }}>
		<Label size="tiny" tone={stateTone()} truncate style={{ textAlign: 'right' }}>
			{describeState()}
		</Label>
		{#if status.state === 'downloading'}
			<View
				style={{ height: 4, radius: radii.full, background: colors.raised, overflow: 'hidden' }}
			>
				<View style={{ width: `${percent}%`, height: 4, background: appearance.accent }} />
			</View>
		{/if}
	</View>
	{#if failed}
		<IconAction
			icon={arrowClockwise}
			iconSize={18}
			color={appearance.dim}
			onpress={() => downloads.retry(videoId)}
		/>
	{/if}
	<IconAction
		icon={x}
		iconSize={18}
		color={appearance.dim}
		onpress={() => downloads.cancel(videoId)}
	/>
</View>
