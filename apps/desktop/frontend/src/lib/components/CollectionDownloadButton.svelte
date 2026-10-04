<script lang="ts">
	import ArrowCircleDownIcon from 'phosphor-svelte/lib/ArrowCircleDownIcon';
	import type { CollectionKind } from '#lib/api/downloads.ts';
	import ConfirmDialog from '#lib/components/ConfirmDialog.svelte';
	import ProgressRing from '#lib/components/ProgressRing.svelte';
	import { downloads } from '#lib/downloads/downloads.svelte.ts';

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
	const label = $derived(describeAction());
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

	function describeAction(): string {
		if (!collection) {
			return `Keep ${title} downloaded`;
		}
		return `Remove downloads of ${title}`;
	}

	function toggle() {
		if (!collection) {
			downloads.keepDownloaded(kind, id);
			return;
		}
		confirmingRemoval = true;
	}
</script>

<button
	type="button"
	aria-label={label}
	aria-pressed={collection !== undefined}
	title={label}
	onclick={toggle}
	class={[
		'duration-fast flex size-8 items-center justify-center transition-transform hover:scale-105',
		complete && 'text-green',
		!complete && 'text-dim hover:text-default'
	]}
>
	{#if complete}
		<ArrowCircleDownIcon size={32} weight="fill" />
	{:else if collection}
		<ProgressRing {fraction} size={28} strokeWidth={3} />
	{:else}
		<ArrowCircleDownIcon size={32} />
	{/if}
</button>

{#if confirmingRemoval}
	<ConfirmDialog
		title="Remove downloads"
		message="Delete the downloaded songs of &quot;{title}&quot; and stop keeping it downloaded? Songs that another downloaded collection or a single download still needs stay."
		confirmLabel="Remove"
		onconfirm={() => downloads.removeCollection(kind, id)}
		onclose={() => (confirmingRemoval = false)}
	/>
{/if}
