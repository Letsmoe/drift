// The library must be imported before anything that pulls in Svelte's client runtime:
// it installs the globals that runtime reads when it loads.
import { start } from '@neoworks/svelte-gpui';
import Boot from './Boot.svelte';

start(
	Boot,
	{},
	{ window: { title: 'Drift', width: 1200, height: 800, minWidth: 720, minHeight: 480 } }
);
