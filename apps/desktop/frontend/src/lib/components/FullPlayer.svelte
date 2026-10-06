<script lang="ts">
	import { FloatingScrollbar } from '@neoworks-dev/ui';
	import CaretDownIcon from 'phosphor-svelte/lib/CaretDownIcon';
	import { fade, fly } from 'svelte/transition';
	import ArtBackdrop from '#lib/components/ArtBackdrop.svelte';
	import ArtistLinks from '#lib/components/ArtistLinks.svelte';
	import Cover from '#lib/components/Cover.svelte';
	import FullPlayerControls from '#lib/components/FullPlayerControls.svelte';
	import FullPlayerLyrics from '#lib/components/FullPlayerLyrics.svelte';
	import UpcomingQueue from '#lib/components/UpcomingQueue.svelte';
	import { panels } from '#lib/layout/panels.svelte.ts';
	import { lyricsView } from '#lib/lyrics/lyricsView.svelte.ts';
	import { player } from '#lib/player/player.svelte.ts';

	type Tab = 'lyrics' | 'queue';
	const tabs: { id: Tab; label: string }[] = [
		{ id: 'lyrics', label: 'Lyrics' },
		{ id: 'queue', label: 'Up next' }
	];

	let activeTab = $state<Tab>('lyrics');

	const videoId = $derived(player.track?.videoId);

	$effect(() => {
		if (videoId) {
			lyricsView.load(videoId);
		}
	});

	// The view is only mounted while open, so the transition duration is read at that moment.
	function transitionDuration(milliseconds: number): number {
		if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
			return 0;
		}
		return milliseconds;
	}

	function handleKeydown(event: KeyboardEvent): void {
		if (event.key === 'Escape') {
			panels.closeFullPlayer();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div
	role="dialog"
	aria-label="Now playing"
	transition:fly={{ y: 48, duration: transitionDuration(320) }}
	class="fixed inset-0 z-40 overflow-hidden bg-canvas text-default"
>
	<ArtBackdrop />
	<div
		aria-hidden="true"
		class="pointer-events-none absolute inset-0 bg-linear-to-b from-transparent to-canvas/60"
	></div>

	<div class="relative flex h-full flex-col gap-6 px-10 py-6">
		<header class="flex shrink-0 items-center">
			<button
				type="button"
				aria-label="Close full-screen player"
				title="Close (Esc)"
				onclick={() => panels.closeFullPlayer()}
				class="rounded-full p-2 text-muted hover:bg-hover hover:text-default"
			>
				<CaretDownIcon size={24} />
			</button>
		</header>

		{#if player.track}
			<div class="grid min-h-0 flex-1 grid-cols-[minmax(0,5fr)_minmax(0,6fr)] gap-14">
				<section class="flex min-h-0 flex-col items-center justify-center gap-8">
					<button
						type="button"
						aria-label="Close full-screen player"
						onclick={() => panels.closeFullPlayer()}
						class="min-h-0 w-full max-w-[min(32rem,50vh)] cursor-pointer rounded-2xl"
					>
						<Cover
							thumbnails={player.track.thumbnails}
							size={512}
							class="w-full rounded-2xl shadow-2xl"
						/>
					</button>
					<div class="w-full max-w-[min(32rem,50vh)] min-w-0">
						<h2 class="truncate text-3xl font-bold">{player.track.title}</h2>
						<p class="truncate text-lg text-muted">
							<ArtistLinks artists={player.track.artists} />
						</p>
					</div>
					<div class="w-full max-w-[min(32rem,50vh)]">
						<FullPlayerControls />
					</div>
				</section>

				<section class="flex min-h-0 flex-col gap-4">
					<div role="tablist" class="flex shrink-0 gap-2">
						{#each tabs as tab (tab.id)}
							<button
								type="button"
								role="tab"
								aria-selected={activeTab === tab.id}
								onclick={() => (activeTab = tab.id)}
								class={[
									'duration-fast rounded-full px-4 py-1.5 text-sm font-semibold transition-colors',
									activeTab === tab.id && 'bg-action text-action-fg',
									activeTab !== tab.id && 'text-muted hover:bg-hover hover:text-default'
								]}
							>
								{tab.label}
							</button>
						{/each}
					</div>

					<FloatingScrollbar class="min-h-0 flex-1">
						{#if activeTab === 'lyrics'}
							<FullPlayerLyrics />
						{:else}
							<div class="flex flex-col gap-4 pr-4">
								<UpcomingQueue size="comfortable" />
							</div>
						{/if}
					</FloatingScrollbar>
				</section>
			</div>
		{:else}
			<p class="flex flex-1 items-center justify-center text-xl text-dim" in:fade>
				Nothing is playing.
			</p>
		{/if}
	</div>
</div>
