<script lang="ts">
	import { LoadingSpinner } from '@neoworks-dev/ui';
	import { whenVisible } from '#lib/whenVisible.ts';

	let { itemCount, onvisible }: { itemCount: number; onvisible: () => void } = $props();

	// The spinner only appears once the sentinel is in view: an off-screen spinner keeps the
	// webview repainting the translucent panels above the backdrop.
	let reachedAtCount = $state(-1);

	function reach(): void {
		reachedAtCount = itemCount;
		onvisible();
	}
</script>

<!-- Re-created per page so it fires again when it is still in view after loading. -->
{#key itemCount}
	<div class="flex min-h-24 justify-center py-8" {@attach whenVisible(reach)}>
		{#if reachedAtCount === itemCount}
			<LoadingSpinner />
		{/if}
	</div>
{/key}
