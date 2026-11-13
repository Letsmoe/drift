<!-- The Listen Together button of the top bar and the panel it opens. -->
<script lang="ts">
	import {
		Button,
		Icon,
		Popover,
		Scroll,
		Tooltip,
		View,
		type Style
	} from '@neoworks/svelte-gpui/ui';
	import { window } from '@neoworks/svelte-gpui';
	import { Divider } from '@neoworks/svelte-gpui-ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import Label from '#lib/components/Label.svelte';
	import TogetherAddFriend from '#lib/components/TogetherAddFriend.svelte';
	import TogetherFriends from '#lib/components/TogetherFriends.svelte';
	import TogetherInvites from '#lib/components/TogetherInvites.svelte';
	import TogetherProfile from '#lib/components/TogetherProfile.svelte';
	import TogetherRequests from '#lib/components/TogetherRequests.svelte';
	import TogetherSession from '#lib/components/TogetherSession.svelte';
	import { usersThree, usersThreeFill } from '#lib/layout/chromeIcons.ts';
	import { listenTogether } from '#lib/together/listenTogether.svelte.ts';
	import { colors, floatingPanel, glassBackground, hoverTransition } from '#lib/theme.ts';

	listenTogether.start();

	const panelWidth = 384;
	const panelPadding = 16;
	// The panel stays inside the window: 36rem at most, else the window less the bar and player.
	const panelMaxHeight = $derived(Math.min(576, window.size.height - 128));

	let open = $state(false);

	const attentionCount = $derived(listenTogether.attentionCount);
	const listenerCount = $derived(listenTogether.otherPeers.length);
	const buttonLabel = $derived(describeButton());
	const buttonIcon = $derived(pickIcon());

	function describeButton(): string {
		if (listenTogether.requests.length > 0) {
			return `Listen together: ${listenTogether.requests.length} waiting to join`;
		}
		if (listenTogether.invites.length > 0) {
			return `Listen together: ${listenTogether.invites.length} invite(s)`;
		}
		if (listenTogether.connected) {
			return `Listen together: listening with ${listenerCount}`;
		}
		return 'Listen together';
	}

	function pickIcon(): string {
		if (listenTogether.connected) {
			return usersThreeFill;
		}
		return usersThree;
	}

	function buttonColor(): string {
		if (listenTogether.roomOpen) {
			return colors.text;
		}
		return appearance.dim;
	}

	function panelStyle(): Style {
		return {
			...floatingPanel,
			width: panelWidth,
			maxHeight: panelMaxHeight,
			padding: 0
		};
	}

	function handleClose(): void {
		listenTogether.clearError();
	}
</script>

{#snippet sectionTitle(title: string)}
	<Label size="tiny" weight="semibold" tone="dim" style={{ uppercase: true }}>{title}</Label>
{/snippet}

<Popover bind:open side="bottom" align="end" offset={8} style={panelStyle()} onclose={handleClose}>
	{#snippet trigger({ toggle })}
		<Tooltip text={buttonLabel} side="bottom">
			<Button
				style={{
					direction: 'row',
					align: 'center',
					justify: 'center',
					gap: 6,
					height: 36,
					minWidth: 36,
					paddingX: 9,
					radius: 9999,
					background: glassBackground,
					color: buttonColor(),
					transition: hoverTransition
				}}
				hover={{ background: colors.hover, color: colors.text }}
				onpress={toggle}
			>
				<Icon svg={buttonIcon} size={18} />
				{#if listenTogether.connected}
					<Label size="body" weight="semibold">{listenerCount + 1}</Label>
				{/if}
				{#if attentionCount > 0}
					<View
						style={{
							position: 'absolute',
							top: 4,
							right: 4,
							size: 8,
							radius: 9999,
							background: appearance.accent
						}}
					/>
				{/if}
			</Button>
		</Tooltip>
	{/snippet}
	{#snippet children()}
		<Scroll style={{ maxHeight: panelMaxHeight }}>
			<View style={{ gap: 16, padding: panelPadding }}>
				<View style={{ gap: 12 }}>
					<Label size="medium" weight="bold" tone="default">Listen together</Label>
					<TogetherProfile />
				</View>

				{#if listenTogether.requests.length > 0}
					<TogetherRequests />
				{/if}
				{#if listenTogether.invites.length > 0}
					<TogetherInvites />
				{/if}

				<Divider />
				<TogetherSession />

				<Divider />
				<View style={{ gap: 12 }}>
					{@render sectionTitle('Friends')}
					<TogetherFriends />
					<TogetherAddFriend />
				</View>

				{#if listenTogether.errorMessage !== ''}
					<Label size="body" tone="danger">{listenTogether.errorMessage}</Label>
				{/if}
			</View>
		</Scroll>
	{/snippet}
</Popover>
