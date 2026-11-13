<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { LoadingSpinner } from '@neoworks/svelte-gpui-ui';
	import Label from '#lib/components/Label.svelte';
	import type { ResourceStatus } from '#lib/pages/remoteResource.svelte.ts';
	import { colors } from '#lib/theme.ts';

	let { status, errorMessage }: { status: ResourceStatus; errorMessage: string } = $props();
</script>

{#if status === 'loading'}
	<View style={{ align: 'center', paddingY: 96, color: colors.text }}>
		<LoadingSpinner />
	</View>
{:else if status === 'failed'}
	<Label size="small" tone="danger" style={{ textAlign: 'center', paddingY: 96 }}>
		{errorMessage}
	</Label>
{/if}
