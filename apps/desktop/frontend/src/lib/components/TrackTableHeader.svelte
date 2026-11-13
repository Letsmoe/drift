<script lang="ts">
	import { Icon, View } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Label from '#lib/components/Label.svelte';
	import {
		actionColumnWidth,
		indexColumnWidth,
		rowGap,
		rowPaddingX
	} from '#lib/components/trackColumns.ts';
	import { clock } from '#lib/icons.ts';
	import { colors } from '#lib/theme.ts';

	let { showAlbum, showDateAdded = false }: { showAlbum: boolean; showDateAdded?: boolean } =
		$props();
</script>

<View
	style={{
		height: 36,
		direction: 'row',
		align: 'center',
		gap: rowGap,
		paddingX: rowPaddingX,
		marginBottom: 8,
		borderBottomWidth: 1,
		borderColor: colors.line,
		color: appearance.dim
	}}
>
	<Label size="small" style={{ width: indexColumnWidth, textAlign: 'right', shrink: 0 }}>#</Label>
	<Label size="small" style={{ flex: 6, minWidth: 0 }}>Title</Label>
	{#if showAlbum}
		<Label size="small" style={{ flex: 4, minWidth: 0 }}>Album</Label>
	{/if}
	{#if showDateAdded}
		<Label size="small" style={{ flex: 3, minWidth: 0 }}>Date added</Label>
	{/if}
	<View
		style={{
			width: actionColumnWidth,
			shrink: 0,
			direction: 'row',
			justify: 'end',
			paddingRight: 36
		}}
	>
		<Icon svg={clock} size={16} />
	</View>
</View>
