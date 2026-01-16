<script lang="ts">
	import { onMount } from 'svelte';

	let showCursor = $state(true);
	let typedText = $state('');
	const fullText = 'Navigate the filesystem. Survive.';

	onMount(() => {
		let i = 0;
		const interval = setInterval(() => {
			if (i < fullText.length) {
				typedText = fullText.slice(0, i + 1);
				i++;
			} else {
				clearInterval(interval);
			}
		}, 50);

		return () => clearInterval(interval);
	});
</script>

<div class="text-center mb-8 md:mb-12">
	<!-- ASCII Logo - Desktop -->
	<pre class="ascii-art text-terminal-green glow hidden md:inline-block text-left text-xs lg:text-sm">
██████╗ ███████╗██╗   ██╗██████╗ ██╗   ██╗███╗   ██╗ ██████╗ ███████╗ ██████╗ ███╗   ██╗
██╔══██╗██╔════╝██║   ██║██╔══██╗██║   ██║████╗  ██║██╔════╝ ██╔════╝██╔═══██╗████╗  ██║
██║  ██║█████╗  ██║   ██║██║  ██║██║   ██║██╔██╗ ██║██║  ███╗█████╗  ██║   ██║██╔██╗ ██║
██║  ██║██╔══╝  ╚██╗ ██╔╝██║  ██║██║   ██║██║╚██╗██║██║   ██║██╔══╝  ██║   ██║██║╚██╗██║
██████╔╝███████╗ ╚████╔╝ ██████╔╝╚██████╔╝██║ ╚████║╚██████╔╝███████╗╚██████╔╝██║ ╚████║
╚═════╝ ╚══════╝  ╚═══╝  ╚═════╝  ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝
	</pre>

	<!-- Mobile Logo -->
	<h1 class="md:hidden text-terminal-green glow text-3xl font-bold tracking-wider">/dev/dungeon</h1>

	<p class="text-terminal-green-dim text-lg md:text-xl mt-4 font-mono">
		<span class="text-terminal-gray">&gt;</span> {typedText}<span class="cursor"></span>
	</p>
</div>

<!-- Quick Connect -->
<div class="terminal-box p-4 md:p-6 mb-6 md:mb-8">
	<h2 class="text-terminal-amber text-base md:text-lg mb-4">$ ssh player@dev-dungeon.com</h2>
	<p class="text-terminal-green-dim mb-4 text-sm md:text-base">
		Connect via SSH to play. Your public key is your identity.
	</p>
	<div class="bg-terminal-bg p-3 md:p-4 font-mono text-xs md:text-sm overflow-x-auto">
		<p class="text-terminal-gray"># Generate a key (if needed)</p>
		<p class="text-terminal-green whitespace-nowrap">ssh-keygen -t ed25519</p>
		<p class="text-terminal-gray mt-2"># Connect to play</p>
		<p class="text-terminal-green whitespace-nowrap">ssh player@dev-dungeon.com</p>
	</div>
</div>

<!-- Main Grid -->
<div class="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6 mb-6 md:mb-8">
	<!-- Register Card -->
	<a href="/register" class="terminal-box p-4 md:p-6 block transition-all hover:border-terminal-green">
		<h3 class="text-terminal-amber mb-2 flex items-center gap-2">
			<span class="text-terminal-green">[</span>
			Register
			<span class="text-terminal-green">]</span>
		</h3>
		<p class="text-terminal-green-dim text-sm">
			Create your account by linking your SSH public key to a username.
		</p>
	</a>

	<!-- Leaderboard Card -->
	<a href="/leaderboard" class="terminal-box p-4 md:p-6 block transition-all hover:border-terminal-green">
		<h3 class="text-terminal-amber mb-2 flex items-center gap-2">
			<span class="text-terminal-green">[</span>
			Leaderboard
			<span class="text-terminal-green">]</span>
		</h3>
		<p class="text-terminal-green-dim text-sm">
			See who has descended the deepest and defeated the Kernel Panic.
		</p>
	</a>

	<!-- Daily Run Card -->
	<a href="/daily" class="terminal-box p-4 md:p-6 block transition-all hover:border-terminal-green">
		<h3 class="text-terminal-amber mb-2 flex items-center gap-2">
			<span class="text-terminal-green">[</span>
			Daily Run
			<span class="text-terminal-green">]</span>
		</h3>
		<p class="text-terminal-green-dim text-sm">
			Today's seed challenge. Same dungeon for everyone. Compete globally.
		</p>
	</a>

	<!-- About Card -->
	<a href="/about" class="terminal-box p-4 md:p-6 block transition-all hover:border-terminal-green">
		<h3 class="text-terminal-amber mb-2 flex items-center gap-2">
			<span class="text-terminal-green">[</span>
			man devdungeon
			<span class="text-terminal-green">]</span>
		</h3>
		<p class="text-terminal-green-dim text-sm">
			Read the manual. Learn the lore. Understand the filesystem.
		</p>
	</a>
</div>

<!-- Game Info -->
<div class="terminal-box p-4 md:p-6">
	<h2 class="text-terminal-amber mb-4 text-base md:text-lg">$ cat /etc/motd</h2>
	<div class="grid grid-cols-2 md:grid-cols-3 gap-4 md:gap-6 text-xs md:text-sm">
		<div>
			<h4 class="text-terminal-green-bright mb-2">STATS</h4>
			<ul class="text-terminal-green-dim space-y-1">
				<li><span class="text-terminal-green">PID</span> - Health</li>
				<li><span class="text-terminal-green">CPU</span> - Attack</li>
				<li><span class="text-terminal-green">MEM</span> - Mana</li>
				<li><span class="text-terminal-green">NICE</span> - Speed</li>
				<li><span class="text-terminal-green">UID</span> - Access</li>
			</ul>
		</div>
		<div>
			<h4 class="text-terminal-green-bright mb-2">ENEMIES</h4>
			<ul class="text-terminal-green-dim space-y-1">
				<li><span class="text-terminal-red">z</span> - zombie</li>
				<li><span class="text-terminal-red">d</span> - daemon</li>
				<li><span class="text-terminal-red">f</span> - fork_bomb</li>
				<li><span class="text-terminal-red">s</span> - segfault</li>
				<li><span class="text-terminal-red">K</span> - KERNEL PANIC</li>
			</ul>
		</div>
		<div>
			<h4 class="text-terminal-green-bright mb-2">THE DESCENT</h4>
			<ul class="text-terminal-green-dim space-y-1">
				<li>/home → /tmp</li>
				<li>/var → /etc</li>
				<li>/usr → /sys</li>
				<li>/dev → <span class="text-terminal-red">/dev/null</span></li>
			</ul>
		</div>
	</div>
</div>

<!-- Footer -->
<footer class="mt-12 text-center text-terminal-gray text-sm">
	<p>Built with <a href="https://github.com/charmbracelet/bubbletea" class="text-terminal-green-dim hover:text-terminal-green">Bubble Tea</a> & <a href="https://github.com/charmbracelet/wish" class="text-terminal-green-dim hover:text-terminal-green">Wish</a></p>
	<p class="mt-2">
		<a href="https://github.com/iheanyi/dev-dungeon" class="text-terminal-green-dim hover:text-terminal-green">
			github.com/iheanyi/dev-dungeon
		</a>
	</p>
</footer>
