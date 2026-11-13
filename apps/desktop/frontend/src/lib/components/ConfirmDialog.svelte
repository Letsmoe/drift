<script lang="ts">
	import { Button } from '@neoworks/svelte-gpui-ui';
	import { View } from '@neoworks/svelte-gpui/ui';
	import Dialog from '#lib/components/Dialog.svelte';
	import Label from '#lib/components/Label.svelte';

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

<Dialog {title} {onclose}>
	<Label size="small" tone="muted" style={{ paddingBottom: 24 }}>{message}</Label>
	<View style={{ direction: 'row', justify: 'end', gap: 8 }}>
		<Button variant="ghost" onpress={onclose}>Cancel</Button>
		<Button variant="danger" onpress={confirm}>{confirmLabel}</Button>
	</View>
</Dialog>
