<script lang="ts">
	import { Scroll, Text, View } from '@neoworks/svelte-gpui/ui';
	import { pages } from 'virtual:pages';
	import { router } from '#lib/router.svelte.ts';

	const Page = $derived(pages[router.page]);
</script>

<View style={{ flex: 1, background: '#141416', color: '#fafafa', padding: 8, gap: 8 }}>
	<Scroll style={{ flex: 1, padding: 16 }}>
		{#key router.href}
			{#if Page}
				<Page params={router.params} query={router.query} />
			{:else}
				<View style={{ gap: 4 }}>
					<Text style={{ fontSize: 18, fontWeight: 'semibold' }}>{router.page}</Text>
					<Text style={{ color: '#a1a1aa' }}>This page is not written yet ({router.href}).</Text>
				</View>
			{/if}
		{/key}
	</Scroll>
</View>
