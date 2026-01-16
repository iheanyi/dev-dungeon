<svelte:head>
	<title>Leaderboard - /dev/dungeon High Scores</title>
	<meta name="description" content="See who has descended the deepest in /dev/dungeon. Global leaderboard showing top players, their classes, floors cleared, and completion times." />
</svelte:head>

<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getLeaderboard, formatTime, formatDate, getFloorName, type LeaderboardEntry } from '$lib/api';

	let entries = $state<LeaderboardEntry[]>([]);
	let loading = $state(true);
	let error = $state('');
	let runType = $state('');

	async function loadLeaderboard() {
		loading = true;
		error = '';

		const result = await getLeaderboard(runType || undefined);

		if (result.success && result.data) {
			entries = result.data;
		} else {
			error = result.error || 'Failed to load leaderboard';
		}

		loading = false;
	}

	onMount(() => {
		// Read type from URL query param
		const typeParam = $page.url.searchParams.get('type');
		if (typeParam && ['standard', 'daily', 'seeded'].includes(typeParam)) {
			runType = typeParam;
		}
		loadLeaderboard();
	});

	function handleTypeChange() {
		// Update URL with query param
		const url = new URL(window.location.href);
		if (runType) {
			url.searchParams.set('type', runType);
		} else {
			url.searchParams.delete('type');
		}
		goto(url.pathname + url.search, { replaceState: true, noScroll: true });
		loadLeaderboard();
	}
</script>

<div class="max-w-4xl mx-auto">
	<a href="/" class="text-terminal-gray hover:text-terminal-green mb-6 inline-block">
		&lt; back
	</a>

	<div class="terminal-box p-6">
		<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
			<h1 class="text-terminal-amber text-2xl">$ top -o score</h1>

			<select
				bind:value={runType}
				onchange={handleTypeChange}
				class="input-terminal w-auto"
			>
				<option value="">All Runs</option>
				<option value="standard">Standard</option>
				<option value="daily">Daily</option>
				<option value="seeded">Seeded</option>
			</select>
		</div>

		{#if loading}
			<div class="text-center py-8">
				<p class="text-terminal-green animate-pulse">Loading...</p>
			</div>
		{:else if error}
			<div class="bg-terminal-bg border border-terminal-red p-4">
				<p class="text-terminal-red">{error}</p>
			</div>
		{:else if entries.length === 0}
			<div class="text-center py-8">
				<p class="text-terminal-gray">No entries yet. Be the first to descend!</p>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="table-terminal w-full text-sm">
					<thead>
						<tr>
							<th class="w-12">#</th>
							<th>Player</th>
							<th>Class</th>
							<th class="text-right">Score</th>
							<th class="text-right">Depth</th>
							<th class="text-right hidden sm:table-cell">Time</th>
							<th class="hidden md:table-cell">Date</th>
						</tr>
					</thead>
					<tbody>
						{#each entries as entry, i}
							<tr class={i < 3 ? 'text-terminal-amber' : ''}>
								<td class="text-center">
									{#if i === 0}
										<span class="text-terminal-amber">1st</span>
									{:else if i === 1}
										<span class="text-terminal-gray-light">2nd</span>
									{:else if i === 2}
										<span class="text-terminal-gray-light">3rd</span>
									{:else}
										{i + 1}
									{/if}
								</td>
								<td>
									<a
										href="/players/{entry.username}"
										class="hover:text-terminal-green-bright"
									>
										{entry.username}
									</a>
								</td>
								<td class="text-terminal-cyan">{entry.class}</td>
								<td class="text-right font-bold">{entry.score.toLocaleString()}</td>
								<td class="text-right">{getFloorName(entry.floors_cleared)}</td>
								<td class="text-right hidden sm:table-cell text-terminal-gray">
									{formatTime(entry.time_seconds)}
								</td>
								<td class="hidden md:table-cell text-terminal-gray">
									{formatDate(entry.created_at)}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>

	<!-- Legend -->
	<div class="terminal-box p-4 mt-6">
		<p class="text-terminal-gray text-sm">
			<span class="text-terminal-green">Score</span> = (floors × 100) + (kills × 10) + (items × 5) + (time bonus)
		</p>
	</div>
</div>
