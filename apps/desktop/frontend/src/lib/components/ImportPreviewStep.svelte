<script lang="ts">
	import { Scroll, View } from '@neoworks/svelte-gpui/ui';
	import { Button, Checkbox } from '@neoworks/svelte-gpui-ui';
	import type { ImportSummary } from '#lib/api/musicImport.ts';
	import Label from '#lib/components/Label.svelte';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';
	import { colors } from '#lib/theme.ts';

	let { flow, summary }: { flow: MusicImportFlow; summary: ImportSummary } = $props();

	const allPlaylistsSelected = $derived(
		summary.playlists.length > 0 && flow.selectedPlaylistKeys.size === summary.playlists.length
	);
	const introduction = $derived(introductionFor(flow.isAccountSource));

	function introductionFor(isAccountSource: boolean): string {
		if (isAccountSource) {
			return 'Choose what to bring over. Your YouTube Music library is imported as it is.';
		}
		return 'Choose what to bring over. Everything is matched to YouTube Music.';
	}
</script>

{#snippet countRow(label: string, count: number, checked: boolean, ontoggle: () => void)}
	<View style={{ direction: 'row', align: 'center', gap: 12, paddingX: 8, paddingY: 4 }}>
		<View style={{ flex: 1, minWidth: 0 }}>
			<Checkbox {label} {checked} disabled={count === 0} onchange={ontoggle} />
		</View>
		<Label size="small" tone="muted">{count}</Label>
	</View>
{/snippet}

<View style={{ gap: 12 }}>
	<Label size="small" tone="muted">{introduction}</Label>

	<View>
		{@render countRow('Liked Songs', summary.liked, flow.includeLiked, () => {
			flow.includeLiked = !flow.includeLiked;
		})}
		{@render countRow('Albums', summary.albums, flow.includeAlbums, () => {
			flow.includeAlbums = !flow.includeAlbums;
		})}
		{@render countRow('Artists', summary.artists, flow.includeArtists, () => {
			flow.includeArtists = !flow.includeArtists;
		})}
		{@render countRow('Podcasts', summary.shows, flow.includeShows, () => {
			flow.includeShows = !flow.includeShows;
		})}
	</View>

	{#if summary.playlists.length > 0}
		<View style={{ paddingTop: 8, borderTopWidth: 1, borderColor: colors.line }}>
			<View style={{ paddingX: 8, paddingY: 4 }}>
				<Checkbox
					label={`Playlists (${summary.playlists.length})`}
					checked={allPlaylistsSelected}
					onchange={() => flow.toggleAllPlaylists()}
				/>
			</View>
			<Scroll style={{ maxHeight: 224 }}>
				{#each summary.playlists as playlist (playlist.key)}
					<View
						style={{
							direction: 'row',
							align: 'center',
							gap: 12,
							paddingLeft: 32,
							paddingRight: 8,
							paddingY: 4
						}}
					>
						<View style={{ flex: 1, minWidth: 0 }}>
							<Checkbox
								label={playlist.name}
								checked={flow.selectedPlaylistKeys.has(playlist.key)}
								onchange={() => flow.togglePlaylist(playlist.key)}
							/>
						</View>
						{#if playlist.trackCount > 0}
							<Label size="small" tone="muted">{playlist.trackCount} tracks</Label>
						{/if}
					</View>
				{/each}
			</Scroll>
		</View>
	{/if}

	{#if summary.trackLimit > 0}
		<Label size="tiny" tone="dim">
			Public links show at most {summary.trackLimit} tracks per playlist; use the data export for the
			full playlist.
		</Label>
	{/if}
	{#if flow.cancelled}
		<Label size="small" tone="muted">
			Import stopped. Tracks matched so far are remembered, so starting again continues quickly.
		</Label>
	{/if}
	{#if flow.errorMessage !== ''}
		<Label size="small" tone="danger">{flow.errorMessage}</Label>
	{/if}

	<View style={{ direction: 'row', justify: 'end', gap: 8, paddingTop: 8 }}>
		<Button variant="ghost" onpress={() => flow.restart()}>Back</Button>
		<Button variant="primary" disabled={!flow.canStart} onpress={() => flow.start()}>Import</Button>
	</View>
</View>
