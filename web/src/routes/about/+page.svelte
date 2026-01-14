<script lang="ts">
	const classes = [
		{ name: 'init', desc: 'Balanced starter. Jack of all trades.', stats: 'Balanced' },
		{ name: 'cron', desc: 'Time-based abilities. Lower NICE = more turns.', stats: 'High Speed' },
		{ name: 'bash', desc: 'Script combos. Chain attacks for bonus damage.', stats: 'High CPU' },
		{ name: 'vim', desc: 'Modal abilities. Complex but powerful.', stats: 'High MEM' },
		{ name: 'sudo', desc: 'Permission bypasses. Access locked areas.', stats: 'High UID' },
	];

	const floors = [
		{ path: '/home', desc: 'Tutorial zone. Minor bugs and lint errors.', difficulty: 'Easy' },
		{ path: '/tmp', desc: 'Temporary chaos. Orphan processes roam.', difficulty: 'Easy-Med' },
		{ path: '/var', desc: 'Variable dangers. Log daemons patrol.', difficulty: 'Medium' },
		{ path: '/etc', desc: 'Configuration maze. Permission errors await.', difficulty: 'Med-Hard' },
		{ path: '/usr', desc: 'User space. Privileged processes guard.', difficulty: 'Hard' },
		{ path: '/sys', desc: 'System level. Kernel threads lurk.', difficulty: 'Very Hard' },
		{ path: '/dev', desc: 'Device layer. I/O errors are deadly.', difficulty: 'Brutal' },
		{ path: '/dev/null', desc: 'The void. Kernel Panic awaits.', difficulty: 'BOSS' },
	];
</script>

