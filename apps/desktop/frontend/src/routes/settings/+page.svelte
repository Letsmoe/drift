<script lang="ts">
	import { Checkbox, Slider } from '@neoworks-dev/ui';
	import { crossfadeSecondsLimits, playbackSettings } from '#lib/player/playbackSettings.svelte.ts';
</script>

<div class="flex flex-col gap-8 p-6">
	<h1 class="text-3xl font-bold">Settings</h1>

	<section class="flex max-w-xl flex-col gap-4">
		<h2 class="text-lg font-semibold">Playback</h2>
		<label class="flex cursor-pointer items-start gap-3 text-sm text-dim">
			<span class="pt-0.5">
				<Checkbox
					bind:checked={
						() => playbackSettings.crossfade, (enabled) => playbackSettings.setCrossfade(enabled)
					}
				/>
			</span>
			<span>
				<span class="block font-medium text-default">Smooth transitions</span>
				The next song fades in while the current one fades out.
			</span>
		</label>
		{#if playbackSettings.crossfade}
			<div class="flex items-center gap-4 pl-8 text-sm">
				<span class="shrink-0 text-dim">Crossfade</span>
				<Slider
					label="Crossfade length in seconds"
					value={playbackSettings.crossfadeSeconds}
					min={crossfadeSecondsLimits.min}
					max={crossfadeSecondsLimits.max}
					oninput={(seconds) => playbackSettings.setCrossfadeSeconds(seconds)}
				/>
				<span class="w-8 shrink-0 text-right text-muted tabular-nums"
					>{playbackSettings.crossfadeSeconds} s</span
				>
			</div>
		{/if}
	</section>
</div>
