<script lang="ts">
	import { Button } from '@neoworks-dev/ui';
	import Dialog from '#lib/components/Dialog.svelte';

	let {
		name,
		description,
		onsave,
		onclose
	}: {
		name: string;
		description: string;
		onsave: (name: string, description: string) => void;
		onclose: () => void;
	} = $props();

	// The drafts start from the current values and are edited independently.
	// svelte-ignore state_referenced_locally
	let draftName = $state(name);
	// svelte-ignore state_referenced_locally
	let draftDescription = $state(description);
	const canSave = $derived(draftName.trim() !== '');

	function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!canSave) {
			return;
		}
		onsave(draftName.trim(), draftDescription.trim());
		onclose();
	}
</script>

<Dialog title="Edit details" {onclose}>
	<form onsubmit={submit} class="flex flex-col gap-3">
		<input
			bind:value={draftName}
			aria-label="Name"
			placeholder="Playlist name"
			class="h-11 rounded-md bg-raised px-3 text-base outline-none placeholder:text-dim focus:border-line-strong"
		/>
		<textarea
			bind:value={draftDescription}
			aria-label="Description"
			placeholder="Add an optional description"
			rows="3"
			class="resize-none rounded-md bg-raised p-3 text-base outline-none placeholder:text-dim"
		></textarea>
		<div class="flex justify-end gap-2 pt-2">
			<Button variant="ghost" onclick={onclose}>Cancel</Button>
			<Button variant="primary" type="submit" disabled={!canSave}>Save</Button>
		</div>
	</form>
</Dialog>
