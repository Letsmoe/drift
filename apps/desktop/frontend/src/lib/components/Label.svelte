<!--
  Text in Drift's type scale and tones. Nested in another Label it is a run inside the same paragraph.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Text, type Style } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { colors, textSizeStyle, type TextSize } from '#lib/theme.ts';

	type Tone = 'default' | 'muted' | 'dim' | 'faint' | 'accent' | 'danger' | 'inherit';
	type Weight = 'normal' | 'medium' | 'semibold' | 'bold';

	let {
		size = 'body',
		tone = 'inherit',
		accented = false,
		weight,
		truncate = false,
		lines,
		style,
		children
	}: {
		size?: TextSize;
		/** `inherit` keeps the colour of the element around it. */
		tone?: Tone;
		/** Draws the text in the accent colour whatever `tone` says, as the playing track's title is. */
		accented?: boolean;
		weight?: Weight;
		/** One line, cut off with an ellipsis. */
		truncate?: boolean;
		/** At most this many lines. */
		lines?: number;
		style?: Style;
		children?: Snippet;
	} = $props();

	const look = $derived(buildLook());

	function toneColor(): string | undefined {
		if (accented) {
			return appearance.accent;
		}
		switch (tone) {
			case 'default':
				return colors.text;
			case 'muted':
				return appearance.muted;
			case 'dim':
				return appearance.dim;
			case 'faint':
				return appearance.faint;
			case 'accent':
				return appearance.accent;
			case 'danger':
				return colors.red;
			default:
				return undefined;
		}
	}

	function buildLook(): Style {
		const built: Style = { ...textSizeStyle(size) };
		const color = toneColor();
		if (color !== undefined) {
			built.color = color;
		}
		if (weight !== undefined) {
			built.fontWeight = weight;
		}
		if (truncate) {
			built.truncate = true;
		}
		if (lines !== undefined) {
			built.maxLines = lines;
		}
		return { ...built, ...style };
	}
</script>

<Text style={look}>{@render children?.()}</Text>
