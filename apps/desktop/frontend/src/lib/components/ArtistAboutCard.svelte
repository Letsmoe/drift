<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import type { Artist } from '#lib/api/music.ts';
	import { createItem } from '#lib/api/music.ts';
	import { artistInfo } from '#lib/artist/artistInfo.svelte.ts';
	import Cover from '#lib/components/Cover.svelte';
	import FollowButton from '#lib/components/FollowButton.svelte';
	import Label from '#lib/components/Label.svelte';
	import { itemHref } from '#lib/format.ts';
	import { goto } from '#lib/router.svelte.ts';

	const coverSide = 56;

	let { artist }: { artist: Artist } = $props();

	const artistPage = $derived(artistInfo.pageOf(artist.id));

	$effect(() => {
		artistInfo.load(artist.id);
	});
</script>

{#if artistPage}
	<View style={{ gap: 12 }}>
		<Label size="small" weight="semibold" tone="muted">About the artist</Label>
		<View style={{ direction: 'row', align: 'center', gap: 12 }}>
			<Cover
				thumbnails={artistPage.thumbnails}
				size={coverSide}
				style={{ size: coverSide, shrink: 0, radius: coverSide / 2 }}
			/>
			<View style={{ flex: 1, minWidth: 0, align: 'start', gap: 6 }}>
				<View
					style={{ maxWidth: '100%', cursor: 'pointer' }}
					hover={{ underline: true }}
					onpress={() => goto(itemHref('artist', artist.id))}
				>
					<Label size="body" weight="semibold" truncate>{artistPage.name}</Label>
				</View>
				<FollowButton
					item={createItem('artist', artist.id, artistPage.name, '', artistPage.thumbnails)}
				/>
			</View>
		</View>
		{#if artistPage.description}
			<Label size="body" tone="muted" lines={4}>{artistPage.description}</Label>
		{/if}
	</View>
{/if}
