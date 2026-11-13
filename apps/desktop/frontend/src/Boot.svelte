<script lang="ts">
	import { Text, View } from '@neoworks/svelte-gpui/ui';
	import type { Component } from 'svelte';
	import { loadBackendAddress } from '#lib/backend/shell.ts';
	import { loadStorage } from '#lib/storage.ts';

	// The host delivers file and command results only once `start` has run, so loading happens in
	// here. Stores such as the panels read their saved values when their module is evaluated, so
	// the app is imported only after the state file and the backend address are loaded.
	let App = $state<Component>();
	let failure = $state('');

	async function boot(): Promise<void> {
		try {
			await Promise.all([loadStorage(), loadBackendAddress()]);
			const loaded = await import('./App.svelte');
			App = loaded.default;
		} catch (error) {
			failure = String(error);
		}
	}

	void boot();
</script>

{#if App}
	<App />
{:else if failure !== ''}
	<View style={{ flex: 1, background: '#141416', padding: 24 }}>
		<Text style={{ color: '#f87171' }}>Drift could not start: {failure}</Text>
	</View>
{:else}
	<View style={{ flex: 1, background: '#141416' }} />
{/if}
