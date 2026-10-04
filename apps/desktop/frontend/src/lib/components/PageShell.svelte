<script lang="ts">
	import type { Snippet } from 'svelte';
	import { gradientColorFor } from '#lib/pageColor.ts';

	let { seed, children }: { seed: string; children: Snippet } = $props();

	const gradient = $derived(
		`linear-gradient(to bottom, ${gradientColorFor(seed)} 0, transparent 26rem)`
	);
</script>

<!-- Cancels the layout padding so the header gradient reaches the pane edges. -->
<div class="-mx-3 -my-4 min-h-full px-6 py-6" style:background-image={gradient}>
	{@render children()}
</div>
