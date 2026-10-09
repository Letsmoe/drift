<script lang="ts">
	import UsersThreeIcon from 'phosphor-svelte/lib/UsersThreeIcon';
	import TogetherAddFriend from '#lib/components/TogetherAddFriend.svelte';
	import TogetherFriends from '#lib/components/TogetherFriends.svelte';
	import TogetherInvites from '#lib/components/TogetherInvites.svelte';
	import TogetherProfile from '#lib/components/TogetherProfile.svelte';
	import TogetherRequests from '#lib/components/TogetherRequests.svelte';
	import TogetherSession from '#lib/components/TogetherSession.svelte';
	import { listenTogether } from '#lib/together/listenTogether.svelte.ts';

	listenTogether.start();

	let open = $state(false);
	let container = $state<HTMLElement>();

	const attentionCount = $derived(listenTogether.attentionCount);
	const listenerCount = $derived(listenTogether.otherPeers.length);
	const buttonLabel = $derived(describeButton());

	function describeButton(): string {
		if (listenTogether.requests.length > 0) {
			return `Listen together: ${listenTogether.requests.length} waiting to join`;
		}
		if (listenTogether.invites.length > 0) {
			return `Listen together: ${listenTogether.invites.length} invite(s)`;
		}
		if (listenTogether.connected) {
			return `Listen together: listening with ${listenerCount}`;
		}
		return 'Listen together';
	}

	function toggle(): void {
		open = !open;
		if (!open) {
			listenTogether.clearError();
		}
	}

	function close(): void {
		open = false;
		listenTogether.clearError();
	}

	function closeOnOutsidePointer(event: PointerEvent): void {
		if (open && container && !container.contains(event.target as Node)) {
			close();
		}
	}

	function closeOnEscape(event: KeyboardEvent): void {
		if (open && event.key === 'Escape') {
			close();
		}
	}
</script>

<svelte:window onpointerdown={closeOnOutsidePointer} onkeydown={closeOnEscape} />

{#snippet sectionTitle(title: string)}
	<h3 class="text-xs font-semibold tracking-wide text-dim uppercase">{title}</h3>
{/snippet}

<div bind:this={container} class="relative">
	<button
		type="button"
		aria-label={buttonLabel}
		title={buttonLabel}
		aria-haspopup="dialog"
		aria-expanded={open}
		onclick={toggle}
		class={[
			'duration-fast glass relative flex h-9 items-center gap-1.5 rounded-full px-3 text-sm transition-colors hover:bg-hover',
			listenTogether.roomOpen && 'text-default',
			!listenTogether.roomOpen && 'text-dim hover:text-default'
		]}
	>
		<UsersThreeIcon size={18} weight={listenTogether.connected ? 'fill' : 'regular'} />
		{#if listenTogether.connected}
			<span class="font-semibold tabular-nums">{listenerCount + 1}</span>
		{/if}
		{#if attentionCount > 0}
			<span class="absolute top-1 right-1 size-2 rounded-full bg-action"></span>
		{/if}
	</button>

	{#if open}
		<div
			role="dialog"
			aria-label="Listen together"
			class="absolute top-full right-0 z-40 mt-2 flex max-h-[min(36rem,calc(100vh-8rem))] w-96 flex-col gap-4 overflow-y-auto rounded-xl border border-line bg-elevated p-4 shadow-lg"
		>
			<header class="flex flex-col gap-3">
				<h2 class="text-base font-bold">Listen together</h2>
				<TogetherProfile />
			</header>

			{#if listenTogether.requests.length > 0}
				<TogetherRequests />
			{/if}
			{#if listenTogether.invites.length > 0}
				<TogetherInvites />
			{/if}

			<div class="h-px bg-line"></div>
			<TogetherSession />

			<div class="h-px bg-line"></div>
			<section class="flex flex-col gap-3">
				{@render sectionTitle('Friends')}
				<TogetherFriends />
				<TogetherAddFriend />
			</section>

			{#if listenTogether.errorMessage !== ''}
				<p class="text-sm text-red">{listenTogether.errorMessage}</p>
			{/if}
		</div>
	{/if}
</div>
