<script lang="ts">
	import { Button, LoadingSpinner } from '@neoworks-dev/ui';
	import UserCircleIcon from 'phosphor-svelte/lib/UserCircleIcon';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import { neoworksAccount } from '#lib/neoworks/neoworksAccount.svelte.ts';

	neoworksAccount.start();

	let open = $state(false);
	let container = $state<HTMLElement>();

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
		const time = new Date(status.lastSyncedAt).toLocaleTimeString('en', {
			hour: '2-digit',
			minute: '2-digit'
		});
		return `Synced at ${time}`;
	}

	function closeOnOutsidePointer(event: PointerEvent): void {
		if (open && container && !container.contains(event.target as Node)) {
			open = false;
		}
	}

	function closeOnEscape(event: KeyboardEvent): void {
		if (open && event.key === 'Escape') {
			open = false;
		}
	}
</script>

<svelte:window onpointerdown={closeOnOutsidePointer} onkeydown={closeOnEscape} />

{#if neoworksAccount.available}
	<div bind:this={container} class="relative">
		<button
			type="button"
			aria-label={buttonLabel}
			title={buttonLabel}
			aria-haspopup="dialog"
			aria-expanded={open}
			onclick={() => (open = !open)}
			class="duration-fast glass flex size-9 items-center justify-center rounded-full text-dim transition-colors hover:bg-hover hover:text-default"
		>
			{#if neoworksAccount.signedIn}
				<PeerAvatar name={neoworksAccount.displayName} />
			{:else}
				<UserCircleIcon size={22} />
			{/if}
		</button>

		{#if open}
			<div
				role="dialog"
				aria-label="Neoworks account"
				class="absolute top-full right-0 z-40 mt-2 flex w-80 flex-col gap-4 rounded-xl border border-line bg-elevated p-4 shadow-lg"
			>
				{#if neoworksAccount.signedIn}
					<header class="flex items-center gap-3">
						<PeerAvatar name={neoworksAccount.displayName} highlighted />
						<div class="flex min-w-0 flex-col">
							<span class="truncate font-semibold">{neoworksAccount.displayName}</span>
							{#if neoworksAccount.status?.name !== ''}
								<span class="truncate text-sm text-dim">{neoworksAccount.status?.email}</span>
							{/if}
						</div>
					</header>
					<div class="flex items-center gap-2 text-sm text-dim">
						{#if neoworksAccount.status?.syncing}
							<LoadingSpinner size={14} label="Syncing" />
						{/if}
						<span>{syncLine}</span>
					</div>
					<div class="flex justify-end gap-2">
						<Button size="sm" variant="ghost" onclick={() => neoworksAccount.signOut()}
							>Sign out</Button
						>
						<Button
							size="sm"
							disabled={neoworksAccount.status?.syncing}
							onclick={() => neoworksAccount.sync()}>Sync now</Button
						>
					</div>
				{:else if neoworksAccount.signingIn}
					<div class="flex items-center gap-3">
						<LoadingSpinner size={18} label="Waiting for the browser" />
						<p class="text-sm">Finish signing in in your browser.</p>
					</div>
					<div class="flex justify-end">
						<Button size="sm" variant="ghost" onclick={() => neoworksAccount.cancelSignIn()}
							>Cancel</Button
						>
					</div>
				{:else}
					<header class="flex flex-col gap-1">
						<h2 class="text-base font-bold">Sync with Neoworks</h2>
						<p class="text-sm text-dim">
							Keep your likes, playlists and history the same on all your devices. They are end to
							end encrypted, so Neoworks cannot read them.
						</p>
					</header>
					<Button variant="primary" full onclick={() => neoworksAccount.signIn()}
						>Sign in with Neoworks</Button
					>
				{/if}

				{#if neoworksAccount.status?.keyInFile}
					<p class="text-xs text-dim">
						No system keyring answered, so the key that protects your Neoworks sign-in is kept in a
						file in Drift's data folder.
					</p>
				{/if}

				{#if neoworksAccount.errorMessage !== ''}
					<p class="text-sm text-red">{neoworksAccount.errorMessage}</p>
				{/if}
			</div>
		{/if}
	</div>
{/if}
