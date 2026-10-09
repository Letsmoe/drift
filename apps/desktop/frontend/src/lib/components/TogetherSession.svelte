<script lang="ts">
	import { Button, LoadingSpinner } from '@neoworks-dev/ui';
	import UserPlusIcon from 'phosphor-svelte/lib/UserPlusIcon';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import { listenTogether } from '#lib/together/listenTogether.svelte.ts';

	const room = $derived(listenTogether.room);
	const status = $derived(listenTogether.roomStatus);
	const hosting = $derived(room !== null && room.hosting);
	const statusText = $derived(describeStatus());
	const working = $derived(
		status === 'waiting' || status === 'connecting' || status === 'awaiting'
	);

	function describeStatus(): string {
		const hostName = listenTogether.hostName;
		if (status === 'waiting') {
			return 'Waiting for friends to join';
		}
		if (status === 'connecting') {
			return `Connecting to ${hostName}…`;
		}
		if (status === 'awaiting') {
			return `Waiting for ${hostName} to let you in…`;
		}
		if (status === 'declined') {
			return `${hostName} did not let you in.`;
		}
		if (status === 'closed') {
			return `${hostName} ended the session.`;
		}
		if (status === 'failed') {
			return `Could not reach ${hostName}. They need Drift open with a session started.`;
		}
		if (hosting) {
			return 'Friends in your session hear what you hear';
		}
		return `Listening along with ${hostName}`;
	}

	function leaveLabel(): string {
		if (status === 'closed' || status === 'failed' || status === 'declined') {
			return 'Close';
		}
		if (hosting) {
			return 'End session';
		}
		return 'Leave';
	}
</script>

<section class="flex flex-col gap-3">
	{#if !room}
		<p class="text-sm text-muted">
			Play the same queue at the same time as your friends. Anyone in the session can play, pause,
			skip and change the queue. You decide who gets in; trusted friends join without asking.
		</p>
		<Button
			variant="primary"
			size="sm"
			disabled={listenTogether.isPending('host')}
			onclick={() => listenTogether.host()}
		>
			Start a session
		</Button>
	{:else}
		<div class="flex items-center gap-2 text-sm">
			{#if working}
				<LoadingSpinner size={14} class="shrink-0 text-dim" />
			{/if}
			<p class="flex-1 text-muted">{statusText}</p>
		</div>
		{#if listenTogether.otherPeers.length > 0}
			<ul class="flex flex-col gap-1">
				{#each listenTogether.otherPeers as peer (peer.id)}
					<li class="flex items-center gap-3 rounded-md px-1 py-1">
						<PeerAvatar name={peer.name} highlighted={peer.userId === room.hostUserId} />
						<div class="min-w-0 flex-1">
							<p class="truncate text-sm font-medium">{peer.name}</p>
							{#if peer.userId === room.hostUserId}
								<p class="text-xs text-dim">Host</p>
							{/if}
						</div>
						{#if !listenTogether.isFriend(peer.userId)}
							<Button
								size="sm"
								variant="ghost"
								icon={UserPlusIcon}
								disabled={listenTogether.isPending('add-friend')}
								onclick={() => listenTogether.addFriend(peer.userId, peer.name)}
							>
								Add friend
							</Button>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
		<div class="flex justify-end">
			<Button
				size="sm"
				variant="ghost"
				disabled={listenTogether.isPending('leave')}
				onclick={() => listenTogether.leave()}
			>
				{leaveLabel()}
			</Button>
		</div>
	{/if}
</section>
