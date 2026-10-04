<script lang="ts">
	import { LoadingSpinner } from '@neoworks-dev/ui';
	import type { ResourceStatus } from '#lib/pages/remoteResource.svelte.ts';

	let { status, errorMessage }: { status: ResourceStatus; errorMessage: string } = $props();
</script>

{#if status === 'loading'}
	<div class="flex justify-center py-24"><LoadingSpinner /></div>
{:else if status === 'failed'}
	<p class="py-24 text-center text-sm text-red">{errorMessage}</p>
{/if}
