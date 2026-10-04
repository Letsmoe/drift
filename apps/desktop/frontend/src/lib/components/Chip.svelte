<script lang="ts">
	let {
		label,
		selected = false,
		onclick
	}: { label: string; selected?: boolean; onclick: () => void } = $props();
</script>

<button
	type="button"
	aria-pressed={selected}
	{onclick}
	class={[
		'duration-fast h-8 shrink-0 rounded-full px-3 text-sm font-medium whitespace-nowrap transition-colors',
		selected && 'bg-action text-action-fg',
		!selected && 'bg-raised text-default hover:bg-hover'
	]}
>
	{label}
</button>
