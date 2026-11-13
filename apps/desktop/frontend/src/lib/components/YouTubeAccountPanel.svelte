<script lang="ts">
	import { Image, View } from '@neoworks/svelte-gpui/ui';
	import { Button } from '@neoworks/svelte-gpui-ui';
	import Label from '#lib/components/Label.svelte';
	import { formatDate } from '#lib/format.ts';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';
	import { colors, radii } from '#lib/theme.ts';

	let { flow }: { flow: MusicImportFlow } = $props();

	const youtube = $derived(flow.youtube);
	const lastSynced = $derived(formatDate(youtube.status?.lastSyncedAt));
	const accountDetails = $derived(accountDetailsOf());

	function accountDetailsOf(): string {
		if (!youtube.status) {
			return '';
		}
		const parts: string[] = [];
		if (youtube.status.account.handle !== '') {
			parts.push(youtube.status.account.handle);
		}
		if (lastSynced !== '') {
			parts.push(`Last synced ${lastSynced}`);
		} else {
			parts.push('Not synced yet');
		}
		return parts.join(' · ');
	}
</script>

<View style={{ gap: 12 }}>
	<View style={{ direction: 'row', align: 'center', gap: 8 }}>
		<Label size="body" weight="semibold">YouTube Music account</Label>
		<Label
			size="tiny"
			tone="muted"
			style={{ paddingX: 8, paddingY: 2, radius: radii.full, background: colors.raised }}
		>
			Recommended
		</Label>
	</View>

	{#if youtube.signedIn && youtube.status}
		<View
			style={{
				direction: 'row',
				align: 'center',
				gap: 12,
				padding: 12,
				radius: radii.large,
				background: colors.raised
			}}
		>
			{#if youtube.status.account.photoUrl !== ''}
				<Image
					src={youtube.status.account.photoUrl}
					fit="cover"
					style={{ size: 40, shrink: 0, radius: 20 }}
				/>
			{/if}
			<View style={{ flex: 1, minWidth: 0 }}>
				<Label size="small" weight="semibold" truncate>
					Signed in as {youtube.status.account.name}
				</Label>
				<Label size="tiny" tone="muted" truncate>{accountDetails}</Label>
			</View>
			<Button size="sm" variant="ghost" onpress={() => youtube.signOut()}>Sign out</Button>
		</View>
		<View style={{ direction: 'row', wrap: true, gap: 8 }}>
			<Button variant="primary" onpress={() => flow.syncNow()}>Sync now</Button>
			<Button onpress={() => flow.loadAccountPreview()}>Choose what to import</Button>
		</View>
		<Label size="tiny" tone="dim">
			Your session stays on this computer and is only used to read your library. Sign out any time.
		</Label>
	{:else}
		<Label size="small" tone="muted">
			Sign in to YouTube Music once and bring over your liked songs, playlists, albums, artists and
			podcasts. Sync again any time to pick up changes.
		</Label>
		<View style={{ direction: 'row', wrap: true, gap: 8 }}>
			<Button variant="primary" onpress={() => flow.openSignIn()}>Sign in to YouTube Music</Button>
		</View>
		<Label size="tiny" tone="dim">
			Coming from Spotify? Copy your library to YouTube Music with TuneMyMusic or Soundiiz first,
			then import here.
		</Label>
	{/if}

	{#if flow.errorMessage !== '' && flow.isAccountSource}
		<Label size="small" tone="danger">{flow.errorMessage}</Label>
	{/if}
	{#if youtube.errorMessage !== ''}
		<Label size="small" tone="danger">{youtube.errorMessage}</Label>
	{/if}
	{#if flow.isAccountSource && flow.errorMessage.includes('sign in again')}
		<View style={{ alignSelf: 'start' }}>
			<Button size="sm" onpress={() => flow.openSignIn()}>Sign in again</Button>
		</View>
	{/if}
</View>
