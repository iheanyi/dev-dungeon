<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { userStore } from '$lib/stores/user';

	let { children } = $props();

	onMount(() => {
		userStore.checkAuth();
	});

	async function handleLogout() {
		await userStore.logout();
	}
</script>

<svelte:head>
	<title>/dev/dungeon - A Unix-themed Terminal Roguelike</title>
	<meta name="description" content="Navigate procedurally generated filesystem dungeons from /home to /dev/null. Battle rogue processes like zombies, daemons, and fork bombs. Free to play over SSH!" />
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
</svelte:head>

<div class="crt min-h-screen">
	<!-- Header/Nav -->
	<header class="border-b border-terminal-green/30 bg-terminal-bg/90 backdrop-blur">
		<div class="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
			<a href="/" class="font-mono text-xl text-terminal-green hover:text-terminal-cyan transition-colors">
				/dev/dungeon
			</a>

			<nav class="flex items-center gap-6">
				<a href="/leaderboard" class="font-mono text-sm text-terminal-gray-light hover:text-terminal-green transition-colors">
					Leaderboard
				</a>
				<a href="/daily" class="font-mono text-sm text-terminal-gray-light hover:text-terminal-green transition-colors">
					Daily
				</a>
				<a href="/about" class="font-mono text-sm text-terminal-gray-light hover:text-terminal-green transition-colors">
					About
				</a>

				{#if $userStore.loading}
					<span class="font-mono text-sm text-terminal-gray">...</span>
				{:else if $userStore.username}
					<div class="flex items-center gap-3">
						<a href="/players/{$userStore.username}" class="font-mono text-sm text-terminal-cyan hover:text-terminal-green transition-colors">
							{$userStore.username}
						</a>
						<button
							onclick={handleLogout}
							class="font-mono text-xs text-terminal-gray hover:text-terminal-red transition-colors"
						>
							logout
						</button>
					</div>
				{:else}
					<span class="font-mono text-xs text-terminal-gray">
						Press Ctrl+L in game to login
					</span>
				{/if}
			</nav>
		</div>
	</header>

	<div class="max-w-6xl mx-auto px-4 py-8">
		{@render children()}
	</div>
</div>
