<!-- The full-window now playing view: cover, controls, and lyrics or the queue beside them. Esc closes it. -->
<script lang="ts">
	import { window } from '@neoworks/svelte-gpui';
	import { Button, Scroll, View } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import ArtBackdrop from '#lib/components/ArtBackdrop.svelte';
	import ArtistLinks from '#lib/components/ArtistLinks.svelte';
	import Cover from '#lib/components/Cover.svelte';
	import FullPlayerControls from '#lib/components/FullPlayerControls.svelte';
	import FullPlayerLyrics from '#lib/components/FullPlayerLyrics.svelte';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import UpcomingQueue from '#lib/components/UpcomingQueue.svelte';
	import { caretDownBold } from '#lib/layout/chromeIcons.ts';
	import { panels } from '#lib/layout/panels.svelte.ts';
	import { lyricsView } from '#lib/lyrics/lyricsView.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	type Tab = 'lyrics' | 'queue';
	const tabs: { id: Tab; label: string }[] = [
		{ id: 'lyrics', label: 'Lyrics' },
		{ id: 'queue', label: 'Up next' }
	];

	const maximumCoverSide = 512;
	const maximumCoverShareOfHeight = 0.5;

	let activeTab = $state<Tab>('lyrics');
	let leftColumnWidth = $state(0);

	// The cover is square, at most 32rem, and leaves room for the controls below it.
	const coverSide = $derived(
		Math.round(
			Math.max(
				120,
				Math.min(maximumCoverSide, leftColumnWidth, window.size.height * maximumCoverShareOfHeight)
			)
		)
	);

	$effect(() => {
		if (player.track) {
			lyricsView.load(player.track);
		}
	});

	function handleKeydown(event: KeyboardEvent): void {
		if (event.key === 'Escape') {
			panels.closeFullPlayer();
		}
	}

	function tabBackground(tab: Tab): string {
		if (activeTab === tab) {
			return appearance.accent;
		}
		return 'transparent';
	}

	function tabColor(tab: Tab): string {
		if (activeTab === tab) {
			return appearance.onAccent;
		}
		return appearance.muted;
	}

	function tabHover(tab: Tab): { background?: string; color?: string } {
		if (activeTab === tab) {
			return {};
		}
		return { background: colors.hover, color: colors.text };
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<View style={{ flex: 1, background: colors.canvas, color: colors.text }}>
	<ArtBackdrop />

	<View style={{ flex: 1, minHeight: 0, gap: 24, paddingX: 40, paddingY: 24 }}>
		<View style={{ direction: 'row', align: 'center', shrink: 0 }}>
			<IconAction
				icon={caretDownBold}
				iconSize={24}
				boxSize={44}
				hoverBackground={colors.hover}
				onpress={() => panels.closeFullPlayer()}
			/>
		</View>

		{#if player.track}
			<View style={{ flex: 1, minHeight: 0, direction: 'row', gap: 56 }}>
				<View
					style={{ flex: 5, minWidth: 0, align: 'center', justify: 'center', gap: 24 }}
					onresize={(layout) => (leftColumnWidth = layout.width)}
				>
					<Button
						style={{
							size: coverSide,
							radius: radii.extraLarge,
							overflow: 'hidden',
							shadow: { y: 24, blur: 48, color: 'rgba(0, 0, 0, 0.5)' }
						}}
						onpress={() => panels.closeFullPlayer()}
					>
						<Cover
							thumbnails={player.track.thumbnails}
							size={512}
							style={{ size: coverSide, radius: radii.extraLarge }}
						/>
					</Button>
					<View style={{ width: coverSide, minWidth: 0 }}>
						<Label size="title" weight="bold" tone="default" truncate>{player.track.title}</Label>
						<View style={{ color: appearance.muted }}>
							<ArtistLinks artists={player.track.artists} size="large" />
						</View>
					</View>
					<View style={{ width: coverSide }}>
						<FullPlayerControls />
					</View>
				</View>

				<View style={{ flex: 6, minWidth: 0, minHeight: 0, gap: 16 }}>
					<View style={{ direction: 'row', gap: 8, shrink: 0 }}>
						{#each tabs as tab (tab.id)}
							<Button
								style={{
									paddingX: 16,
									paddingY: 6,
									radius: radii.full,
									background: tabBackground(tab.id),
									color: tabColor(tab.id),
									transition: hoverTransition
								}}
								hover={tabHover(tab.id)}
								onpress={() => (activeTab = tab.id)}
							>
								<Label size="small" weight="semibold">{tab.label}</Label>
							</Button>
						{/each}
					</View>

					{#if activeTab === 'lyrics'}
						<FullPlayerLyrics />
					{:else}
						<Scroll style={{ flex: 1, minHeight: 0 }}>
							<View style={{ gap: 16, paddingRight: 16 }}>
								<UpcomingQueue size="comfortable" />
							</View>
						</Scroll>
					{/if}
				</View>
			</View>
		{:else}
			<View style={{ flex: 1, align: 'center', justify: 'center' }}>
				<Label size="xlarge" tone="dim">Nothing is playing.</Label>
			</View>
		{/if}
	</View>
</View>
