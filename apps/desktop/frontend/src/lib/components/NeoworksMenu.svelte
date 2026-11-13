<!-- The Neoworks account button of the top bar and the panel for signing in, syncing and signing out. -->
<script lang="ts">
	import {
		Button as PressableButton,
		Icon,
		Popover,
		Tooltip,
		View
	} from '@neoworks/svelte-gpui/ui';
	import { Button, LoadingSpinner } from '@neoworks/svelte-gpui-ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Label from '#lib/components/Label.svelte';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import { userCircle } from '#lib/layout/chromeIcons.ts';
	import { neoworksAccount } from '#lib/neoworks/neoworksAccount.svelte.ts';
	import { colors, floatingPanel, glassBackground, hoverTransition } from '#lib/theme.ts';

	neoworksAccount.start();

	let open = $state(false);

	const buttonLabel = $derived(describeButton());
	const syncLine = $derived(describeSync());

	function describeButton(): string {
		if (neoworksAccount.signedIn) {
			return `Neoworks: ${neoworksAccount.displayName}`;
		}
		return 'Sign in with Neoworks';
	}

	function describeSync(): string {
		const status = neoworksAccount.status;
		if (!status) {
			return '';
		}
		if (status.syncing) {
			return 'Syncing…';
		}
		if (status.lastSyncedAt === null) {
			return 'Not synced yet';
		}
		return `Synced at ${formatTime(status.lastSyncedAt)}`;
	}

	function formatTime(timestamp: string): string {
		const date = new Date(timestamp);
		const hours = String(date.getHours()).padStart(2, '0');
		const minutes = String(date.getMinutes()).padStart(2, '0');
		return `${hours}:${minutes}`;
	}
</script>

{#if neoworksAccount.available}
	<Popover
		bind:open
		side="bottom"
		align="end"
		offset={8}
		style={{ ...floatingPanel, width: 320, padding: 16 }}
	>
		{#snippet trigger({ toggle })}
			<Tooltip text={buttonLabel} side="bottom">
				<PressableButton
					style={{
						size: 36,
						radius: 9999,
						align: 'center',
						justify: 'center',
						background: glassBackground,
						color: appearance.dim,
						transition: hoverTransition
					}}
					hover={{ background: colors.hover, color: colors.text }}
					onpress={toggle}
				>
					{#if neoworksAccount.signedIn}
						<PeerAvatar name={neoworksAccount.displayName} />
					{:else}
						<Icon svg={userCircle} size={22} />
					{/if}
				</PressableButton>
			</Tooltip>
		{/snippet}
		<View style={{ gap: 16 }}>
			{#if neoworksAccount.signedIn}
				<View style={{ direction: 'row', align: 'center', gap: 12 }}>
					<PeerAvatar name={neoworksAccount.displayName} highlighted />
					<View style={{ flex: 1, minWidth: 0 }}>
						<Label size="body" weight="semibold" tone="default" truncate>
							{neoworksAccount.displayName}
						</Label>
						{#if neoworksAccount.status?.name !== ''}
							<Label size="small" tone="dim" truncate>{neoworksAccount.status?.email}</Label>
						{/if}
					</View>
				</View>
				<View style={{ direction: 'row', align: 'center', gap: 8 }}>
					{#if neoworksAccount.status?.syncing}
						<LoadingSpinner size={14} label="Syncing" color={colors.dim} />
					{/if}
					<Label size="small" tone="dim">{syncLine}</Label>
				</View>
				<View style={{ direction: 'row', justify: 'end', gap: 8 }}>
					<Button size="sm" variant="ghost" onpress={() => neoworksAccount.signOut()}>
						Sign out
					</Button>
					<Button
						size="sm"
						disabled={neoworksAccount.status?.syncing}
						onpress={() => neoworksAccount.sync()}
					>
						Sync now
					</Button>
				</View>
			{:else if neoworksAccount.signingIn}
				<View style={{ direction: 'row', align: 'center', gap: 12 }}>
					<LoadingSpinner size={18} label="Waiting for the browser" color={colors.dim} />
					<Label size="small" tone="default">Finish signing in in your browser.</Label>
				</View>
				<View style={{ direction: 'row', justify: 'end' }}>
					<Button size="sm" variant="ghost" onpress={() => neoworksAccount.cancelSignIn()}>
						Cancel
					</Button>
				</View>
			{:else}
				<View style={{ gap: 4 }}>
					<Label size="medium" weight="bold" tone="default">Sync with Neoworks</Label>
					<Label size="small" tone="dim">
						Keep your likes, playlists and history the same on all your devices. They are end to end
						encrypted, so Neoworks cannot read them.
					</Label>
				</View>
				<Button variant="primary" full onpress={() => neoworksAccount.signIn()}>
					Sign in with Neoworks
				</Button>
			{/if}

			{#if neoworksAccount.status?.keyInFile}
				<Label size="tiny" tone="dim">
					No system keyring answered, so the key that protects your Neoworks sign-in is kept in a
					file in Drift's data folder.
				</Label>
			{/if}

			{#if neoworksAccount.errorMessage !== ''}
				<Label size="small" tone="danger">{neoworksAccount.errorMessage}</Label>
			{/if}
		</View>
	</Popover>
{/if}
