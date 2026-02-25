<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Profile } from '$lib/entites/profile';
	import type { Error } from '$lib/entites/error';
	import { onMount } from 'svelte';
	import ProfileService from '$lib/services/profileService';

	let profile = $state<Profile>();
	let isFetching = $state<boolean>(true);
	let error = $state<Error | null>(null);

	const profileService = new ProfileService();

	onMount(async () => {
		const result = await profileService.getCurrentUser();

		result.match(
			(_profile) => {
				profile = _profile;
				setTimeout(() => goto('/dashboard'), 1500);
			},
			(_error) => {
				error = _error;
				setTimeout(() => goto('/new'), 1500);
			}
		);
		isFetching = false;
	});
</script>

<div class="screen-wrapper">
	{#if isFetching}
		<div class="status connecting">
			<span class="cursor">_</span> Initialize connection...
		</div>
	{:else if error}
		<div class="status error-state">
			<div class="err-text">ERR: {error.error}</div>
			<div class="sub-text">Redirecting to new profile...</div>
		</div>
	{:else if profile}
		<div class="status success">
			<div class="auth-text">AUTH OK: <span class="accent">{profile.user_name}</span></div>
			<div class="sub-text">Loading dashboard module...</div>
		</div>
	{/if}
</div>

<style>
	.screen-wrapper {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100vh;
		width: 100vw;
		background-color: var(--bg-primary);
		color: var(--text-primary);
	}

	.status {
		font-family: var(--font-mono);
		font-size: 14px;
		letter-spacing: 0.05em;
	}

	.connecting {
		color: var(--text-secondary);
	}

	.error-state .err-text {
		color: #ff3333;
		margin-bottom: 8px;
	}

	.success .auth-text {
		color: var(--text-secondary);
		margin-bottom: 8px;
	}

	.accent {
		color: var(--text-accent);
	}

	.sub-text {
		font-size: 12px;
		color: var(--text-secondary);
		opacity: 0.5;
	}

	.cursor {
		display: inline-block;
		animation: blink 1s step-end infinite;
		color: var(--text-accent);
		margin-right: 4px;
	}

	@keyframes blink {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0;
		}
	}
</style>
