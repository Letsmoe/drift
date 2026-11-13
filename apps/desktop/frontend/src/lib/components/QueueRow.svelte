<script lang="ts">
	import { Button, Icon, View } from '@neoworks/svelte-gpui/ui';
	import type { QueueEntry } from '#lib/api/queue.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Cover from '#lib/components/Cover.svelte';
	import DownloadedIndicator from '#lib/components/DownloadedIndicator.svelte';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { formatArtists } from '#lib/format.ts';
	import { playFill, sparkleFill, x } from '#lib/layout/chromeIcons.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	let {
		entry,
		current = false,
		played = false,
		size = 'compact'
	}: {
		entry: QueueEntry;
		current?: boolean;
		/** Already played: shown dimmed, above the current entry. */
		played?: boolean;
		/** Compact rows fit the side panel; comfortable ones the full-screen player. */
		size?: 'compact' | 'comfortable';
	} = $props();

	let hovered = $state(false);

	const comfortable = $derived(size === 'comfortable');
	const coverSize = $derived(pickCoverSize());
	const rowOpacity = $derived(pickOpacity());

	$effect(() => {
		downloads.ensureStates([entry.track.videoId]);
	});

	function pickCoverSize(): number {
		if (comfortable) {
			return 48;
		}
		return 32;
	}

	function pickOpacity(): number {
		if (played && !hovered) {
			return 0.6;
		}
		return 1;
	}

	// Compact rows are tighter, as the old side panel's were.
	function pickRowPadding(): number {
		if (comfortable) {
			return 8;
		}
		return 4;
	}

	function pickGap(): number {
		if (comfortable) {
			return 12;
		}
		return 10;
	}

	function pickCoverRadius(): number {
		if (comfortable) {
			return radii.medium;
		}
		return 4;
	}

	function pickTitleSize(): 'medium' | 'small' {
		if (comfortable) {
			return 'medium';
		}
		return 'small';
	}

	function playEntry(): void {
		if (!current) {
			player.jump(entry.id);
		}
	}
</script>

<Button
	style={{
		direction: 'row',
		align: 'center',
		gap: pickGap(),
		padding: pickRowPadding(),
		radius: radii.medium,
		opacity: rowOpacity,
		transition: hoverTransition
	}}
	hover={{ background: colors.hover }}
	onhover={(isHovered) => (hovered = isHovered)}
	onpress={playEntry}
>
	<View style={{ size: coverSize, shrink: 0, radius: pickCoverRadius(), overflow: 'hidden' }}>
		<Cover thumbnails={entry.track.thumbnails} size={coverSize} style={{ size: coverSize }} />
	</View>
	<View style={{ flex: 1, minWidth: 0 }}>
		<View style={{ direction: 'row', align: 'center', gap: 6, minWidth: 0 }}>
			<View style={{ flex: 1, minWidth: 0 }}>
				<Label size={pickTitleSize()} weight="medium" tone="default" accented={current} truncate>
					{entry.track.title}
				</Label>
			</View>
			{#if entry.magic}
				<Icon svg={sparkleFill} size={12} color={appearance.accent} style={{ shrink: 0 }} />
			{/if}
			<DownloadedIndicator videoId={entry.track.videoId} size={12} />
		</View>
		<Label size="tiny" tone="dim" truncate>{formatArtists(entry.track.artists)}</Label>
	</View>
	{#if !current && hovered}
		<View style={{ direction: 'row', align: 'center', shrink: 0 }}>
			<IconAction icon={playFill} iconSize={14} boxSize={24} onpress={playEntry} />
			<IconAction icon={x} iconSize={14} boxSize={24} onpress={() => player.remove(entry.id)} />
		</View>
	{/if}
</Button>
