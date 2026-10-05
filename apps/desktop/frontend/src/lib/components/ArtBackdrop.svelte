<script lang="ts">
	import type { ClassValue } from 'svelte/elements';

	// Reads --art-1..4 from the surrounding .art-shell; place it inside a positioned parent.
	let { class: className = '' }: { class?: ClassValue } = $props();
</script>

<div
	aria-hidden="true"
	class={['art-backdrop pointer-events-none absolute inset-0', className]}
></div>

<style>
	/* One static layer: it repaints only while the colors fade to a new song. A drifting
	   animation repainted four window-sized gradients every frame (~75% CPU in WebKitGTK). */
	.art-backdrop {
		background:
			linear-gradient(color-mix(in oklab, var(--bg) 40%, transparent)),
			radial-gradient(60% 70% at 10% 0%, var(--art-1), transparent),
			radial-gradient(55% 65% at 95% 5%, var(--art-2), transparent),
			radial-gradient(60% 70% at 30% 105%, var(--art-3), transparent),
			radial-gradient(50% 60% at 100% 100%, var(--art-4), transparent);
		opacity: 0.85;
	}
</style>
