<!-- What a friend is doing: their ID while offline, "Online", or the song they play with its position. -->
<script lang="ts">
	import { Icon, View } from '@neoworks/svelte-gpui/ui';
	import type { FriendPresence } from '#lib/api/together.ts';
	import Label from '#lib/components/Label.svelte';
	import { formatDuration } from '#lib/format.ts';
	import { pauseFill, playFill } from '#lib/layout/chromeIcons.ts';
	import { livePositionSeconds, listenTogether } from '#lib/together/listenTogether.svelte.ts';
	import { colors } from '#lib/theme.ts';

	let { presence, userIdText }: { presence: FriendPresence | null; userIdText: string } = $props();

	const nowPlaying = $derived(presence?.nowPlaying);
	const artists = $derived(describeArtists());
	const position = $derived(describePosition());
	const stateIcon = $derived(pickStateIcon());

	function describeArtists(): string {
		if (!nowPlaying) {
			return '';
		}
		return nowPlaying.track.artists.map((artist) => artist.name).join(', ');
	}

	function describePosition(): string {
		if (!nowPlaying) {
			return '';
		}
		const seconds = Math.floor(livePositionSeconds(nowPlaying, listenTogether.nowMs));
		const elapsed = formatDuration(seconds) || '0:00';
		const total = formatDuration(nowPlaying.track.durationSeconds);
		if (total === '') {
			return elapsed;
		}
		return `${elapsed} / ${total}`;
	}

	function pickStateIcon(): string {
		if (nowPlaying?.paused) {
			return pauseFill;
		}
		return playFill;
	}

	function describeSong(): string {
		if (!nowPlaying) {
			return '';
		}
		if (artists === '') {
			return nowPlaying.track.title;
		}
		return `${nowPlaying.track.title} · ${artists}`;
	}
</script>

{#if !presence}
	<Label size="tiny" tone="dim" truncate style={{ fontFamily: 'mono' }}>{userIdText}</Label>
{:else if nowPlaying}
	<View style={{ direction: 'row', align: 'center', gap: 6, minWidth: 0, color: colors.muted }}>
		<Icon svg={stateIcon} size={10} style={{ shrink: 0 }} />
		<View style={{ flex: 1, minWidth: 0 }}>
			<Label size="tiny" tone="muted" truncate>{describeSong()}</Label>
		</View>
		<View style={{ shrink: 0 }}>
			<Label size="tiny" tone="muted">{position}</Label>
		</View>
	</View>
{:else}
	<Label size="tiny" tone="muted">Online</Label>
{/if}
