<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { Button, ProgressIndicator } from '@neoworks/svelte-gpui-ui';
	import Label from '#lib/components/Label.svelte';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';

	let { flow }: { flow: MusicImportFlow } = $props();

	const currentItem = $derived(currentItemOf());

	function currentItemOf(): string {
		if (flow.progress) {
			return flow.progress.current;
		}
		return '';
	}
</script>

<View style={{ gap: 16 }}>
	<View style={{ direction: 'row', align: 'center', justify: 'between' }}>
		<Label size="small" weight="semibold">{flow.phaseLabel}</Label>
		{#if flow.progress}
			<Label size="small" tone="muted">{flow.progress.done} / {flow.progress.total}</Label>
		{/if}
	</View>
	<ProgressIndicator variant="linear" value={flow.progressFraction * 100} label="Import progress" />
	<Label size="small" tone="muted" truncate style={{ minHeight: 20 }}>{currentItem}</Label>
	<View style={{ direction: 'row', justify: 'end', gap: 8 }}>
		<Button variant="ghost" onpress={() => flow.cancel()}>Cancel</Button>
		<Button onpress={() => flow.closeDialog()}>Continue in background</Button>
	</View>
</View>
