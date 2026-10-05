<script lang="ts">
	// Considered from @neoworks-dev/ui: none; it has no slider. Music-specific for
	// now (seek, volume); move it into the ui package if another app needs one.
	let {
		value,
		max,
		label,
		step = 1,
		disabled = false,
		oninput
	}: {
		value: number;
		max: number;
		label: string;
		step?: number;
		disabled?: boolean;
		oninput: (value: number) => void;
	} = $props();

	const fillPercent = $derived(computeFillPercent(value, max));

	function computeFillPercent(value: number, max: number): number {
		if (max <= 0) {
			return 0;
		}
		return Math.min(100, (value / max) * 100);
	}

	function handleInput(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		oninput(Number(input.value));
	}
</script>

<input
	type="range"
	aria-label={label}
	min="0"
	{max}
	{step}
	{value}
	{disabled}
	oninput={handleInput}
	style="--fill: {fillPercent}%"
	class="range-slider h-3 w-full cursor-pointer appearance-none bg-transparent disabled:cursor-default"
/>

<style>
	.range-slider::-webkit-slider-runnable-track {
		height: 4px;
		border-radius: 9999px;
		background: linear-gradient(
			to right,
			var(--primary) var(--fill),
			var(--surface-hover) var(--fill)
		);
	}

	.range-slider::-webkit-slider-thumb {
		appearance: none;
		width: 12px;
		height: 12px;
		margin-top: -4px;
		border-radius: 9999px;
		background: var(--text);
		opacity: 0;
		transition: opacity var(--dur-fast);
	}

	.range-slider:hover:not(:disabled)::-webkit-slider-thumb,
	.range-slider:focus-visible::-webkit-slider-thumb {
		opacity: 1;
	}

	.range-slider:disabled::-webkit-slider-runnable-track {
		background: var(--surface-hover);
	}
</style>
