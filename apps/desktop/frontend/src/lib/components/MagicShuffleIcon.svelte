<script lang="ts">
	import ShuffleIcon from 'phosphor-svelte/lib/ShuffleIcon';
	import SparkleIcon from 'phosphor-svelte/lib/SparkleIcon';

	type IconWeight = 'bold' | 'duotone' | 'fill' | 'light' | 'thin' | 'regular';

	let { size = 18, weight = 'regular' }: { size?: number | string; weight?: IconWeight } = $props();

	const sparkleSize = $derived(sparkleSizeFor(size));

	// Phosphor sizes may be CSS strings ("1em"); those get a fixed sparkle.
	function sparkleSizeFor(iconSize: number | string): number {
		if (typeof iconSize === 'number') {
			return Math.round(iconSize * 0.55);
		}
		return 10;
	}
</script>

<span class="relative inline-flex">
	<ShuffleIcon {size} {weight} />
	<SparkleIcon size={sparkleSize} weight="fill" class="absolute -top-1 -right-1.5" />
</span>
