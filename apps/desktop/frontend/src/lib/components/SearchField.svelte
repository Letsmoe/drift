<script lang="ts">
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';

	let {
		value = $bindable(''),
		placeholder = 'Search',
		onsubmit
	}: {
		value?: string;
		placeholder?: string;
		onsubmit: () => void;
	} = $props();

	function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		onsubmit();
	}
</script>

<form
	role="search"
	onsubmit={handleSubmit}
	class="duration-fast flex h-12 items-center gap-3 rounded-full border border-transparent bg-surface px-4 transition-colors focus-within:border-line-strong hover:border-line hover:bg-raised"
>
	<MagnifyingGlassIcon size={20} class="shrink-0 text-dim" />
	<input
		type="search"
		bind:value
		{placeholder}
		autocomplete="off"
		spellcheck="false"
		class="h-full min-w-0 flex-1 bg-transparent text-base text-default outline-none placeholder:text-dim"
	/>
</form>
