<script lang="ts">
	import { Button } from '@neoworks-dev/ui';
	import CheckIcon from 'phosphor-svelte/lib/CheckIcon';
	import ShieldCheckIcon from 'phosphor-svelte/lib/ShieldCheckIcon';
	import TrashIcon from 'phosphor-svelte/lib/TrashIcon';
	import FriendPresence from '#lib/components/FriendPresence.svelte';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import type { TogetherFriend } from '#lib/api/together.ts';
	import { formatUserId, listenTogether } from '#lib/together/listenTogether.svelte.ts';

	// Online friends first, those with an open session before the rest.
	const sortedFriends = $derived(listenTogether.friends.toSorted(compareFriends));

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
</script>

{#snippet trustToggle(friend: TogetherFriend)}
	<button
		type="button"
		aria-label={trustLabel(friend)}
		aria-pressed={friend.trusted}
		title={trustLabel(friend)}
		disabled={listenTogether.isPending(`trust:${friend.userId}`)}
		onclick={() => listenTogether.setTrusted(friend.userId, !friend.trusted)}
		class={[
			'duration-fast flex size-8 shrink-0 items-center justify-center rounded-full transition-colors hover:bg-hover',
			friend.trusted && 'text-default',
			!friend.trusted && 'text-faint opacity-0 group-hover:opacity-100 focus-visible:opacity-100'
		]}
	>
		<ShieldCheckIcon size={16} weight={friend.trusted ? 'fill' : 'regular'} />
	</button>
{/snippet}

{#snippet sessionActions(friend: TogetherFriend)}
	{#if inRoomWith(friend)}
		<span class="px-2 text-xs text-muted">Listening</span>
	{:else}
		{#if listenTogether.invitedUserId === friend.userId}
			<span class="flex items-center gap-1 px-2 text-xs text-muted">
				<CheckIcon size={12} /> Invited
			</span>
		{:else}
			<Button
				size="sm"
				variant="ghost"
				disabled={listenTogether.isPending(`invite:${friend.userId}`)}
				onclick={() => listenTogether.invite(friend.userId)}
			>
				Invite
			</Button>
		{/if}
		{#if !inTheirRoom(friend)}
			<Button
				size="sm"
				variant={friend.presence?.sessionOpen ? 'primary' : 'ghost'}
				disabled={listenTogether.isPending(`join:${friend.userId}`)}
				onclick={() => listenTogether.join(friend.userId)}
			>
				Join
			</Button>
		{/if}
	{/if}
{/snippet}

{#if listenTogether.friends.length === 0}
	<p class="text-sm text-muted">
		Add friends by their ID to see what they play, invite them or join their sessions. Share yours
		from the top of this panel.
	</p>
{:else}
	<ul class="flex flex-col">
		{#each sortedFriends as friend (friend.userId)}
			<li class="group flex items-center gap-2 rounded-md px-1 py-1.5 hover:bg-hover">
				<span class="relative">
					<PeerAvatar name={displayName(friend)} />
					{#if friend.presence}
						<span
							class="absolute -right-0.5 -bottom-0.5 size-2.5 rounded-full border-2 border-elevated bg-green"
						></span>
					{/if}
				</span>
				<div class="min-w-0 flex-1 pl-1">
					<p class="truncate text-sm font-medium">{displayName(friend)}</p>
					<FriendPresence presence={friend.presence} userIdText={formatUserId(friend.userId)} />
				</div>
				{@render sessionActions(friend)}
				{@render trustToggle(friend)}
				<span class="opacity-0 group-hover:opacity-100 focus-within:opacity-100">
					<Button
						size="sm"
						round
						variant="ghost"
						icon={TrashIcon}
						disabled={listenTogether.isPending(`remove:${friend.userId}`)}
						onclick={() => listenTogether.removeFriend(friend.userId)}
					/>
				</span>
			</li>
		{/each}
	</ul>
{/if}
