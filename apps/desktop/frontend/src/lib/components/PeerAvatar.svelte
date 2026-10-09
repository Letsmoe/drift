<script lang="ts">
	let { name, highlighted = false }: { name: string; highlighted?: boolean } = $props();

	const initial = $derived(name.trim().charAt(0).toUpperCase());
</script>

<span
	aria-hidden="true"
	class={[
		'flex size-8 shrink-0 items-center justify-center rounded-full text-sm font-semibold',
		highlighted && 'bg-action text-action-fg',
		!highlighted && 'bg-raised text-default'
	]}
>
	{initial}
</span>
