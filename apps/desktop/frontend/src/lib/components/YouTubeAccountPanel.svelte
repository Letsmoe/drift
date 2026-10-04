<script lang="ts">
	import { Button } from '@neoworks-dev/ui';
	import { formatDate } from '#lib/format.ts';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';

	let { flow }: { flow: MusicImportFlow } = $props();

	const youtube = $derived(flow.youtube);
	const lastSynced = $derived(formatDate(youtube.status?.lastSyncedAt));
</script>

<section class="flex flex-col gap-3">
	<div class="flex items-center gap-2">
		<h3 class="text-base font-semibold">YouTube Music account</h3>
		<span class="rounded-full bg-raised px-2 py-0.5 text-xs text-muted">Recommended</span>
	</div>

	{#if youtube.signedIn && youtube.status}
		<div class="flex items-center gap-3 rounded-lg bg-raised p-3">
			{#if youtube.status.account.photoUrl !== ''}
				<img src={youtube.status.account.photoUrl} alt="" class="size-10 rounded-full" />
			{/if}
			<div class="min-w-0 flex-1">
				<p class="truncate text-sm font-semibold">Signed in as {youtube.status.account.name}</p>
				<p class="truncate text-xs text-muted">
					{#if youtube.status.account.handle !== ''}{youtube.status.account.handle} ·
					{/if}{#if lastSynced !== ''}Last synced {lastSynced}{:else}Not synced yet{/if}
				</p>
			</div>
			<Button size="sm" variant="ghost" onclick={() => youtube.signOut()}>Sign out</Button>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button variant="primary" onclick={() => flow.syncNow()}>Sync now</Button>
			<Button onclick={() => flow.loadAccountPreview()}>Choose what to import</Button>
		</div>
		<p class="text-xs text-dim">
			Your session stays on this computer and is only used to read your library. Sign out any time.
		</p>
	{:else}
		<p class="text-sm text-muted">
			Sign in to YouTube Music once and bring over your liked songs, playlists, albums, artists and
			podcasts. Sync again any time to pick up changes.
		</p>
		<div class="flex flex-wrap gap-2">
			<Button variant="primary" onclick={() => flow.openSignIn()}>Sign in to YouTube Music</Button>
		</div>
		<p class="text-xs text-dim">
			Coming from Spotify? Copy your library to YouTube Music with TuneMyMusic or Soundiiz first,
			then import here.
		</p>
	{/if}

	{#if flow.errorMessage !== '' && flow.isAccountSource}
		<p class="text-sm text-red">{flow.errorMessage}</p>
	{/if}
	{#if youtube.errorMessage !== ''}
		<p class="text-sm text-red">{youtube.errorMessage}</p>
	{/if}
	{#if flow.isAccountSource && flow.errorMessage.includes('sign in again')}
		<div class="self-start">
			<Button size="sm" onclick={() => flow.openSignIn()}>Sign in again</Button>
		</div>
	{/if}
</section>
