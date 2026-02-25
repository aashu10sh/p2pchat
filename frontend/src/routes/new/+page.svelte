<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Error } from '$lib/entites/error';
	import type { PageProps } from './$types';
	import ProfileService from '$lib/services/profileService';
	import { onMount, tick } from 'svelte';

	let { data }: PageProps = $props();
	let user_name = $state('');
	let error = $state<Error | null>(null);
	let message = $state<string>('');
	let inputRef = $state<HTMLInputElement | null>(null);
	let isFocused = $state(false);

	const profileService = new ProfileService();

	onMount(async () => {
		const already = await profileService.getCurrentUser();

		if (already.isOk()) {
			const err = {
				error: 'A Profile Already exists! Redirecting!'
			};
			error = err;
			setTimeout(() => goto('/dashboard'), 1500);
			return;
		}

		// Auto-focus the input
		await tick();
		if (inputRef) inputRef.focus();
	});

	async function handleUserCreation(e: Event) {
		e.preventDefault();
		if (!user_name.trim()) return;

		const result = await profileService.createNewProfile(user_name);

		result.match(
			(p) => {
				message = `[OK] Allocated handle: ${p.user_name} on network ${p.wifi_name}`;
				setTimeout(() => goto('/'), 1500);
			},
			(err) => {
				error = err;
				setTimeout(() => goto('/'), 1500);
			}
		);
	}

	function handleContainerClick() {
		if (inputRef) inputRef.focus();
	}
</script>

<div class="terminal-container" onclick={handleContainerClick}>
	<div class="terminal-content">
		{#if data.wifiName}
			<div class="sys-info">
				<span class="label">SYS:</span> Environment initialized
				<br />
				<span class="label">NET:</span>
				{data.wifiName}
			</div>

			<div class="prompt-text">Enter desired handle to configure node identity:</div>

			<form class="cli-form" onsubmit={handleUserCreation}>
				<span class="prompt-symbol {isFocused ? 'active' : ''}">>_</span>
				<input
					bind:this={inputRef}
					bind:value={user_name}
					type="text"
					spellcheck="false"
					autocomplete="off"
					required
					onfocus={() => (isFocused = true)}
					onblur={() => (isFocused = false)}
				/>
			</form>
		{:else}
			<div class="sys-info loading">Scanning network interfaces...</div>
		{/if}

		{#if message}
			<div class="sys-msg success">{message}</div>
		{/if}

		{#if error}
			<div class="sys-msg error">ERR: {error.error}</div>
		{/if}
	</div>
</div>

<style>
	.terminal-container {
		height: 100vh;
		width: 100vw;
		background-color: var(--bg-primary);
		color: var(--text-primary);
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: var(--font-mono);
		padding: 2rem;
		cursor: text;
	}

	.terminal-content {
		max-width: 600px;
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.sys-info {
		color: var(--text-secondary);
		font-size: 0.85rem;
		line-height: 1.6;
		margin-bottom: 1rem;
		opacity: 0.8;
	}

	.label {
		color: var(--text-accent);
		opacity: 0.7;
	}

	.loading {
		animation: pulse 2s infinite ease-in-out;
	}

	.prompt-text {
		font-size: 0.95rem;
		color: var(--text-primary);
		letter-spacing: 0.02em;
	}

	.cli-form {
		display: flex;
		align-items: center;
		margin-top: 0.5rem;
		border-bottom: 1px solid var(--border-color);
		padding-bottom: 0.5rem;
		transition: border-color 0.2s ease;
	}

	.cli-form:focus-within {
		border-bottom-color: var(--text-accent);
	}

	.prompt-symbol {
		color: var(--text-secondary);
		margin-right: 0.75rem;
		font-weight: bold;
		transition: color 0.2s ease;
	}

	.prompt-symbol.active {
		color: var(--text-accent);
	}

	input {
		flex: 1;
		background: transparent;
		border: none;
		color: var(--text-primary);
		font-family: var(--font-mono);
		font-size: 1.2rem;
		outline: none;
		letter-spacing: 0.05em;
	}

	input::selection {
		background: var(--text-accent);
		color: var(--bg-primary);
	}

	.sys-msg {
		margin-top: 1rem;
		font-size: 0.85rem;
	}

	.success {
		color: var(--text-accent);
	}

	.error {
		color: #ff3333;
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 0.5;
		}
		50% {
			opacity: 1;
		}
	}
</style>
