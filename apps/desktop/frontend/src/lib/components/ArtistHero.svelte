<script lang="ts">
	import type { Thumbnail } from '#lib/api/music.ts';
	import { pickThumbnail } from '#lib/format.ts';

	let { name, thumbnails }: { name: string; thumbnails: Thumbnail[] } = $props();

	const imageUrl = $derived(pickThumbnail(thumbnails, 1200));
	const backgroundImage = $derived(heroBackground(imageUrl));

	function heroBackground(url: string | undefined): string {
		if (!url) {
			return 'none';
		}
		return `url("${url}")`;
	}
</script>

<!-- Cancels the layout padding so the banner fills the top of the pane. -->
<header
	class="relative -mx-3 -mt-4 flex h-80 items-end bg-raised bg-cover bg-[position:center_20%] px-9 pb-6"
	style:background-image={backgroundImage}
>
	<div class="absolute inset-0 bg-linear-to-t from-black/80 via-black/20 to-black/10"></div>
	<h1 class="relative line-clamp-2 text-6xl leading-tight font-bold text-white xl:text-8xl">
		{name}
	</h1>
</header>
