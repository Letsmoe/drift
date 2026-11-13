<!--
  The last block of a feed: it asks for the next page as soon as it is built, which a `PageList` does only
  when the block is near the window. Give it a key that changes with the item count, so it is built again,
  and asks again, when the page it asked for did not fill the window.
-->
<script lang="ts">
	import { untrack } from 'svelte';
	import { View } from '@neoworks/svelte-gpui/ui';
	import { LoadingSpinner } from '@neoworks/svelte-gpui-ui';
	import { colors } from '#lib/theme.ts';

	let { onvisible }: { onvisible: () => void } = $props();

	$effect(() => {
		untrack(() => onvisible());
	});
</script>

<View
	style={{ minHeight: 96, align: 'center', justify: 'center', paddingY: 32, color: colors.text }}
>
	<LoadingSpinner />
</View>