<div class="max-w-4xl mx-auto">
	<a href="/" class="text-terminal-gray hover:text-terminal-green mb-6 inline-block">
		&lt; back
	</a>

	<div class="terminal-box p-6 mb-6">
		<h1 class="text-terminal-amber text-2xl mb-2">$ man devdungeon</h1>
		<p class="text-terminal-green-dim">
			/dev/dungeon(1) - Unix-themed terminal roguelike dungeon crawler
		</p>
	</div>

	<!-- SYNOPSIS -->
	<div class="terminal-box p-6 mb-6">
		<h2 class="text-terminal-amber mb-4">SYNOPSIS</h2>
		<div class="text-terminal-green-dim">
			<p class="mb-4">
				You are a rogue process. The system is corrupted. Descend through the filesystem,
				battle daemons and zombie processes, and reach <span class="text-terminal-red">/dev/null</span> to defeat the Kernel Panic.
			</p>
			<p>
				Death is not the end — it's an upgrade opportunity. Earn <span class="text-terminal-amber">exit codes</span> and
				unlock new classes, permanent stat bonuses, and items for future runs.
			</p>
		</div>
	</div>

	<!-- STATS -->
	<div class="terminal-box p-6 mb-6">
		<h2 class="text-terminal-amber mb-4">STATS</h2>
		<div class="grid sm:grid-cols-2 gap-4">
			<div class="bg-terminal-bg p-3">
				<span class="text-terminal-green font-bold">PID</span>
				<span class="text-terminal-gray"> - Process ID (Health)</span>
				<p class="text-terminal-green-dim text-sm mt-1">Your life force. Reach 0 and you're killed.</p>
			</div>
			<div class="bg-terminal-bg p-3">
				<span class="text-terminal-green font-bold">CPU</span>
				<span class="text-terminal-gray"> - Processing Power (Attack)</span>
				<p class="text-terminal-green-dim text-sm mt-1">Damage dealt with basic attacks.</p>
			</div>
			<div class="bg-terminal-bg p-3">
				<span class="text-terminal-green font-bold">MEM</span>
				<span class="text-terminal-gray"> - Memory (Mana)</span>
				<p class="text-terminal-green-dim text-sm mt-1">Resource for special abilities and hacks.</p>
			</div>
			<div class="bg-terminal-bg p-3">
				<span class="text-terminal-green font-bold">NICE</span>
				<span class="text-terminal-gray"> - Priority (Speed)</span>
				<p class="text-terminal-green-dim text-sm mt-1">Lower is faster. Determines turn order.</p>
			</div>
			<div class="bg-terminal-bg p-3 sm:col-span-2">
				<span class="text-terminal-green font-bold">UID</span>
				<span class="text-terminal-gray"> - User ID (Permissions)</span>
				<p class="text-terminal-green-dim text-sm mt-1">Access level. Higher UID unlocks restricted areas and chests.</p>
			</div>
		</div>
	</div>

	<!-- CLASSES -->
	<div class="terminal-box p-6 mb-6">
		<h2 class="text-terminal-amber mb-4">CLASSES</h2>
		<div class="space-y-3">
			{#each classes as cls}
				<div class="flex items-start gap-4 py-2 border-b border-terminal-gray last:border-0">
					<span class="text-terminal-cyan font-bold w-16">{cls.name}</span>
					<span class="flex-1 text-terminal-green-dim">{cls.desc}</span>
					<span class="text-terminal-gray text-sm hidden sm:block">{cls.stats}</span>
				</div>
			{/each}
		</div>
	</div>

	<!-- FLOORS -->
	<div class="terminal-box p-6 mb-6">
		<h2 class="text-terminal-amber mb-4">THE DESCENT</h2>
		<div class="space-y-2">
			{#each floors as floor, i}
				<div class="flex items-start gap-4 py-2 border-b border-terminal-gray last:border-0">
					<span class="text-terminal-green font-bold w-24 {floor.path === '/dev/null' ? 'text-terminal-red' : ''}">{floor.path}</span>
					<span class="flex-1 text-terminal-green-dim text-sm">{floor.desc}</span>
					<span class="text-terminal-gray text-xs hidden sm:block">{floor.difficulty}</span>
				</div>
			{/each}
		</div>
	</div>

	<!-- CONTROLS -->
	<div class="terminal-box p-6 mb-6">
		<h2 class="text-terminal-amber mb-4">CONTROLS</h2>
		<div class="grid sm:grid-cols-2 gap-6">
			<div>
				<h3 class="text-terminal-green-bright mb-2">Exploration</h3>
				<div class="text-sm space-y-1">
					<p><span class="text-terminal-green">WASD / Arrow keys</span> <span class="text-terminal-gray">- Move</span></p>
					<p><span class="text-terminal-green">I</span> <span class="text-terminal-gray">- Inventory</span></p>
					<p><span class="text-terminal-green">P / Esc</span> <span class="text-terminal-gray">- Pause</span></p>
				</div>
			</div>
			<div>
				<h3 class="text-terminal-green-bright mb-2">Combat</h3>
				<div class="text-sm space-y-1">
					<p><span class="text-terminal-green">A / 1</span> <span class="text-terminal-gray">- Attack</span></p>
					<p><span class="text-terminal-green">H / 2</span> <span class="text-terminal-gray">- Hack (ability)</span></p>
					<p><span class="text-terminal-green">I / 3</span> <span class="text-terminal-gray">- Use item</span></p>
					<p><span class="text-terminal-green">F / 4</span> <span class="text-terminal-gray">- Flee</span></p>
				</div>
			</div>
		</div>
	</div>

	<!-- SEE ALSO -->
	<div class="terminal-box p-6">
		<h2 class="text-terminal-amber mb-4">SEE ALSO</h2>
		<div class="text-terminal-green-dim text-sm space-y-1">
			<p>
				<a href="https://github.com/iheanyi/devdungeon" class="text-terminal-green hover:glow">
					github.com/iheanyi/devdungeon
				</a>
				<span class="text-terminal-gray"> - Source code</span>
			</p>
			<p>
				<a href="https://github.com/charmbracelet/bubbletea" class="text-terminal-green hover:glow">
					Bubble Tea
				</a>
				<span class="text-terminal-gray"> - TUI framework</span>
			</p>
			<p>
				<a href="https://github.com/charmbracelet/wish" class="text-terminal-green hover:glow">
					Wish
				</a>
				<span class="text-terminal-gray"> - SSH server library</span>
			</p>
		</div>
	</div>
</div>
