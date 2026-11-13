<!-- Guests waiting for the user to let them into the session. -->
<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { Button } from '@neoworks/svelte-gpui-ui';
	import Label from '#lib/components/Label.svelte';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import { formatUserId, listenTogether } from '#lib/together/listenTogether.svelte.ts';
	import { colors, radii } from '#lib/theme.ts';

	function describeRequester(userId: string): string {
		if (listenTogether.isFriend(userId)) {
			return 'Friend';
		}
		return `Not a friend yet · ${formatUserId(userId)}`;
	}
</script>

<View style={{ gap: 4 }}>
	{#each listenTogether.requests as guest (guest.id)}
		<View style={{ gap: 8, padding: 8, radius: radii.large, background: colors.raised }}>
			<View style={{ direction: 'row', align: 'center', gap: 12 }}>
				<PeerAvatar name={guest.name} highlighted />
				<View style={{ flex: 1, minWidth: 0 }}>
					<Label size="body" tone="muted" truncate>
						<Label weight="semibold" tone="default">{guest.name}</Label> wants to listen with you
					</Label>
					<Label size="tiny" tone="dim" truncate>{describeRequester(guest.userId)}</Label>
				</View>
			</View>
			<View style={{ direction: 'row', justify: 'end', gap: 8 }}>
				<Button
					size="sm"
					variant="ghost"
					disabled={listenTogether.isPending(`decline:${guest.id}`)}
					onpress={() => listenTogether.declineJoin(guest.id)}
				>
					Decline
				</Button>
				<Button
					size="sm"
					disabled={listenTogether.isPending(`accept:${guest.id}`)}
					onpress={() => listenTogether.acceptJoin(guest.id, true)}
				>
					Always let in
				</Button>
				<Button
					size="sm"
					variant="primary"
					disabled={listenTogether.isPending(`accept:${guest.id}`)}
					onpress={() => listenTogether.acceptJoin(guest.id, false)}
				>
					Let in
				</Button>
			</View>
		</View>
	{/each}
</View>
