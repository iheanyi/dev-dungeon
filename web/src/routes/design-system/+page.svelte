<script lang="ts">
	const colors = [
		{ name: 'terminal-bg', value: '#0a0a0a', desc: 'Primary background' },
		{ name: 'terminal-bg-light', value: '#121212', desc: 'Elevated surfaces' },
		{ name: 'terminal-green', value: '#00ff00', desc: 'Primary text, accents' },
		{ name: 'terminal-green-dim', value: '#00cc00', desc: 'Secondary text' },
		{ name: 'terminal-green-bright', value: '#33ff33', desc: 'Highlights, focus states' },
		{ name: 'terminal-amber', value: '#ffb000', desc: 'Warnings, prompts' },
		{ name: 'terminal-amber-dim', value: '#cc8c00', desc: 'Muted warnings' },
		{ name: 'terminal-red', value: '#ff3333', desc: 'Errors, danger, enemies' },
		{ name: 'terminal-cyan', value: '#00ffff', desc: 'Info, links' },
		{ name: 'terminal-purple', value: '#cc00ff', desc: 'Magic, special' },
		{ name: 'terminal-gray', value: '#9a9a9a', desc: 'Muted text, borders' },
		{ name: 'terminal-gray-light', value: '#bbbbbb', desc: 'Placeholder text' }
	];

	const enemies = [
		{ glyph: 'z', name: 'zombie', desc: 'Undead process' },
		{ glyph: 'd', name: 'daemon', desc: 'Background process' },
		{ glyph: 'f', name: 'fork_bomb', desc: 'Recursive spawner' },
		{ glyph: 's', name: 'segfault', desc: 'Memory corruptor' },
		{ glyph: 'K', name: 'KERNEL PANIC', desc: 'Final boss' }
	];

	const stats = [
		{ name: 'RAM', maps: 'Health', desc: 'When you run out, OOM killer gets you' },
		{ name: 'CPU', maps: 'Attack', desc: 'Processing power for damage' },
		{ name: 'FD', maps: 'Mana/Abilities', desc: 'File descriptors are limited resources' },
		{ name: 'NICE', maps: 'Speed', desc: 'Lower nice = higher priority' },
		{ name: 'UID', maps: 'Access Level', desc: '0 = root = god mode' }
	];

	const floors = [
		{ path: '/home', desc: 'Starting area, user space' },
		{ path: '/tmp', desc: 'Temporary, volatile' },
		{ path: '/var', desc: 'Variable data, logs' },
		{ path: '/etc', desc: 'Configuration files' },
		{ path: '/usr', desc: 'User programs' },
		{ path: '/sys', desc: 'System internals' },
		{ path: '/dev', desc: 'Device files' },
		{ path: '/dev/null', desc: 'The void. Final floor.' }
	];
</script>

<svelte:head>
	<title>Design System | /dev/dungeon</title>
	<meta name="description" content="The design system and visual language of /dev/dungeon" />
</svelte:head>

