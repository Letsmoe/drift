<script lang="ts">
	let {
		fraction,
		size = 18,
		strokeWidth = 2.5
	}: {
		/** Completed share between 0 and 1; anything else shows an indeterminate ring. */
		fraction: number;
		size?: number;
		strokeWidth?: number;
	} = $props();

	const radius = $derived((size - strokeWidth) / 2);
	const circumference = $derived(2 * Math.PI * radius);
	const known = $derived(fraction > 0 && fraction <= 1);
	const visibleFraction = $derived(visibleShare(known, fraction));

	function visibleShare(isKnown: boolean, share: number): number {
		if (isKnown) {
			return share;
		}
		return 0.25;
	}
</script>

<svg
	width={size}
	height={size}
	viewBox="0 0 {size} {size}"
	aria-hidden="true"
	class={['-rotate-90', !known && 'animate-spin']}
>
	<circle
		cx={size / 2}
		cy={size / 2}
		r={radius}
		fill="none"
		stroke="currentColor"
		stroke-opacity="0.25"
		stroke-width={strokeWidth}
	/>
	<circle
		cx={size / 2}
		cy={size / 2}
		r={radius}
		fill="none"
		stroke="currentColor"
		stroke-width={strokeWidth}
		stroke-linecap="round"
		stroke-dasharray={circumference}
		stroke-dashoffset={circumference * (1 - visibleFraction)}
	/>
</svg>
