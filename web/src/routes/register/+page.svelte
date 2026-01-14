<script lang="ts">
	import { register } from '$lib/api';

	let username = $state('');
	let publicKey = $state('');
	let loading = $state(false);
	let error = $state('');
	let success = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		loading = true;
		error = '';
		success = '';

		const result = await register(username, publicKey);

		if (result.success && result.data) {
			success = result.data.message;
			username = '';
			publicKey = '';
		} else {
			error = result.error || 'Registration failed';
		}

		loading = false;
	}

	function isValidUsername(name: string): boolean {
		return name.length >= 3 && name.length <= 20 && /^[a-zA-Z0-9_]+$/.test(name);
	}
</script>

<div class="max-w-2xl mx-auto">
	<a href="/" class="text-terminal-gray hover:text-terminal-green mb-6 inline-block">
		&lt; back
	</a>

	<div class="terminal-box p-6">
		<h1 class="text-terminal-amber text-2xl mb-2">$ useradd</h1>
		<p class="text-terminal-green-dim mb-6">
			Register your SSH key to create an account
		</p>

		{#if success}
			<div class="bg-terminal-bg border border-terminal-green p-4 mb-6">
				<p class="text-terminal-green">{success}</p>
			</div>
		{/if}

		{#if error}
			<div class="bg-terminal-bg border border-terminal-red p-4 mb-6">
				<p class="text-terminal-red">{error}</p>
			</div>
		{/if}

		<form onsubmit={handleSubmit} class="space-y-6">
			<div>
				<label for="username" class="block text-terminal-green-bright mb-2">
					Username
				</label>
				<input
					type="text"
					id="username"
					bind:value={username}
					class="input-terminal"
					placeholder="your_username"
					minlength="3"
					maxlength="20"
					pattern="[a-zA-Z0-9_]+"
					required
				/>
				<p class="text-terminal-gray text-sm mt-1">
					3-20 characters, alphanumeric and underscore only
				</p>
			</div>

			<div>
				<label for="publicKey" class="block text-terminal-green-bright mb-2">
					SSH Public Key
				</label>
				<textarea
					id="publicKey"
					bind:value={publicKey}
					class="input-terminal min-h-[120px] resize-y"
					placeholder="ssh-ed25519 AAAA... your@email.com"
					required
				></textarea>
				<p class="text-terminal-gray text-sm mt-1">
					Paste the contents of ~/.ssh/id_ed25519.pub or ~/.ssh/id_rsa.pub
				</p>
			</div>

			<button
				type="submit"
				class="btn-terminal w-full"
				disabled={loading || !isValidUsername(username) || !publicKey}
			>
				{#if loading}
					Creating account...
				{:else}
					[Enter] Create Account
				{/if}
			</button>
		</form>
	</div>

	<!-- Instructions -->
	<div class="terminal-box p-6 mt-6">
		<h2 class="text-terminal-amber mb-4">$ cat /etc/ssh/readme</h2>

		<div class="space-y-4 text-sm">
			<div>
				<p class="text-terminal-green-bright mb-1">Generate a new key (optional):</p>
				<code class="text-terminal-green bg-terminal-bg px-2 py-1 block">
					ssh-keygen -t ed25519 -f ~/.ssh/id_devdungeon
				</code>
			</div>

			<div>
				<p class="text-terminal-green-bright mb-1">View your public key:</p>
				<code class="text-terminal-green bg-terminal-bg px-2 py-1 block">
					cat ~/.ssh/id_ed25519.pub
				</code>
			</div>

			<div>
				<p class="text-terminal-green-bright mb-1">After registering, connect:</p>
				<code class="text-terminal-green bg-terminal-bg px-2 py-1 block">
					ssh -p 2222 {username || 'username'}@devdungeon.io
				</code>
			</div>
		</div>
	</div>
</div>
