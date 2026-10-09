<script lang="ts">
	import { Checkbox } from '@neoworks-dev/ui';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import CopyIcon from 'phosphor-svelte/lib/CopyIcon';
	import PencilSimpleIcon from 'phosphor-svelte/lib/PencilSimpleIcon';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import { formatUserId, listenTogether } from '#lib/together/listenTogether.svelte.ts';

	const copiedResetMs = 1500;

	const self = $derived(listenTogether.state?.self);
	let editing = $state(false);
	let draftName = $state('');
	let copied = $state(false);

	function startEditing(): void {
		if (!self) {
			return;
		}
		draftName = self.name;
		editing = true;
	}

	async function saveName(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		if (await listenTogether.rename(draftName)) {
			editing = false;
		}
	}

	function cancelOnEscape(event: KeyboardEvent): void {
		if (event.key === 'Escape') {
			event.stopPropagation();
			editing = false;
		}
	}

	function toggleSharing(event: Event): void {
		const checkbox = event.currentTarget as HTMLInputElement;
		listenTogether.setSharePresence(checkbox.checked);
	}

	async function copyId(): Promise<void> {
		if (!self) {
			return;
		}
		await navigator.clipboard.writeText(formatUserId(self.userId));
		copied = true;
		setTimeout(() => (copied = false), copiedResetMs);
	}
</script>

{#if self}
	<div class="flex items-center gap-3">
		<PeerAvatar name={self.name} highlighted />
		<div class="min-w-0 flex-1">
			{#if editing}
				<form onsubmit={saveName}>
					<!-- svelte-ignore a11y_autofocus -->
					<input
						bind:value={draftName}
						aria-label="Your name"
						maxlength="40"
						autofocus
						onkeydown={cancelOnEscape}
						onblur={() => (editing = false)}
						class="h-7 w-full rounded-md bg-raised px-2 text-sm outline-none"
					/>
				</form>
			{:else}
				<button
					type="button"
					onclick={startEditing}
					title="Change the name friends see"
					class="group flex max-w-full items-center gap-1.5 text-sm font-semibold"
				>
					<span class="truncate">{self.name}</span>
					<PencilSimpleIcon size={12} class="shrink-0 text-dim group-hover:text-default" />
				</button>
			{/if}
			<button
				type="button"
				onclick={copyId}
				title="Copy your ID; friends add you with it"
				class="flex items-center gap-1.5 font-mono text-xs text-muted hover:text-default"
			>
				{formatUserId(self.userId)}
				{#if copied}
					<CheckIcon size={12} />
				{:else}
					<CopyIcon size={12} />
				{/if}
			</button>
		</div>
	</div>
	<label class="flex cursor-pointer items-center gap-2 text-xs text-muted">
		<Checkbox
			size="sm"
			checked={listenTogether.sharePresence}
			disabled={listenTogether.isPending('share-presence')}
			onchange={toggleSharing}
		/>
		Show friends what I'm playing
	</label>
{/if}
