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

<div class="screen">
	<div class="content">
		{#if isFetching}
			<div class="loader">
				<div class="spinner"></div>
				<p class="label">Connecting...</p>
			</div>
		{:else if error}
			<div class="loader">
				<div class="icon">👋</div>
				<p class="label">Welcome</p>
				<p class="sublabel">Setting up your profile...</p>
			</div>
		{:else if profile}
			<div class="loader">
				<div class="avatar">{profile.user_name.charAt(0).toUpperCase()}</div>
				<p class="label">{profile.user_name}</p>
				<p class="sublabel">Loading your conversations...</p>
			</div>
		{/if}
	</div>
</div>

<style>
	.screen {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100vh;
		width: 100vw;
		background-color: var(--bg-primary);
	}

	.content {
		animation: fadeIn 0.6s ease-out;
	}

	.loader {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 16px;
	}

	.spinner {
		width: 32px;
		height: 32px;
		border: 2.5px solid var(--bg-tertiary);
		border-top-color: var(--accent);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	.icon {
		font-size: 40px;
		animation: float 2s ease-in-out infinite;
	}

	.avatar {
		width: 56px;
		height: 56px;
		border-radius: 50%;
		background: linear-gradient(135deg, var(--accent), #5856d6);
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 22px;
		font-weight: 600;
	}

	.label {
		font-size: 17px;
		font-weight: 500;
		color: var(--text-primary);
		margin: 0;
	}

	.sublabel {
		font-size: 13px;
		color: var(--text-secondary);
		margin: 0;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@keyframes float {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-6px);
		}
	}
</style>
