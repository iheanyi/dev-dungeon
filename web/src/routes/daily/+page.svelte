<svelte:head>
	<title>Daily Challenge - /dev/dungeon</title>
	<meta name="description" content="Today's /dev/dungeon daily challenge. Same seed for everyone, compete globally. Resets at midnight UTC. Can you top the leaderboard?" />
</svelte:head>

<script lang="ts">
	import { onMount } from 'svelte';
	import { getDailySeed, getLeaderboard, formatTime, type LeaderboardEntry, type DailySeed } from '$lib/api';

	let daily = $state<DailySeed | null>(null);
	let entries = $state<LeaderboardEntry[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		loading = true;

		const [dailyResult, leaderboardResult] = await Promise.all([
			getDailySeed(),
			getLeaderboard('daily'),
		]);

		if (dailyResult.success && dailyResult.data) {
			daily = dailyResult.data;
		}

		if (leaderboardResult.success && leaderboardResult.data) {
			entries = leaderboardResult.data.slice(0, 10); // Top 10 for daily
		}

		if (!dailyResult.success) {
			error = dailyResult.error || 'Failed to load daily challenge';
		}

		loading = false;
	});

	function formatSeed(seed: number): string {
		return seed.toString(16).toUpperCase().slice(0, 8);
	}
</script>

<div class="max-w-3xl mx-auto">
	<a href="/" class="text-terminal-gray hover:text-terminal-green mb-6 inline-block">
		&lt; back
	</a>

	<div class="terminal-box p-6 mb-6">
		<h1 class="text-terminal-amber text-2xl mb-2">$ cron.daily</h1>
		<p class="text-terminal-green-dim mb-6">
			Same dungeon, same seed. Compete globally. Resets at midnight UTC.
		</p>

		{#if loading}
			<div class="text-center py-8">
				<p class="text-terminal-green animate-pulse">Loading daily challenge...</p>
			</div>
		{:else if error}
			<div class="bg-terminal-bg border border-terminal-red p-4">
				<p class="text-terminal-red">{error}</p>
			</div>
		{:else if daily}
			<div class="grid sm:grid-cols-2 gap-6">
				<div class="bg-terminal-bg p-4 border border-terminal-gray">
					<p class="text-terminal-gray text-sm mb-1">DATE</p>
					<p class="text-terminal-green text-2xl font-bold">{daily.date}</p>
				</div>
				<div class="bg-terminal-bg p-4 border border-terminal-gray">
					<p class="text-terminal-gray text-sm mb-1">SEED</p>
					<p class="text-terminal-amber text-2xl font-bold font-mono">
						0x{formatSeed(daily.seed)}
					</p>
				</div>
			</div>

			<div class="mt-6 bg-terminal-bg p-4">
				<p class="text-terminal-gray mb-2"># Join today's run:</p>
				<code class="text-terminal-green">
					ssh -p 2222 player@dev-dungeon.com --daily
				</code>
			</div>
		{/if}
	</div>

	<!-- Today's Top 10 -->
	{#if entries.length > 0}
		<div class="terminal-box p-6">
			<h2 class="text-terminal-amber mb-4">$ head -10 /var/log/daily.log</h2>

			<div class="space-y-2">
				{#each entries as entry, i}
					<div class="flex items-center gap-4 py-2 border-b border-terminal-gray last:border-0">
						<span class="w-8 text-center {i < 3 ? 'text-terminal-amber' : 'text-terminal-gray'}">
							{i + 1}.
						</span>
						<span class="flex-1 text-terminal-green">{entry.username}</span>
						<span class="text-terminal-cyan">{entry.class}</span>
						<span class="text-terminal-green-bright font-bold">
							{entry.score.toLocaleString()}
						</span>
						<span class="text-terminal-gray text-sm hidden sm:block">
							{formatTime(entry.time_seconds)}
						</span>
					</div>
				{/each}
			</div>

			<a href="/leaderboard?type=daily" class="text-terminal-green-dim hover:text-terminal-green block mt-4 text-sm">
				View full daily leaderboard →
			</a>
		</div>
	{/if}

	<!-- Rules -->
	<div class="terminal-box p-6 mt-6">
		<h2 class="text-terminal-amber mb-4">$ man daily</h2>
		<ul class="space-y-2 text-terminal-green-dim text-sm">
			<li>• One attempt per day per player</li>
			<li>• Same seed = same dungeon layout & enemies</li>
			<li>• Resets at 00:00 UTC</li>
			<li>• Death ends your run — no retries until tomorrow</li>
			<li>• Score based on depth, kills, items, and time</li>
		</ul>
	</div>
</div>
