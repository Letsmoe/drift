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
	class="duration-fast flex h-11 items-center gap-2 rounded-lg border border-line bg-input px-3 transition-colors focus-within:border-line-strong"
>
	<MagnifyingGlassIcon size={16} class="shrink-0 text-dim" />
	<input
		type="search"
		bind:value
		{placeholder}
		autocomplete="off"
		spellcheck="false"
		class="h-full min-w-0 flex-1 bg-transparent text-sm text-default placeholder:text-faint"
	/>
</form>
