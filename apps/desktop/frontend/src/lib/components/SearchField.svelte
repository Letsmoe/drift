<script lang="ts">
	import MagnifyingGlassIcon from 'phosphor-svelte/lib/MagnifyingGlassIcon';
	import XIcon from 'phosphor-svelte/lib/XIcon';

	let {
		value = $bindable(''),
		placeholder = 'Search',
		expanded = false,
		onsubmit,
		onfocus = undefined,
		oninput = undefined,
		onkeydown = undefined
	}: {
		value?: string;
		placeholder?: string;
		/** Whether a suggestions list is open under the field. */
		expanded?: boolean;
		onsubmit: () => void;
		onfocus?: () => void;
		oninput?: () => void;
		onkeydown?: (event: KeyboardEvent) => void;
	} = $props();

	let inputElement = $state<HTMLInputElement>();

	function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		onsubmit();
	}

	function clear() {
		value = '';
		oninput?.();
		inputElement?.focus();
	}
</script>

<form
	role="search"
	onsubmit={handleSubmit}
	class="duration-fast flex h-12 items-center gap-3 rounded-full border border-transparent bg-surface px-4 transition-colors focus-within:border-line-strong hover:border-line hover:bg-raised"
>
	<MagnifyingGlassIcon size={20} class="shrink-0 text-dim" />
	<input
		bind:this={inputElement}
		type="text"
		role="combobox"
		aria-expanded={expanded}
		aria-controls="search-suggestions"
		aria-autocomplete="list"
		bind:value
		{placeholder}
		{onfocus}
		{oninput}
		{onkeydown}
		autocomplete="off"
		spellcheck="false"
		class="h-full min-w-0 flex-1 bg-transparent text-base text-default outline-none placeholder:text-dim"
	/>
	{#if value !== ''}
		<button
			type="button"
			aria-label="Clear search"
			onclick={clear}
			class="shrink-0 text-dim hover:text-default"
		>
			<XIcon size={18} />
		</button>
	{/if}
</form>
