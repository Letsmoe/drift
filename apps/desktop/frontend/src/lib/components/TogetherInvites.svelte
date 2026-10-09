<script lang="ts">
	import { Button } from '@neoworks-dev/ui';
	import XIcon from 'phosphor-svelte/lib/XIcon';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import { listenTogether } from '#lib/together/listenTogether.svelte.ts';
</script>

<ul class="flex flex-col gap-1">
	{#each listenTogether.invites as invite (invite.from.userId)}
		<li class="flex items-center gap-3 rounded-lg bg-raised p-2">
			<PeerAvatar name={invite.from.name} highlighted />
			<p class="min-w-0 flex-1 text-sm">
				<span class="font-semibold">{invite.from.name}</span>
				<span class="text-muted">invites you to listen together</span>
			</p>
			<Button
				size="sm"
				variant="primary"
				disabled={listenTogether.isPending(`join:${invite.from.userId}`)}
				onclick={() => listenTogether.join(invite.from.userId)}
			>
				Join
			</Button>
			<Button
				size="sm"
				round
				variant="ghost"
				icon={XIcon}
				onclick={() => listenTogether.dismiss(invite.from.userId)}
			/>
		</li>
	{/each}
</ul>
