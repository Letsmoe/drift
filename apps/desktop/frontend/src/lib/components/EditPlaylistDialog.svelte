<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { Button, Dialog } from '@neoworks/svelte-gpui-ui';
	import TextArea from '#lib/components/TextArea.svelte';
	import TextInput from '#lib/layout/TextInput.svelte';
	import { colors, fontSizes } from '#lib/theme.ts';

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

	function submit(): void {
		if (!canSave) {
			return;
		}
		onsave(draftName.trim(), draftDescription.trim());
		onclose();
	}
</script>

<Dialog open title="Edit details" {onclose}>
	<View style={{ gap: 12 }}>
		<TextInput
			bind:value={draftName}
			placeholder="Playlist name"
			height={44}
			autofocus
			onsubmit={submit}
		/>
		<TextArea
			bind:value={draftDescription}
			rows={3}
			placeholder="Add an optional description"
			style={{ background: colors.raised, fontSize: fontSizes.body, lineHeight: 21, padding: 12 }}
		/>
	</View>
	{#snippet actions()}
		<Button variant="ghost" onpress={onclose}>Cancel</Button>
		<Button variant="primary" disabled={!canSave} onpress={submit}>Save</Button>
	{/snippet}
</Dialog>
