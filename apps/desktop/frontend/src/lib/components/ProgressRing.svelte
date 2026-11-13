<script lang="ts">
	import { Canvas } from '@neoworks/svelte-gpui/ui';
	import { appearance } from '#lib/appearance.svelte.ts';
	import { withAlpha } from '#lib/theme.ts';

	let {
		fraction,
		size = 18,
		strokeWidth = 2.5,
		color
	}: {
		/** Completed share between 0 and 1; anything else shows an indeterminate, spinning ring. */
		fraction: number;
		size?: number;
		strokeWidth?: number;
		/** An #rrggbb colour; the secondary text colour when left out. */
		color?: string;
	} = $props();

	/** Share of the ring an indeterminate arc covers. */
	const indeterminateShare = 0.25;
	const spinStepMs = 50;
	const spinStepRadians = 0.45;
	const segmentsPerFullCircle = 48;
	const trackColor = withAlpha('#ffffff', 0.25);

	let canvas = $state<Canvas>();
	let spin = $state(0);
	const known = $derived(fraction > 0 && fraction <= 1);

	// The indeterminate ring turns until a share is known.
	$effect(() => {
		if (known) {
			return;
		}
		const timer = setInterval(() => (spin += spinStepRadians), spinStepMs);
		return () => clearInterval(timer);
	});

	$effect(() => {
		if (!canvas) {
			return;
		}
		const share = known ? fraction : indeterminateShare;
		const start = known ? -Math.PI / 2 : spin;
		paint(canvas, share, start);
	});

	function ringColor(): string {
		if (color !== undefined) {
			return color;
		}
		return appearance.muted;
	}

	function paint(target: Canvas, share: number, startAngle: number): void {
		const middle = size / 2;
		const radius = (size - strokeWidth) / 2;
		target.clear();
		drawArc(target, middle, radius, -Math.PI / 2, 1, trackColor);
		drawArc(target, middle, radius, startAngle, share, ringColor());
	}

	function drawArc(
		target: Canvas,
		middle: number,
		radius: number,
		startAngle: number,
		share: number,
		strokeColor: string
	): void {
		const segments = Math.max(2, Math.ceil(segmentsPerFullCircle * share));
		for (let step = 0; step <= segments; step += 1) {
			const angle = startAngle + (2 * Math.PI * share * step) / segments;
			const x = middle + radius * Math.cos(angle);
			const y = middle + radius * Math.sin(angle);
			if (step === 0) {
				target.begin(x, y, { color: strokeColor, width: strokeWidth });
				continue;
			}
			target.lineTo(x, y);
		}
		target.end();
	}
</script>

<Canvas bind:this={canvas} vector style={{ size }} />
