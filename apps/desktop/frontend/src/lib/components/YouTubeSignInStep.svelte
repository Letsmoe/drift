<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { Button, Disclosure, LoadingSpinner } from '@neoworks/svelte-gpui-ui';
	import Label from '#lib/components/Label.svelte';
	import TextArea from '#lib/components/TextArea.svelte';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	const cookieSteps = [
		'Open music.youtube.com and sign in.',
		"Open the browser's DevTools and go to the Network tab.",
		'Reload, click any request to music.youtube.com and copy its Cookie request header.',
		'Paste it here. A whole request or a "Copy as cURL" command works too.'
	];

	let { flow }: { flow: MusicImportFlow } = $props();

	const youtube = $derived(flow.youtube);
</script>

<View style={{ gap: 16 }}>
	<View style={{ gap: 8 }}>
		<Label size="body" weight="semibold">Use a browser where you are signed in</Label>
		<Label size="small" tone="muted">
			If you are already signed in to music.youtube.com in a browser on this computer, Drift can
			pick up that session. This reads the browser's cookie store, and only when you press the
			button.
		</Label>
		<View style={{ alignSelf: 'start' }}>
			<Button variant="primary" disabled={youtube.busy} onpress={() => youtube.searchBrowsers()}>
				Look for signed-in browsers
			</Button>
		</View>

		{#if youtube.busy}
			<View style={{ align: 'center', paddingY: 12, color: colors.text }}><LoadingSpinner /></View>
		{/if}

		{#if youtube.searched}
			{#if youtube.usableSessions.length === 0}
				<Label size="small" tone="muted">No browser with a YouTube Music session was found.</Label>
			{/if}
			{#each youtube.usableSessions as session (session.browser + session.profile)}
				<View
					style={{
						direction: 'row',
						align: 'center',
						gap: 12,
						paddingX: 8,
						paddingY: 8,
						radius: radii.medium,
						transition: hoverTransition
					}}
					hover={{ background: colors.hover }}
				>
					<Label size="small" truncate style={{ flex: 1, minWidth: 0 }}>
						{session.browser}
						<Label size="small" tone="dim">{session.profile}</Label>
					</Label>
					<Button
						size="sm"
						variant="primary"
						disabled={youtube.busy}
						onpress={() => flow.useBrowserSession(session)}
					>
						Use this session
					</Button>
				</View>
			{/each}
			{#if youtube.unusableSessions.length > 0}
				<Disclosure
					title={`${youtube.unusableSessions.length} other profiles without a usable session`}
				>
					<View style={{ gap: 4 }}>
						{#each youtube.unusableSessions as session (session.browser + session.profile)}
							<Label size="tiny" tone="dim">
								{session.browser}
								{session.profile}: {session.reason}
							</Label>
						{/each}
					</View>
				</Disclosure>
			{/if}
		{/if}
	</View>

	<Disclosure title="Paste cookies instead">
		<View style={{ gap: 8 }}>
			<View style={{ gap: 2 }}>
				{#each cookieSteps as step, index (index)}
					<Label size="small" tone="muted">{index + 1}. {step}</Label>
				{/each}
			</View>
			<TextArea
				bind:value={youtube.pastedCookies}
				rows={4}
				mono
				placeholder="SID=...; SAPISID=...; ..."
			/>
			<View style={{ alignSelf: 'start' }}>
				<Button
					disabled={youtube.busy || youtube.pastedCookies.trim() === ''}
					onpress={() => flow.usePastedCookies()}
				>
					Sign in with these cookies
				</Button>
			</View>
		</View>
	</Disclosure>

	<Label size="tiny" tone="dim">
		Your session stays on this computer and is only used to read your library. Sign out any time.
	</Label>

	{#if youtube.errorMessage !== ''}
		<Label size="small" tone="danger">{youtube.errorMessage}</Label>
	{/if}

	<View style={{ direction: 'row', justify: 'end' }}>
		<Button variant="ghost" onpress={() => flow.closeSignIn()}>Back</Button>
	</View>
</View>
