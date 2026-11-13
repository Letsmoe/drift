<!-- Friends who invited the user to a session. -->
<script lang="ts">
	import { View } from '@neoworks/svelte-gpui/ui';
	import { Button, IconButton } from '@neoworks/svelte-gpui-ui';
	import Label from '#lib/components/Label.svelte';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import { x } from '#lib/layout/chromeIcons.ts';
	import { listenTogether } from '#lib/together/listenTogether.svelte.ts';
	import { colors, radii } from '#lib/theme.ts';
</script>

<View style={{ gap: 4 }}>
	{#each listenTogether.invites as invite (invite.from.userId)}
		<View
			style={{
				direction: 'row',
				align: 'center',
				gap: 12,
				padding: 8,
				radius: radii.large,
				background: colors.raised
			}}
		>
			<PeerAvatar name={invite.from.name} highlighted />
			<View style={{ flex: 1, minWidth: 0 }}>
				<Label size="body" tone="muted">
					<Label weight="semibold" tone="default">{invite.from.name}</Label> invites you to listen together
				</Label>
			</View>
			<Button
				size="sm"
				variant="primary"
				disabled={listenTogether.isPending(`join:${invite.from.userId}`)}
				onpress={() => listenTogether.join(invite.from.userId)}
			>
				Join
			</Button>
			<IconButton
				icon={x}
				label="Dismiss"
				size="sm"
				onpress={() => listenTogether.dismiss(invite.from.userId)}
			/>
		</View>
	{/each}
</View>
