<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { Button } from '@neoworks/svelte-gpui-ui';
	import Label from '#lib/components/Label.svelte';
	import YouTubeAccountPanel from '#lib/components/YouTubeAccountPanel.svelte';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';
	import TextInput from '#lib/layout/TextInput.svelte';
	import { colors } from '#lib/theme.ts';

	let { flow }: { flow: MusicImportFlow } = $props();

	const sectionStyle = { gap: 8, paddingTop: 16, borderTopWidth: 1, borderColor: colors.line };
</script>

<View style={{ gap: 20 }}>
	<YouTubeAccountPanel {flow} />

	<View style={sectionStyle}>
		<Label size="body" weight="semibold">Spotify data export</Label>
		<Label size="small" tone="muted">
			On spotify.com open Account, then Privacy, and request "Account data". Spotify emails the
			download in a few days. Choose the zip you get (<Label
				size="small"
				style={{ fontFamily: 'mono' }}>my_spotify_data.zip</Label
			>), or its extracted folder. It holds your liked songs, playlists, saved albums, followed
			artists and podcasts.
		</Label>
		<Label size="small" tone="muted">
			For a quick start, exportify.net exports your playlists, liked songs and saved albums as a zip
			of CSV files in seconds, with no waiting for Spotify's email.
		</Label>
		<View style={{ direction: 'row', wrap: true, gap: 8 }}>
			<Button variant="primary" onpress={() => flow.chooseExportFile()}>
				Choose zip, JSON or CSV file
			</Button>
			<Button onpress={() => flow.chooseExportFolder()}>Choose extracted folder</Button>
		</View>
	</View>

	<View style={sectionStyle}>
		<Label size="body" weight="semibold">A public playlist or album link</Label>
		<Label size="small" tone="muted">
			Paste an open.spotify.com link. Public links import up to 100 tracks; use the data export for
			the full playlist.
		</Label>
		<View style={{ direction: 'row', gap: 8, align: 'center' }}>
			<TextInput
				bind:value={flow.linkText}
				placeholder="https://open.spotify.com/playlist/..."
				height={44}
				style={{ flex: 1 }}
				onsubmit={() => flow.loadLink()}
			/>
			<Button disabled={flow.linkText.trim() === ''} onpress={() => flow.loadLink()}>Load</Button>
		</View>
	</View>

	{#if flow.errorMessage !== '' && !flow.isAccountSource}
		<Label size="small" tone="danger">{flow.errorMessage}</Label>
	{/if}
</View>
