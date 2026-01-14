<script lang="ts">
</script>

<div class="max-w-2xl mx-auto">
	<a href="/" class="text-terminal-gray hover:text-terminal-green mb-6 inline-block">
		&lt; back
	</a>

	<div class="terminal-box p-6">
		<h1 class="text-terminal-amber text-2xl mb-2">$ ssh player@dev-dungeon.com</h1>
		<p class="text-terminal-green-dim mb-6">
			Registration happens in-terminal when you first connect
		</p>

		<div class="bg-terminal-bg border border-terminal-green p-4 mb-6">
			<p class="text-terminal-green">
				Connect via SSH and your account will be created automatically.
				Your SSH key is your identity!
			</p>
		</div>
	</div>

	<!-- Instructions -->
	<div class="terminal-box p-6 mt-6">
		<h2 class="text-terminal-amber mb-4">$ cat /etc/ssh/readme</h2>

		<div class="space-y-4 text-sm">
			<div>
				<p class="text-terminal-green-bright mb-1">1. Generate a key (if you don't have one):</p>
				<code class="text-terminal-green bg-terminal-bg px-2 py-1 block">
					ssh-keygen -t ed25519
				</code>
			</div>

			<div>
				<p class="text-terminal-green-bright mb-1">2. Connect to start playing:</p>
				<code class="text-terminal-green bg-terminal-bg px-2 py-1 block">
					ssh player@dev-dungeon.com
				</code>
			</div>

			<div>
				<p class="text-terminal-green-bright mb-1">3. Choose your username when prompted</p>
				<p class="text-terminal-gray">
					First connection will guide you through account creation.
				</p>
			</div>

			<div>
				<p class="text-terminal-green-bright mb-1">4. Access your account in browser:</p>
				<p class="text-terminal-gray">
					Press <code class="text-terminal-green bg-terminal-bg px-1">Ctrl+L</code> in-game
					to generate a magic link for browser authentication.
				</p>
			</div>
		</div>
	</div>
</div>
