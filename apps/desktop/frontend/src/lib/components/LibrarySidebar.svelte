<script lang="ts">
	import BooksIcon from 'phosphor-svelte/lib/BooksIcon';
</script>

<aside class="flex w-72 shrink-0 flex-col gap-2 rounded-lg bg-surface p-2">
	<header class="flex h-12 items-center gap-3 px-2 text-dim">
		<BooksIcon size={22} />
		<h2 class="text-base font-semibold">Your Library</h2>
	</header>

	<div class="flex flex-col gap-1 rounded-lg bg-raised p-4">
		<p class="text-base font-semibold">Your library is empty</p>
		<p class="text-sm text-muted">Songs, albums and playlists you save will show up here.</p>
	</div>
</aside>
