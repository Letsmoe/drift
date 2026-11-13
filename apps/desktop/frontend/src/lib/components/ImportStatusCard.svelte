<script lang="ts">
	import { Icon, View } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Label from '#lib/components/Label.svelte';
	import { checkCircleFill, downloadSimple } from '#lib/icons.ts';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	let { flow }: { flow: MusicImportFlow } = $props();

	const finished = $derived(flow.step === 'done');
</script>

<View
	style={{
		gap: 8,
		padding: 12,
		radius: radii.large,
		background: colors.raised,
		cursor: 'pointer',
		transition: hoverTransition
	}}
	hover={{ background: colors.hover }}
	onpress={() => flow.openDialog()}
>
	<View style={{ direction: 'row', align: 'center', gap: 8 }}>
		{#if finished}
			<Icon svg={checkCircleFill} size={16} color={appearance.accent} style={{ shrink: 0 }} />
			<Label size="small" weight="semibold" truncate style={{ flex: 1, minWidth: 0 }}>
				Import finished
			</Label>
			<Label size="small" tone="muted">See summary</Label>
		{:else}
			<Icon svg={downloadSimple} size={16} color={appearance.dim} style={{ shrink: 0 }} />
			<Label size="small" weight="semibold" truncate style={{ flex: 1, minWidth: 0 }}>
				{flow.phaseLabel}
			</Label>
			{#if flow.progress}
				<Label size="small" tone="muted">{flow.progress.done} / {flow.progress.total}</Label>
			{/if}
		{/if}
	</View>
	{#if !finished}
		<View style={{ height: 4, radius: radii.full, background: colors.hover, overflow: 'hidden' }}>
			<View
				style={{
					width: `${Math.round(flow.progressFraction * 100)}%`,
					height: 4,
					radius: radii.full,
					background: appearance.accent
				}}
			/>
		</View>
	{/if}
</View>