<div class="max-w-4xl mx-auto">
	<header class="mb-12">
		<h1 class="text-terminal-green text-3xl md:text-4xl font-bold glow mb-4">Design System</h1>
		<p class="text-terminal-green-dim text-lg">
			The visual language of <span class="text-terminal-green">/dev/dungeon</span> — a terminal
			aesthetic inspired by CRT monitors, Unix systems, and classic roguelikes.
		</p>
	</header>

	<!-- Philosophy -->
	<section class="terminal-box p-6 mb-8">
		<h2 class="text-terminal-amber text-xl mb-4">$ cat /etc/design-philosophy</h2>
		<div class="text-terminal-green-dim space-y-4">
			<p>
				Every visual choice reinforces the <span class="text-terminal-green">terminal/Unix</span>
				theme. We're not just making a game that looks like a terminal — we're making a game that
				<em>feels</em> like you're SSH'd into a haunted server.
			</p>
			<ul class="list-disc list-inside space-y-2 ml-4">
				<li><span class="text-terminal-green">Monospace everything</span> — JetBrains Mono primary</li>
				<li><span class="text-terminal-green">Phosphor glow</span> — text-shadow for that CRT feel</li>
				<li><span class="text-terminal-green">Scanlines</span> — subtle overlay for authenticity</li>
				<li><span class="text-terminal-green">Dark-first</span> — #0a0a0a is our canvas</li>
				<li><span class="text-terminal-green">Green dominance</span> — classic terminal phosphor</li>
			</ul>
		</div>
	</section>

	<!-- Colors -->
	<section class="terminal-box p-6 mb-8">
		<h2 class="text-terminal-amber text-xl mb-4">$ ls /etc/colors/</h2>
		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			{#each colors as color}
				<div class="flex items-center gap-4 p-3 bg-terminal-bg rounded">
					<div
						class="w-12 h-12 rounded border border-terminal-gray flex-shrink-0"
						style="background-color: {color.value};"
					></div>
					<div>
						<code class="text-terminal-green text-sm">{color.name}</code>
						<div class="text-terminal-gray text-xs">{color.value}</div>
						<div class="text-terminal-green-dim text-xs">{color.desc}</div>
					</div>
				</div>
			{/each}
		</div>
	</section>

	<!-- Typography -->
	<section class="terminal-box p-6 mb-8">
		<h2 class="text-terminal-amber text-xl mb-4">$ fc-list | grep mono</h2>
		<div class="space-y-6">
			<div>
				<h3 class="text-terminal-green-bright mb-2">Primary Font</h3>
				<p class="text-terminal-green text-2xl font-mono">JetBrains Mono</p>
				<p class="text-terminal-green-dim text-sm mt-1">
					Fallbacks: Fira Code, SF Mono, Monaco, Inconsolata, monospace
				</p>
			</div>

			<div>
				<h3 class="text-terminal-green-bright mb-2">Type Scale</h3>
				<div class="space-y-2">
					<p class="text-4xl text-terminal-green glow">Heading 1 — 2.25rem</p>
					<p class="text-2xl text-terminal-green">Heading 2 — 1.5rem</p>
					<p class="text-xl text-terminal-green-dim">Heading 3 — 1.25rem</p>
					<p class="text-base text-terminal-green-dim">Body — 1rem</p>
					<p class="text-sm text-terminal-gray">Small — 0.875rem</p>
					<p class="text-xs text-terminal-gray">Caption — 0.75rem</p>
				</div>
			</div>
		</div>
	</section>

	<!-- Effects -->
	<section class="terminal-box p-6 mb-8">
		<h2 class="text-terminal-amber text-xl mb-4">$ cat /proc/effects</h2>
		<div class="space-y-6">
			<div>
				<h3 class="text-terminal-green-bright mb-2">Phosphor Glow</h3>
				<p class="text-terminal-green glow text-xl mb-2">This text has the glow effect</p>
				<code class="text-terminal-gray text-sm">text-shadow: 0 0 5px currentColor, 0 0 10px currentColor;</code>
			</div>

			<div>
				<h3 class="text-terminal-green-bright mb-2">Strong Glow</h3>
				<p class="text-terminal-green glow-strong text-xl mb-2">This text has strong glow</p>
				<code class="text-terminal-gray text-sm">text-shadow: 0 0 5px, 0 0 10px, 0 0 20px currentColor;</code>
			</div>

			<div>
				<h3 class="text-terminal-green-bright mb-2">Cursor Blink</h3>
				<p class="text-terminal-green text-xl">
					Command prompt<span class="cursor"></span>
				</p>
				<code class="text-terminal-gray text-sm">animation: blink 1s step-end infinite;</code>
			</div>
		</div>
	</section>

	<!-- Components -->
	<section class="terminal-box p-6 mb-8">
		<h2 class="text-terminal-amber text-xl mb-4">$ ls /usr/lib/components/</h2>
		<div class="space-y-6">
			<div>
				<h3 class="text-terminal-green-bright mb-2">Terminal Box</h3>
				<div class="terminal-box p-4">
					<p class="text-terminal-green-dim">A container with border glow and hover states.</p>
				</div>
				<code class="text-terminal-gray text-sm block mt-2">.terminal-box</code>
			</div>

			<div>
				<h3 class="text-terminal-green-bright mb-2">Buttons</h3>
				<div class="flex gap-4 flex-wrap">
					<button class="btn-terminal">Normal</button>
					<button class="btn-terminal" disabled>Disabled</button>
				</div>
				<code class="text-terminal-gray text-sm block mt-2">.btn-terminal</code>
			</div>

			<div>
				<h3 class="text-terminal-green-bright mb-2">Input</h3>
				<input type="text" class="input-terminal" placeholder="Enter command..." />
				<code class="text-terminal-gray text-sm block mt-2">.input-terminal</code>
			</div>

			<div>
				<h3 class="text-terminal-green-bright mb-2">Table</h3>
				<table class="table-terminal text-sm">
					<thead>
						<tr>
							<th>PID</th>
							<th>USER</th>
							<th>COMMAND</th>
						</tr>
					</thead>
					<tbody>
						<tr>
							<td class="text-terminal-green">1337</td>
							<td>player</td>
							<td>./devdungeon</td>
						</tr>
						<tr>
							<td class="text-terminal-red">666</td>
							<td>daemon</td>
							<td>fork_bomb</td>
						</tr>
					</tbody>
				</table>
				<code class="text-terminal-gray text-sm block mt-2">.table-terminal</code>
			</div>
		</div>
	</section>

	<!-- Game Elements -->
	<section class="terminal-box p-6 mb-8">
		<h2 class="text-terminal-amber text-xl mb-4">$ man devdungeon</h2>

		<div class="space-y-6">
			<div>
				<h3 class="text-terminal-green-bright mb-3">Enemy Glyphs</h3>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-2">
					{#each enemies as enemy}
						<div class="flex items-center gap-4 p-2">
							<span class="text-terminal-red text-2xl font-bold w-8 text-center">{enemy.glyph}</span>
							<div>
								<span class="text-terminal-green">{enemy.name}</span>
								<span class="text-terminal-gray text-sm ml-2">— {enemy.desc}</span>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<div>
				<h3 class="text-terminal-green-bright mb-3">Stat Mappings</h3>
				<div class="space-y-2">
					{#each stats as stat}
						<div class="flex items-center gap-4 p-2 border-b border-terminal-gray/20">
							<span class="text-terminal-green font-bold w-12">{stat.name}</span>
							<span class="text-terminal-amber w-20">{stat.maps}</span>
							<span class="text-terminal-green-dim text-sm">{stat.desc}</span>
						</div>
					{/each}
				</div>
			</div>

			<div>
				<h3 class="text-terminal-green-bright mb-3">The Descent</h3>
				<div class="font-mono text-sm">
					{#each floors as floor, i}
						<div class="flex items-center gap-2">
							<span class="text-terminal-gray">{i + 1}.</span>
							<span class={floor.path === '/dev/null' ? 'text-terminal-red' : 'text-terminal-green'}>
								{floor.path}
							</span>
							<span class="text-terminal-gray">—</span>
							<span class="text-terminal-green-dim">{floor.desc}</span>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</section>

	<!-- ASCII Art -->
	<section class="terminal-box p-6 mb-8">
		<h2 class="text-terminal-amber text-xl mb-4">$ figlet "ASCII"</h2>
		<pre class="ascii-art text-terminal-green glow text-xs md:text-sm overflow-x-auto">
██████╗ ███████╗██╗   ██╗██████╗ ██╗   ██╗███╗   ██╗ ██████╗ ███████╗ ██████╗ ███╗   ██╗
██╔══██╗██╔════╝██║   ██║██╔══██╗██║   ██║████╗  ██║██╔════╝ ██╔════╝██╔═══██╗████╗  ██║
██║  ██║█████╗  ██║   ██║██║  ██║██║   ██║██╔██╗ ██║██║  ███╗█████╗  ██║   ██║██╔██╗ ██║
██║  ██║██╔══╝  ╚██╗ ██╔╝██║  ██║██║   ██║██║╚██╗██║██║   ██║██╔══╝  ██║   ██║██║╚██╗██║
██████╔╝███████╗ ╚████╔╝ ██████╔╝╚██████╔╝██║ ╚████║╚██████╔╝███████╗╚██████╔╝██║ ╚████║
╚═════╝ ╚══════╝  ╚═══╝  ╚═════╝  ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝
		</pre>
		<p class="text-terminal-green-dim text-sm mt-4">
			ASCII art uses box-drawing characters (╔═╗║╚╝) and Unicode block elements (█▀▄).
			Keep line-height tight (1.2) and use <code class="text-terminal-green">.ascii-art</code> class.
		</p>
	</section>

	<!-- Credits -->
	<footer class="text-center text-terminal-gray text-sm mt-12 mb-8">
		<p>Design system by <a href="https://twitter.com/kwuchu" class="text-terminal-green-dim hover:text-terminal-green">@kwuchu</a></p>
		<p class="mt-2">
			<a href="/" class="text-terminal-green-dim hover:text-terminal-green">← Back to /dev/dungeon</a>
		</p>
	</footer>
</div>
