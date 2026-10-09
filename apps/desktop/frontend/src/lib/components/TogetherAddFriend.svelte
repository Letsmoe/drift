<script lang="ts">
	import { Button } from '@neoworks-dev/ui';
	import { listenTogether } from '#lib/together/listenTogether.svelte.ts';

	let friendId = $state('');
	let friendName = $state('');

	const hasId = $derived(friendId.trim() !== '');

	async function add(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		if (await listenTogether.addFriend(friendId, friendName)) {
			friendId = '';
			friendName = '';
		}
	}

	async function joinById(): Promise<void> {
		if (await listenTogether.join(friendId)) {
			friendId = '';
			friendName = '';
		}
	}
</script>

<form onsubmit={add} class="flex flex-col gap-2">
	<div class="flex gap-2">
		<input
			bind:value={friendId}
			aria-label="Friend's ID"
			placeholder="Friend's ID"
			spellcheck="false"
			autocomplete="off"
			class="h-9 min-w-0 flex-1 rounded-md bg-raised px-3 font-mono text-sm uppercase outline-none placeholder:font-sans placeholder:text-dim placeholder:normal-case"
		/>
		<input
			bind:value={friendName}
			aria-label="Name (optional)"
			placeholder="Name (optional)"
			maxlength="40"
			class="h-9 w-32 rounded-md bg-raised px-3 text-sm outline-none placeholder:text-dim"
		/>
	</div>
	<div class="flex justify-end gap-2">
		<Button size="sm" variant="ghost" disabled={!hasId} onclick={joinById}
			>Join without adding</Button
		>
		<Button size="sm" type="submit" disabled={!hasId || listenTogether.isPending('add-friend')}>
			Add friend
		</Button>
	</div>
</form>
