<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { Attachment } from 'svelte/attachments';

	let { title, onclose, children }: { title: string; onclose: () => void; children: Snippet } =
		$props();

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
	class="m-auto w-full max-w-md rounded-xl bg-elevated p-6 text-default shadow-lg backdrop:bg-black/60"
>
	<h2 class="pb-4 text-2xl font-bold">{title}</h2>
	{@render children()}
</dialog>
