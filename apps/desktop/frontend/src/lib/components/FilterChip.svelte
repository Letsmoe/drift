<!--
  A chip that narrows a list. Whether it is selected is the page's to say (a radio-like filter, or one that a
  second press clears), so the kit chip's own toggling is not allowed to differ from it.
-->
<script lang="ts">
	import { Chip } from '@neoworks/svelte-gpui-ui';

	let { label, selected, onpress }: { label: string; selected: boolean; onpress: () => void } =
		$props();
</script>

<Chip variant="filter" {label} bind:selected={() => selected, () => {}} {onpress} />
