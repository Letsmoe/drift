<script lang="ts">
	import { Button } from '@neoworks-dev/ui';
	import YouTubeAccountPanel from '#lib/components/YouTubeAccountPanel.svelte';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';

	let { flow }: { flow: MusicImportFlow } = $props();

	function submitLink(event: SubmitEvent) {
		event.preventDefault();
		flow.loadLink();
	}
</script>

<div class="flex flex-col gap-5">
	<YouTubeAccountPanel {flow} />

	<section class="flex flex-col gap-2 border-t border-line pt-4">
		<h3 class="text-base font-semibold">Spotify data export</h3>
		<p class="text-sm text-muted">
			On spotify.com open Account, then Privacy, and request "Account data". Spotify emails the
			download in a few days. Choose the zip you get (<code>my_spotify_data.zip</code>), or its
			extracted folder. It holds your liked songs, playlists, saved albums, followed artists and
			podcasts.
		</p>
		<div class="flex flex-wrap gap-2">
			<Button variant="primary" onclick={() => flow.chooseExportFile()}
				>Choose zip or JSON file</Button
			>
			<Button onclick={() => flow.chooseExportFolder()}>Choose extracted folder</Button>
		</div>
	</section>

	<section class="flex flex-col gap-2 border-t border-line pt-4">
		<h3 class="text-base font-semibold">A public playlist or album link</h3>
		<p class="text-sm text-muted">
			Paste an open.spotify.com link. Public links import up to 100 tracks; use the data export for
			the full playlist.
		</p>
		<form onsubmit={submitLink} class="flex gap-2">
			<input
				bind:value={flow.linkText}
				aria-label="Spotify playlist or album link"
				placeholder="https://open.spotify.com/playlist/..."
				class="h-11 min-w-0 flex-1 rounded-md bg-raised px-3 text-base outline-none placeholder:text-dim focus:border-line-strong"
			/>
			<Button type="submit" disabled={flow.linkText.trim() === ''}>Load</Button>
		</form>
	</section>

	{#if flow.errorMessage !== '' && !flow.isAccountSource}
		<p class="text-sm text-red">{flow.errorMessage}</p>
	{/if}
</div>
