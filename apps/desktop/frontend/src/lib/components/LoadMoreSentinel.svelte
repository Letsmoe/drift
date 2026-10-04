<script lang="ts">
	import { LoadingSpinner } from '@neoworks-dev/ui';
	import { whenVisible } from '#lib/whenVisible.ts';

	let { itemCount, onvisible }: { itemCount: number; onvisible: () => void } = $props();
</script>

<!-- Re-created per page so it fires again when it is still in view after loading. -->
{#key itemCount}
	<div class="flex justify-center py-8" {@attach whenVisible(onvisible)}>
		<LoadingSpinner />
	</div>
{/key}
