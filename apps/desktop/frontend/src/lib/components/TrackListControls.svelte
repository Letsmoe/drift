<script lang="ts">
	import { Icon, View } from '@neoworks/svelte-gpui/ui';
	import { Select } from '@neoworks/svelte-gpui-ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Label from '#lib/components/Label.svelte';
	import { magnifyingGlass } from '#lib/icons.ts';
	import TextInput from '#lib/layout/TextInput.svelte';
	import {
		trackSorts,
		type TrackListFilter,
		type TrackSort
	} from '#lib/pages/trackListFilter.svelte.ts';
	import { radii } from '#lib/theme.ts';

	let { filter }: { filter: TrackListFilter } = $props();

	const sortOptions = trackSorts.map((option) => ({ value: option.value, label: option.label }));

	function changeSort(value: string | string[]): void {
		if (typeof value === 'string') {
			filter.sort = value as TrackSort;
		}
	}
</script>

<View style={{ direction: 'row', align: 'center', justify: 'between', gap: 12, paddingBottom: 8 }}>
	<View style={{ width: 224, justify: 'center' }}>
		<TextInput
			bind:value={filter.query}
			placeholder="Search in playlist"
			height={32}
			style={{ paddingLeft: 36, radius: radii.full }}
		/>
		<View
			style={{
				position: 'absolute',
				left: 12,
				top: 0,
				bottom: 0,
				justify: 'center',
				color: appearance.dim
			}}
		>
			<Icon svg={magnifyingGlass} size={16} />
		</View>
	</View>
	<View style={{ direction: 'row', align: 'center', gap: 8 }}>
		<Label size="small" tone="dim">Sort by</Label>
		<Select
			options={sortOptions}
			value={filter.sort}
			onChange={changeSort}
			size="sm"
			style={{ width: 150 }}
		/>
	</View>
</View>
