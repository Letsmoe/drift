<!-- Starting, joining and leaving a session, and who is in it. -->
<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { Button, LoadingSpinner } from '@neoworks/svelte-gpui-ui';
	import Label from '#lib/components/Label.svelte';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import { userPlus } from '#lib/layout/chromeIcons.ts';
	import { listenTogether } from '#lib/together/listenTogether.svelte.ts';
	import { colors } from '#lib/theme.ts';

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

{#if !room}
	<View style={{ gap: 12 }}>
		<Label size="body" tone="muted">
			Play the same queue at the same time as your friends. Anyone in the session can play, pause,
			skip and change the queue. You decide who gets in; trusted friends join without asking.
		</Label>
		<Button
			variant="primary"
			size="sm"
			full
			disabled={listenTogether.isPending('host')}
			onpress={() => listenTogether.host()}
		>
			Start a session
		</Button>
	</View>
{:else}
	<View style={{ gap: 12 }}>
		<View style={{ direction: 'row', align: 'center', gap: 8 }}>
			{#if working}
				<LoadingSpinner size={14} color={colors.dim} style={{ shrink: 0 }} />
			{/if}
			<View style={{ flex: 1, minWidth: 0 }}>
				<Label size="body" tone="muted">{statusText}</Label>
			</View>
		</View>
		{#each listenTogether.otherPeers as peer (peer.id)}
			<View style={{ direction: 'row', align: 'center', gap: 12, padding: 4 }}>
				<PeerAvatar name={peer.name} highlighted={peer.userId === room.hostUserId} />
				<View style={{ flex: 1, minWidth: 0 }}>
					<Label size="body" weight="medium" tone="default" truncate>{peer.name}</Label>
					{#if peer.userId === room.hostUserId}
						<Label size="tiny" tone="dim">Host</Label>
					{/if}
				</View>
				{#if !listenTogether.isFriend(peer.userId)}
					<Button
						size="sm"
						variant="ghost"
						icon={userPlus}
						disabled={listenTogether.isPending('add-friend')}
						onpress={() => listenTogether.addFriend(peer.userId, peer.name)}
					>
						Add friend
					</Button>
				{/if}
			</View>
		{/each}
		<View style={{ direction: 'row', justify: 'end' }}>
			<Button
				size="sm"
				variant="ghost"
				disabled={listenTogether.isPending('leave')}
				onpress={() => listenTogether.leave()}
			>
				{leaveLabel()}
			</Button>
		</View>
	</View>
{/if}
