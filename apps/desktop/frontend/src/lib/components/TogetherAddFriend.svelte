<!-- Adds a friend by their ID, or joins their session without adding them. -->
<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { Button } from '@neoworks/svelte-gpui-ui';
	import TextInput from '#lib/layout/TextInput.svelte';
	import { listenTogether } from '#lib/together/listenTogether.svelte.ts';

	let friendId = $state('');
	let friendName = $state('');

	const hasId = $derived(friendId.trim() !== '');

	async function add(): Promise<void> {
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

<View style={{ gap: 8 }}>
	<View style={{ direction: 'row', gap: 8 }}>
		<TextInput
			bind:value={friendId}
			placeholder="Friend's ID"
			mono
			style={{ flex: 1 }}
			onsubmit={add}
		/>
		<TextInput
			bind:value={friendName}
			placeholder="Name (optional)"
			maxLength={40}
			style={{ width: 128 }}
			onsubmit={add}
		/>
	</View>
	<View style={{ direction: 'row', justify: 'end', gap: 8 }}>
		<Button size="sm" variant="ghost" disabled={!hasId} onpress={joinById}>
			Join without adding
		</Button>
		<Button size="sm" disabled={!hasId || listenTogether.isPending('add-friend')} onpress={add}>
			Add friend
		</Button>
	</View>
</View>
