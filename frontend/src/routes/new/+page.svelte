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
				error: 'A profile already exists. Redirecting...'
			};
			error = err;
			setTimeout(() => goto('/dashboard'), 1500);
			return;
		}

		await tick();
		if (inputRef) inputRef.focus();
	});

	async function handleUserCreation(e: Event) {
		e.preventDefault();
		if (!user_name.trim()) return;

		const result = await profileService.createNewProfile(user_name);

		result.match(
			(p) => {
				message = `Welcome, ${p.user_name}! Setting things up...`;
				setTimeout(() => goto('/'), 1500);
			},
			(err) => {
				error = err;
				setTimeout(() => goto('/'), 1500);
			}
		);
	}
</script>

<div class="screen">
	<div class="card">
		<div class="card-header">
			<div class="logo-mark">
				<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
					<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
				</svg>
			</div>
			<h1>Create your profile</h1>
			<p class="subtitle">Choose a username to get started on the network.</p>
		</div>

		{#if data.wifiName}
			<div class="network-badge">
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M5 12.55a11 11 0 0 1 14.08 0"></path>
					<path d="M1.42 9a16 16 0 0 1 21.16 0"></path>
					<path d="M8.53 16.11a6 6 0 0 1 6.95 0"></path>
					<circle cx="12" cy="20" r="1"></circle>
				</svg>
				<span>{data.wifiName}</span>
			</div>

			<form class="form" onsubmit={handleUserCreation}>
				<div class="input-group" class:focused={isFocused}>
					<input
						bind:this={inputRef}
						bind:value={user_name}
						type="text"
						placeholder="Username"
						spellcheck="false"
						autocomplete="off"
						required
						onfocus={() => (isFocused = true)}
						onblur={() => (isFocused = false)}
					/>
				</div>
				<button type="submit" class="submit-btn" disabled={!user_name.trim()}>
					Continue
				</button>
			</form>
		{:else}
			<div class="loading-state">
				<div class="spinner"></div>
				<span>Detecting network...</span>
			</div>
		{/if}

		{#if message}
			<div class="feedback success">{message}</div>
		{/if}

		{#if error}
			<div class="feedback error">{error.error}</div>
		{/if}
	</div>
</div>

<style>
	.screen {
		height: 100vh;
		width: 100vw;
		background-color: var(--bg-secondary);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 24px;
	}

	.card {
		background: var(--bg-primary);
		border-radius: var(--radius-2xl);
		padding: 48px 40px;
		width: 100%;
		max-width: 420px;
		box-shadow: var(--shadow-lg);
		border: 1px solid var(--border-color);
		animation: slideUp 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
	}

	.card-header {
		text-align: center;
		margin-bottom: 32px;
	}

	.logo-mark {
		width: 52px;
		height: 52px;
		border-radius: var(--radius-lg);
		background: var(--accent-light);
		color: var(--accent);
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto 20px;
	}

	.card-header h1 {
		font-size: 22px;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 8px;
		letter-spacing: -0.02em;
	}

	.subtitle {
		font-size: 14px;
		color: var(--text-secondary);
		margin: 0;
		line-height: 1.5;
	}

	.network-badge {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 14px;
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
		font-size: 13px;
		color: var(--text-secondary);
		margin-bottom: 24px;
	}

	.form {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.input-group {
		border-radius: var(--radius-md);
		border: 1.5px solid var(--border-hover);
		transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
		overflow: hidden;
	}

	.input-group.focused {
		border-color: var(--accent);
		box-shadow: 0 0 0 3px var(--accent-light);
	}

	input {
		width: 100%;
		padding: 14px 16px;
		background: transparent;
		border: none;
		color: var(--text-primary);
		font-family: var(--font-sans);
		font-size: 16px;
		outline: none;
	}

	input::placeholder {
		color: var(--text-tertiary);
	}

	.submit-btn {
		width: 100%;
		padding: 14px;
		background: var(--accent);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-family: var(--font-sans);
		font-size: 15px;
		font-weight: 600;
		cursor: pointer;
		transition: background var(--transition-fast), transform var(--transition-fast);
	}

	.submit-btn:hover:not(:disabled) {
		background: var(--accent-hover);
	}

	.submit-btn:active:not(:disabled) {
		transform: scale(0.98);
	}

	.submit-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.loading-state {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 10px;
		padding: 20px;
		color: var(--text-secondary);
		font-size: 14px;
	}

	.spinner {
		width: 18px;
		height: 18px;
		border: 2px solid var(--bg-tertiary);
		border-top-color: var(--accent);
		border-radius: 50%;
		animation: spin 0.7s linear infinite;
	}

	.feedback {
		margin-top: 16px;
		padding: 12px 14px;
		border-radius: var(--radius-sm);
		font-size: 13px;
		text-align: center;
		animation: fadeIn 0.3s ease-out;
	}

	.feedback.success {
		background: rgba(52, 199, 89, 0.08);
		color: var(--success);
	}

	.feedback.error {
		background: rgba(255, 59, 48, 0.08);
		color: var(--danger);
	}

	@keyframes slideUp {
		from {
			opacity: 0;
			transform: translateY(16px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}
</style>
