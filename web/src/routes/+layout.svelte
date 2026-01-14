<script lang="ts">
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<title>/dev/dungeon</title>
	<meta name="description" content="A Unix-themed terminal roguelike" />
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
</svelte:head>

<div class="crt min-h-screen">
	<div class="max-w-6xl mx-auto px-4 py-8">
		{@render children()}
	</div>
</div>
