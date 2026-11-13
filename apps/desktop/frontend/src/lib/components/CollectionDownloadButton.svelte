<script lang="ts">
	import { Button } from '@neoworks/svelte-gpui/ui';
	import { ProgressIndicator } from '@neoworks/svelte-gpui-ui';
	import type { CollectionKind } from '#lib/api/downloads.ts';
	import { appearance } from '#lib/appearance.svelte.ts';
	import ConfirmDialog from '#lib/components/ConfirmDialog.svelte';
	import IconAction from '#lib/components/IconAction.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';
	import { progressPercent } from '#lib/downloads/downloadProgress.ts';
	import { arrowCircleDown, arrowCircleDownFill } from '#lib/icons.ts';
	import { colors } from '#lib/theme.ts';

	let {
		kind,
		id,
		title
	}: {
		kind: CollectionKind;
		id: string;
		title: string;
	} = $props();

	const collection = $derived(downloads.collection(kind, id));
	const complete = $derived(
		collection !== undefined &&
			collection.trackCount > 0 &&
			collection.doneCount >= collection.trackCount
	);
	const fraction = $derived(progressFraction());
	let confirmingRemoval = $state(false);

	// A kept YouTube collection looks for new tracks whenever its page opens.
	$effect(() => {
		downloads.refreshOnOpen(kind, id);
	});

	function progressFraction(): number {
		if (!collection || collection.trackCount === 0) {
			return 0;
		}
		return collection.doneCount / collection.trackCount;
	}

	function colorFor(isComplete: boolean): string {
		if (isComplete) {
			return appearance.accent;
		}
		return appearance.dim;
	}

	function idleIconFor(isComplete: boolean): string {
		if (isComplete) {
			return arrowCircleDownFill;
		}
		return arrowCircleDown;
	}

	function toggle(): void {
		if (!collection) {
			downloads.keepDownloaded(kind, id);
			return;
		}
		confirmingRemoval = true;
	}
</script>

{#if collection && !complete}
	<Button
		style={{ size: 32, align: 'center', justify: 'center', color: appearance.dim }}
		hover={{ color: colors.text }}
		onpress={toggle}
	>
		<ProgressIndicator
			variant="circular"
			label="Downloading"
			value={progressPercent(fraction)}
			size={28}
			thickness={5}
			tone="inherit"
			color={appearance.muted}
		/>
	</Button>
{:else}
	<IconAction
		icon={idleIconFor(complete)}
		iconSize={32}
		boxSize={32}
		color={colorFor(complete)}
		hoverColor={colors.text}
		onpress={toggle}
	/>
{/if}

{#if confirmingRemoval}
	<ConfirmDialog
		title="Remove downloads"
		message={`Delete the downloaded songs of "${title}" and stop keeping it downloaded? Songs that another downloaded collection or a single download still needs stay.`}
		confirmLabel="Remove"
		onconfirm={() => downloads.removeCollection(kind, id)}
		onclose={() => (confirmingRemoval = false)}
	/>
{/if}
