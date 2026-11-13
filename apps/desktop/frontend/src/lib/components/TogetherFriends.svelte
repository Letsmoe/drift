<!-- The friend list: who is online and what they play, with invite, join, trust and remove. -->
<script lang="ts">
	import { Icon, Tooltip, View } from '@neoworks/svelte-gpui/ui';
	import { Button } from '@neoworks/svelte-gpui-ui';
	import type { TogetherFriend } from '#lib/api/together.ts';
	import FriendPresence from '#lib/components/FriendPresence.svelte';
	import IconAction from '#lib/components/IconAction.svelte';
	import Label from '#lib/components/Label.svelte';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import { check, shieldCheck, shieldCheckFill, trash } from '#lib/layout/chromeIcons.ts';
	import { formatUserId, listenTogether } from '#lib/together/listenTogether.svelte.ts';
	import { colors, hoverTransition, radii } from '#lib/theme.ts';

	// Online friends first, those with an open session before the rest.
	const sortedFriends = $derived([...listenTogether.friends].sort(compareFriends));
	let hoveredUserId = $state('');

	function compareFriends(left: TogetherFriend, right: TogetherFriend): number {
		return onlineRank(left) - onlineRank(right);
	}

	function onlineRank(friend: TogetherFriend): number {
		if (friend.presence?.sessionOpen) {
			return 0;
		}
		if (friend.presence) {
			return 1;
		}
		return 2;
	}

	function inRoomWith(friend: TogetherFriend): boolean {
		const room = listenTogether.room;
		if (!room || !listenTogether.roomOpen) {
			return false;
		}
		return room.peers.some((peer) => peer.userId === friend.userId);
	}

	function inTheirRoom(friend: TogetherFriend): boolean {
		const room = listenTogether.room;
		return room !== null && listenTogether.roomOpen && room.hostUserId === friend.userId;
	}

	function displayName(friend: TogetherFriend): string {
		if (friend.name !== '') {
			return friend.name;
		}
		return formatUserId(friend.userId);
	}

	function trustLabel(friend: TogetherFriend): string {
		if (friend.trusted) {
			return `${displayName(friend)} joins your sessions without asking. Click to ask every time.`;
		}
		return `Let ${displayName(friend)} join your sessions without asking`;
	}

	function joinVariant(friend: TogetherFriend): 'primary' | 'ghost' {
		if (friend.presence?.sessionOpen) {
			return 'primary';
		}
		return 'ghost';
	}

	function rowBackground(friend: TogetherFriend): string {
		if (hoveredUserId === friend.userId) {
			return colors.hover;
		}
		return 'transparent';
	}

	function trustOpacity(friend: TogetherFriend): number {
		if (friend.trusted || hoveredUserId === friend.userId) {
			return 1;
		}
		return 0;
	}

	function removeOpacity(friend: TogetherFriend): number {
		if (hoveredUserId === friend.userId) {
			return 1;
		}
		return 0;
	}

	function trustColor(friend: TogetherFriend): string {
		if (friend.trusted) {
			return colors.text;
		}
		return colors.faint;
	}

	function trustIcon(friend: TogetherFriend): string {
		if (friend.trusted) {
			return shieldCheckFill;
		}
		return shieldCheck;
	}
</script>

{#snippet sessionActions(friend: TogetherFriend)}
	{#if inRoomWith(friend)}
		<View style={{ paddingX: 8 }}><Label size="tiny" tone="muted">Listening</Label></View>
	{:else}
		{#if listenTogether.invitedUserId === friend.userId}
			<View style={{ direction: 'row', align: 'center', gap: 4, paddingX: 8, color: colors.muted }}>
				<Icon svg={check} size={12} />
				<Label size="tiny" tone="muted">Invited</Label>
			</View>
		{:else}
			<Button
				size="sm"
				variant="ghost"
				disabled={listenTogether.isPending(`invite:${friend.userId}`)}
				onpress={() => listenTogether.invite(friend.userId)}
			>
				Invite
			</Button>
		{/if}
		{#if !inTheirRoom(friend)}
			<Button
				size="sm"
				variant={joinVariant(friend)}
				disabled={listenTogether.isPending(`join:${friend.userId}`)}
				onpress={() => listenTogether.join(friend.userId)}
			>
				Join
			</Button>
		{/if}
	{/if}
{/snippet}

{#if listenTogether.friends.length === 0}
	<Label size="body" tone="muted">
		Add friends by their ID to see what they play, invite them or join their sessions. Share yours
		from the top of this panel.
	</Label>
{:else}
	<View>
		{#each sortedFriends as friend (friend.userId)}
			<View
				style={{
					direction: 'row',
					align: 'center',
					gap: 8,
					padding: 4,
					radius: radii.medium,
					background: rowBackground(friend),
					transition: hoverTransition
				}}
				onhover={(hovered) => (hoveredUserId = hovered ? friend.userId : '')}
			>
				<View style={{ shrink: 0 }}>
					<PeerAvatar name={displayName(friend)} />
					{#if friend.presence}
						<View
							style={{
								position: 'absolute',
								right: -2,
								bottom: -2,
								size: 12,
								radius: 9999,
								borderWidth: 2,
								borderColor: colors.elevated,
								background: colors.green
							}}
						/>
					{/if}
				</View>
				<View style={{ flex: 1, minWidth: 0, paddingLeft: 4 }}>
					<Label size="body" weight="medium" tone="default" truncate>{displayName(friend)}</Label>
					<FriendPresence presence={friend.presence} userIdText={formatUserId(friend.userId)} />
				</View>
				{@render sessionActions(friend)}
				<View style={{ opacity: trustOpacity(friend) }}>
					<Tooltip text={trustLabel(friend)} side="top">
						<IconAction
							icon={trustIcon(friend)}
							iconSize={16}
							boxSize={32}
							color={trustColor(friend)}
							hoverBackground={colors.pressed}
							disabled={listenTogether.isPending(`trust:${friend.userId}`)}
							onpress={() => listenTogether.setTrusted(friend.userId, !friend.trusted)}
						/>
					</Tooltip>
				</View>
				<View style={{ opacity: removeOpacity(friend) }}>
					<Tooltip text="Remove friend" side="top">
						<IconAction
							icon={trash}
							iconSize={16}
							boxSize={32}
							hoverBackground={colors.pressed}
							disabled={listenTogether.isPending(`remove:${friend.userId}`)}
							onpress={() => listenTogether.removeFriend(friend.userId)}
						/>
					</Tooltip>
				</View>
			</View>
		{/each}
	</View>
{/if}
