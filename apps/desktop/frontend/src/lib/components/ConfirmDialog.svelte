<script lang="ts">
	import { Text } from '@neoworks/svelte-gpui/ui';
	import { Button, Dialog } from '@neoworks/svelte-gpui-ui';

	let {
		title,
		message,
		confirmLabel,
		onconfirm,
		onclose
	}: {
		title: string;
		message: string;
		confirmLabel: string;
		onconfirm: () => void;
		onclose: () => void;
	} = $props();

	function confirm(): void {
		onclose();
		onconfirm();
	}
</script>

<Dialog open {title} {onclose}>
	<Text>{message}</Text>
	{#snippet actions()}
		<Button variant="ghost" onpress={onclose}>Cancel</Button>
		<Button variant="danger" onpress={confirm}>{confirmLabel}</Button>
	{/snippet}
</Dialog>
