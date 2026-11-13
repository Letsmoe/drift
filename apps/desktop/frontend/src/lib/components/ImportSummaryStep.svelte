<script lang="ts">
	import { Scroll, View } from '@neoworks/svelte-gpui/ui';
	import { Button, Disclosure } from '@neoworks/svelte-gpui-ui';
	import type { ImportResult } from '#lib/api/musicImport.ts';
	import Label from '#lib/components/Label.svelte';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';
	import { goto } from '#lib/router.svelte.ts';

	let {
		flow,
		result,
		onclose
	}: { flow: MusicImportFlow; result: ImportResult; onclose: () => void } = $props();

	const categoryRows = $derived([
		{ label: 'Liked Songs', count: result.liked },
		{ label: 'Albums', count: result.albums },
		{ label: 'Artists', count: result.artists },
		{ label: 'Podcasts', count: result.shows }
	]);
	const visibleRows = $derived(
		categoryRows.filter((row) => row.count.matched + row.count.unmatched > 0)
	);
	const introduction = $derived(introductionFor(flow.isAccountSource));
	const foundLabel = $derived(wordFor(flow.isAccountSource, 'imported', 'matched'));
	const missingLabel = $derived(wordFor(flow.isAccountSource, 'skipped', 'not found'));
	const unmatchedTitle = $derived(unmatchedTitleFor(flow.isAccountSource));

	function introductionFor(isAccountSource: boolean): string {
		if (isAccountSource) {
			return 'Your YouTube Music library is synced. Run it again any time to pick up changes.';
		}
		return 'Import finished. Items without a confident match were skipped.';
	}

	function wordFor(isAccountSource: boolean, accountWord: string, matchedWord: string): string {
		if (isAccountSource) {
			return accountWord;
		}
		return matchedWord;
	}

	function unmatchedTitleFor(isAccountSource: boolean): string {
		if (isAccountSource) {
			return `${result.unmatched.length} could not be imported`;
		}
		return `${result.unmatched.length} not found on YouTube Music`;
	}

	function countText(matched: number, unmatched: number): string {
		if (unmatched > 0) {
			return `${matched} ${foundLabel}, ${unmatched} ${missingLabel}`;
		}
		return `${matched} ${foundLabel}`;
	}

	async function openPlaylist(playlistId: string): Promise<void> {
		onclose();
		goto(`/library/playlist/${encodeURIComponent(playlistId)}`);
	}

	function describeUnmatched(artists: string[], playlist: string): string {
		const parts = [...artists];
		if (playlist !== '') {
			parts.push(`in ${playlist}`);
		}
		return parts.join(' · ');
	}
</script>

<View style={{ gap: 16 }}>
	<Label size="small" tone="muted">{introduction}</Label>

	<View>
		{#each visibleRows as row (row.label)}
			<View
				style={{ direction: 'row', align: 'center', justify: 'between', paddingX: 8, paddingY: 6 }}
			>
				<Label size="small">{row.label}</Label>
				<Label size="small" tone="muted">{countText(row.count.matched, row.count.unmatched)}</Label>
			</View>
		{/each}
		{#each result.playlists as playlist (playlist.key)}
			<View style={{ direction: 'row', align: 'center', gap: 12, paddingX: 8, paddingY: 6 }}>
				<Label size="small" truncate style={{ flex: 1, minWidth: 0 }}>{playlist.name}</Label>
				<Label size="small" tone="muted">{countText(playlist.matched, playlist.unmatched)}</Label>
				<Button size="sm" variant="ghost" onpress={() => openPlaylist(playlist.playlistId)}>
					Open
				</Button>
			</View>
		{/each}
	</View>

	{#if result.unmatched.length > 0}
		<Disclosure title={unmatchedTitle}>
			<Scroll style={{ maxHeight: 224 }}>
				<View style={{ gap: 4 }}>
					{#each result.unmatched as item, index (index)}
						<Label size="small">
							{item.title}
							<Label size="small" tone="dim">{describeUnmatched(item.artists, item.playlist)}</Label
							>
						</Label>
					{/each}
				</View>
			</Scroll>
		</Disclosure>
	{/if}

	<View style={{ direction: 'row', justify: 'end', gap: 8 }}>
		<Button variant="ghost" onpress={() => flow.restart()}>Import more</Button>
		<Button variant="primary" onpress={onclose}>Done</Button>
	</View>
</View>
