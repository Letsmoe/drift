<script lang="ts">
	import { Button } from '@neoworks-dev/ui';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import { formatUserId, listenTogether } from '#lib/together/listenTogether.svelte.ts';

	function describeRequester(userId: string): string {
		if (listenTogether.isFriend(userId)) {
			return 'Friend';
		}
		return `Not a friend yet · ${formatUserId(userId)}`;
	}
</script>

<ul class="flex flex-col gap-1">
	{#each listenTogether.requests as guest (guest.id)}
		<li class="flex flex-col gap-2 rounded-lg bg-raised p-2">
			<div class="flex items-center gap-3">
				<PeerAvatar name={guest.name} highlighted />
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm">
						<span class="font-semibold">{guest.name}</span>
						<span class="text-muted">wants to listen with you</span>
					</p>
					<p class="truncate text-xs text-dim">{describeRequester(guest.userId)}</p>
				</div>
			</div>
			<div class="flex justify-end gap-2">
				<Button
					size="sm"
					variant="ghost"
					disabled={listenTogether.isPending(`decline:${guest.id}`)}
					onclick={() => listenTogether.declineJoin(guest.id)}
				>
					Decline
				</Button>
				<Button
					size="sm"
					disabled={listenTogether.isPending(`accept:${guest.id}`)}
					onclick={() => listenTogether.acceptJoin(guest.id, true)}
				>
					Always let in
				</Button>
				<Button
					size="sm"
					variant="primary"
					disabled={listenTogether.isPending(`accept:${guest.id}`)}
					onclick={() => listenTogether.acceptJoin(guest.id, false)}
				>
					Let in
				</Button>
			</div>
		</li>
	{/each}
</ul>
