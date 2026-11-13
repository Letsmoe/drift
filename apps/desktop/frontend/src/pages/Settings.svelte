<script lang="ts">
	import { Scroll, View } from '@neoworks/svelte-gpui/ui';
	import { Checkbox, Slider } from '@neoworks/svelte-gpui-ui';
	import Label from '#lib/components/Label.svelte';
	import { crossfadeSecondsLimits, playbackSettings } from '#lib/player/playbackSettings.svelte.ts';
</script>

<Scroll style={{ flex: 1, minHeight: 0 }}>
	<View style={{ gap: 32, padding: 24 }}>
		<Label size="title" weight="bold">Settings</Label>

		<View style={{ maxWidth: 576, gap: 16 }}>
			<Label size="large" weight="semibold">Playback</Label>
			<Checkbox
				label="Smooth transitions"
				supportingText="The next song fades in while the current one fades out."
				checked={playbackSettings.crossfade}
				onchange={(enabled) => playbackSettings.setCrossfade(enabled)}
			/>
			{#if playbackSettings.crossfade}
				<View style={{ direction: 'row', align: 'center', gap: 16, paddingLeft: 32 }}>
					<Label size="small" tone="dim" style={{ shrink: 0 }}>Crossfade</Label>
					<View style={{ flex: 1, minWidth: 0 }}>
						<Slider
							value={playbackSettings.crossfadeSeconds}
							min={crossfadeSecondsLimits.min}
							max={crossfadeSecondsLimits.max}
							step={1}
							oninput={(seconds) => playbackSettings.setCrossfadeSeconds(seconds)}
						/>
					</View>
					<Label size="small" tone="muted" style={{ width: 32, textAlign: 'right', shrink: 0 }}>
						{playbackSettings.crossfadeSeconds} s
					</Label>
				</View>
			{/if}
		</View>
	</View>
</Scroll>
