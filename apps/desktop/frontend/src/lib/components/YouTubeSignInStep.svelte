<script lang="ts">
	import { Button, LoadingSpinner } from '@neoworks-dev/ui';
	import type { MusicImportFlow } from '#lib/import/musicImport.svelte.ts';

	let { flow }: { flow: MusicImportFlow } = $props();

	const youtube = $derived(flow.youtube);
</script>

<div class="flex flex-col gap-4">
	<section class="flex flex-col gap-2">
		<h3 class="text-base font-semibold">Use a browser where you are signed in</h3>
		<p class="text-sm text-muted">
			If you are already signed in to music.youtube.com in a browser on this computer, Drift can
			pick up that session. This reads the browser's cookie store, and only when you press the
			button.
		</p>
		<div class="self-start">
			<Button variant="primary" disabled={youtube.busy} onclick={() => youtube.searchBrowsers()}
				>Look for signed-in browsers</Button
			>
		</div>

		{#if youtube.busy}
			<div class="flex justify-center py-3"><LoadingSpinner /></div>
		{/if}

		{#if youtube.searched}
			{#if youtube.usableSessions.length === 0}
				<p class="text-sm text-muted">No browser with a YouTube Music session was found.</p>
			{/if}
			<ul class="flex flex-col">
				{#each youtube.usableSessions as session (session.browser + session.profile)}
					<li class="flex items-center gap-3 rounded-md px-2 py-2 hover:bg-hover">
						<span class="min-w-0 flex-1 truncate text-sm">
							{session.browser}
							<span class="text-dim">{session.profile}</span>
						</span>
						<Button
							size="sm"
							variant="primary"
							disabled={youtube.busy}
							onclick={() => flow.useBrowserSession(session)}>Use this session</Button
						>
					</li>
				{/each}
			</ul>
			{#if youtube.unusableSessions.length > 0}
				<details class="text-sm">
					<summary class="cursor-pointer text-muted">
						{youtube.unusableSessions.length} other profiles without a usable session
					</summary>
					<ul class="mt-2 flex flex-col gap-1">
						{#each youtube.unusableSessions as session (session.browser + session.profile)}
							<li class="text-xs text-dim">
								{session.browser}
								{session.profile}: {session.reason}
							</li>
						{/each}
					</ul>
				</details>
			{/if}
		{/if}
	</section>

	<details class="rounded-md bg-raised px-3 py-2">
		<summary class="cursor-pointer text-sm font-semibold">Paste cookies instead</summary>
		<div class="mt-2 flex flex-col gap-2">
			<ol class="list-decimal pl-5 text-sm text-muted">
				<li>Open music.youtube.com and sign in.</li>
				<li>Open the browser's DevTools and go to the Network tab.</li>
				<li>Reload, click any request to music.youtube.com and copy its Cookie request header.</li>
				<li>Paste it here. A whole request or a "Copy as cURL" command works too.</li>
			</ol>
			<textarea
				bind:value={youtube.pastedCookies}
				aria-label="Cookie header"
				rows="4"
				placeholder="SID=...; SAPISID=...; ..."
				class="min-w-0 rounded-md bg-canvas px-3 py-2 font-mono text-xs outline-none placeholder:text-dim focus:border-line-strong"
			></textarea>
			<div class="self-start">
				<Button
					disabled={youtube.busy || youtube.pastedCookies.trim() === ''}
					onclick={() => flow.usePastedCookies()}>Sign in with these cookies</Button
				>
			</div>
		</div>
	</details>

	<p class="text-xs text-dim">
		Your session stays on this computer and is only used to read your library. Sign out any time.
	</p>

	{#if youtube.errorMessage !== ''}
		<p class="text-sm text-red">{youtube.errorMessage}</p>
	{/if}

	<div class="flex justify-end">
		<Button variant="ghost" onclick={() => flow.closeSignIn()}>Back</Button>
	</div>
</div>
