<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { Attachment } from 'svelte/attachments';

	let {
		title,
		onclose,
		wide = false,
		children
	}: { title: string; onclose: () => void; wide?: boolean; children: Snippet } = $props();

	const showModal: Attachment<HTMLDialogElement> = (element) => {
		element.showModal();
	};

	function closeOnBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			onclose();
		}
	}
</script>

<dialog
	{@attach showModal}
	{onclose}
	aria-label={title}
	onclick={closeOnBackdropClick}
	class:max-w-md={!wide}
	class:max-w-2xl={wide}
	class="m-auto w-full rounded-xl bg-elevated p-6 text-default shadow-lg backdrop:bg-black/60"
>
	<h2 class="pb-4 text-2xl font-bold">{title}</h2>
	{@render children()}
</dialog>
