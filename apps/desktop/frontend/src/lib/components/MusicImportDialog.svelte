<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { LoadingSpinner } from '@neoworks/svelte-gpui-ui';
	import Dialog from '#lib/components/Dialog.svelte';
	import ImportPreviewStep from '#lib/components/ImportPreviewStep.svelte';
	import ImportProgressStep from '#lib/components/ImportProgressStep.svelte';
	import ImportSourceStep from '#lib/components/ImportSourceStep.svelte';
	import ImportSummaryStep from '#lib/components/ImportSummaryStep.svelte';
	import YouTubeSignInStep from '#lib/components/YouTubeSignInStep.svelte';
	import { musicImport as flow } from '#lib/import/musicImport.svelte.ts';

	const onclose = () => flow.closeDialog();
</script>

<Dialog title="Import music" wide {onclose}>
	{#if flow.step === 'source'}
		<ImportSourceStep {flow} />
	{:else if flow.step === 'signin'}
		<YouTubeSignInStep {flow} />
	{:else if flow.step === 'loading'}
		<View style={{ align: 'center', paddingY: 40 }}><LoadingSpinner /></View>
	{:else if flow.step === 'preview' && flow.summary}
		<ImportPreviewStep {flow} summary={flow.summary} />
	{:else if flow.step === 'running'}
		<ImportProgressStep {flow} />
	{:else if flow.step === 'done' && flow.result}
		<ImportSummaryStep {flow} result={flow.result} {onclose} />
	{/if}
</Dialog>
