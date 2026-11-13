<!-- The user's own name and ID in the Listen Together panel, and whether friends see what they play. -->
<script lang="ts">
	import { clipboard } from '@neoworks/svelte-gpui';
	import { Button, Icon, View } from '@neoworks/svelte-gpui/ui';
	import { Checkbox } from '@neoworks/svelte-gpui-ui';
	import Label from '#lib/components/Label.svelte';
	import PeerAvatar from '#lib/components/PeerAvatar.svelte';
	import { check, copy, pencilSimple } from '#lib/layout/chromeIcons.ts';
	import TextInput from '#lib/layout/TextInput.svelte';
	import { formatUserId, listenTogether } from '#lib/together/listenTogether.svelte.ts';
	import { colors, hoverTransition } from '#lib/theme.ts';

	const copiedResetMs = 1500;

	const self = $derived(listenTogether.state?.self);
	let editing = $state(false);
	let draftName = $state('');
	let copied = $state(false);

	function startEditing(): void {
		if (!self) {
			return;
		}
		draftName = self.name;
		editing = true;
	}

	async function saveName(): Promise<void> {
		if (await listenTogether.rename(draftName)) {
			editing = false;
		}
	}

	async function copyId(): Promise<void> {
		if (!self) {
			return;
		}
		await clipboard.write(formatUserId(self.userId));
		copied = true;
		setTimeout(() => (copied = false), copiedResetMs);
	}

	function copyIcon(): string {
		if (copied) {
			return check;
		}
		return copy;
	}
</script>

{#if self}
	<View style={{ gap: 12 }}>
		<View style={{ direction: 'row', align: 'center', gap: 12 }}>
			<PeerAvatar name={self.name} highlighted />
			<View style={{ flex: 1, minWidth: 0, gap: 2 }}>
				{#if editing}
					<TextInput
						bind:value={draftName}
						maxLength={40}
						height={28}
						autofocus
						onsubmit={saveName}
						onescape={() => (editing = false)}
						onblur={() => (editing = false)}
					/>
				{:else}
					<Button
						style={{
							direction: 'row',
							align: 'center',
							gap: 6,
							alignSelf: 'start',
							maxWidth: 'full'
						}}
						hover={{ color: colors.text }}
						onpress={startEditing}
					>
						<View style={{ minWidth: 0, shrink: 1 }}>
							<Label size="body" weight="semibold" tone="default" truncate>{self.name}</Label>
						</View>
						<Icon svg={pencilSimple} size={12} color={colors.dim} style={{ shrink: 0 }} />
					</Button>
				{/if}
				<Button
					style={{
						direction: 'row',
						align: 'center',
						gap: 6,
						alignSelf: 'start',
						color: colors.muted,
						transition: hoverTransition
					}}
					hover={{ color: colors.text }}
					onpress={copyId}
				>
					<Label size="tiny" style={{ fontFamily: 'mono' }}>{formatUserId(self.userId)}</Label>
					<Icon svg={copyIcon()} size={12} />
				</Button>
			</View>
		</View>
		<Checkbox
			size="sm"
			label="Show friends what I'm playing"
			checked={listenTogether.sharePresence}
			disabled={listenTogether.isPending('share-presence')}
			onchange={(share) => listenTogether.setSharePresence(share)}
		/>
	</View>
{/if}
