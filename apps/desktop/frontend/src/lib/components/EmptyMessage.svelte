<script lang="ts">
	import type { Snippet } from 'svelte';
	import Label from '#lib/components/Label.svelte';

	let { children }: { children: Snippet } = $props();
</script>

<Label size="small" tone="dim" style={{ textAlign: 'center', paddingY: 48 }}>
	{@render children()}
</Label>
