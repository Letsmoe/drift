<script lang="ts">
	import CheckCircleIcon from 'phosphor-svelte/lib/CheckCircleIcon';
	import DownloadSimpleIcon from 'phosphor-svelte/lib/DownloadSimpleIcon';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';

	let { flow }: { flow: MusicImportFlow } = $props();

	const finished = $derived(flow.step === 'done');
</script>

<button
	type="button"
	onclick={() => flow.openDialog()}
	title="Open the import"
	class="flex flex-col gap-2 rounded-lg bg-raised p-3 text-left hover:bg-hover"
>
	<div class="flex items-center gap-2 text-sm">
		{#if finished}
			<CheckCircleIcon size={16} weight="fill" class="shrink-0 text-action" />
			<span class="flex-1 truncate font-semibold">Import finished</span>
			<span class="text-muted">See summary</span>
		{:else}
			<DownloadSimpleIcon size={16} class="shrink-0 text-dim" />
			<span class="flex-1 truncate font-semibold">{flow.phaseLabel}</span>
			{#if flow.progress}
				<span class="text-muted tabular-nums">{flow.progress.done} / {flow.progress.total}</span>
			{/if}
		{/if}
	</div>
	{#if !finished}
		<div
			class="h-1 overflow-hidden rounded-full bg-hover"
			role="progressbar"
			aria-label="Import progress"
			aria-valuemin="0"
			aria-valuemax="100"
			aria-valuenow={Math.round(flow.progressFraction * 100)}
		>
			<div
				class="h-full rounded-full bg-action transition-all"
				style:width="{flow.progressFraction * 100}%"
			></div>
		</div>
	{/if}
</button>
