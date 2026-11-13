<script lang="ts">
	import { Button } from '@neoworks/svelte-gpui/ui';
	import type { Item, ItemKind } from '#lib/api/music.ts';
	import Label from '#lib/components/Label.svelte';
	import { library } from '#lib/library/library.svelte.ts';
	import { colors, hoverTransition, pressedSolidLook, radii } from '#lib/theme.ts';

	let { item }: { item: Item } = $props();

	const following = $derived(library.isSaved(item.kind as ItemKind, item.id));
	const label = $derived(labelFor(following));

	function labelFor(isFollowing: boolean): string {
		if (isFollowing) {
			return 'Following';
		}
		return 'Follow';
	}
</script>

<Button
	style={{
		height: 32,
		paddingX: 16,
		radius: radii.full,
		borderWidth: 1,
		borderColor: colors.lineStrong,
		align: 'center',
		justify: 'center',
		transition: hoverTransition
	}}
	hover={{ borderColor: colors.text }}
	active={pressedSolidLook}
	onpress={() => library.toggleSaved(item)}
>
	<Label size="small" weight="semibold" tone="default">{label}</Label>
</Button>
