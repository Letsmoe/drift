<script lang="ts">
	import { Button } from '@neoworks-dev/ui';
	import Dialog from '#lib/components/Dialog.svelte';

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

	function confirm() {
		onclose();
		onconfirm();
	}
</script>

<Dialog {title} {onclose}>
	<p class="pb-6 text-sm text-muted">{message}</p>
	<div class="flex justify-end gap-2">
		<Button variant="ghost" onclick={onclose}>Cancel</Button>
		<Button variant="danger" onclick={confirm}>{confirmLabel}</Button>
	</div>
</Dialog>
