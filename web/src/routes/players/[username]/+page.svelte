<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { getPlayerProfile, formatDate, getFloorName, type PlayerProfile } from '$lib/api';

	let profile: PlayerProfile | null = $state(null);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		const username = $page.params.username;
		if (!username) {
			error = 'Invalid player';
			loading = false;
			return;
		}
		const response = await getPlayerProfile(username);

		if (response.success && response.data) {
			profile = response.data;
		} else {
			error = response.error || 'Player not found';
		}
		loading = false;
	});
</script>

<svelte:head>
	<title>{profile?.username || 'Player'} - /dev/dungeon</title>
</svelte:head>

<div class="space-y-8">
	{#if loading}
		<div class="text-center py-12">
			<p class="text-terminal-green animate-pulse">Loading player profile...</p>
		</div>
	{:else if error}
		<div class="terminal-box p-8 text-center">
			<p class="text-terminal-red">{error}</p>
			<a href="/leaderboard" class="text-terminal-green hover:underline mt-4 inline-block">
				Back to leaderboard
			</a>
		</div>
	{:else if profile}
		<div class="terminal-box p-6">
			<h1 class="text-2xl text-terminal-green glow mb-2">{profile.username}</h1>
			<p class="text-terminal-gray text-sm">
				Player since {formatDate(profile.created_at)}
			</p>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
			<div class="terminal-box p-4 text-center">
				<p class="text-3xl text-terminal-green glow">{profile.runs_completed}</p>
				<p class="text-terminal-gray text-sm mt-1">Runs Completed</p>
			</div>
			<div class="terminal-box p-4 text-center">
				<p class="text-3xl text-terminal-cyan glow">{getFloorName(profile.deepest_floor)}</p>
				<p class="text-terminal-gray text-sm mt-1">Deepest Floor</p>
			</div>
			<div class="terminal-box p-4 text-center">
				<p class="text-3xl text-terminal-red">{profile.total_deaths}</p>
				<p class="text-terminal-gray text-sm mt-1">Total Deaths</p>
			</div>
		</div>

		{#if profile.unlocked_classes && profile.unlocked_classes.length > 0}
			<div class="terminal-box p-6">
				<h2 class="text-lg text-terminal-green mb-4">Unlocked Classes</h2>
				<div class="flex flex-wrap gap-2">
					{#each profile.unlocked_classes as className}
						<span class="px-3 py-1 border border-terminal-green/50 text-terminal-green text-sm">
							{className}
						</span>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>
