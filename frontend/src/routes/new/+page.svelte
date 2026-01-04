<script lang="ts">
	import { goto } from '$app/navigation';
	import type { Error } from '$lib/entites/error';
	import type { PageProps } from './$types';
	import ProfileService from '$lib/services/profileService';
	import { onMount } from 'svelte';

	let { data }: PageProps = $props();
	let user_name = $state('');
	let error = $state<Error | null>(null);
	let message = $state<string>('');

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
	});

	async function handleUserCreation() {
		const result = await profileService.createNewProfile(user_name);

		result.match(
			(p) => {
				message = `Created profile ${p.user_name} on ${p.wifi_name}! Redirecting`;
				setTimeout(() => goto('/'), 1500);
			},
			(err) => {
				error = err;
				setTimeout(() => goto('/'), 1500);
			}
		);
	}
</script>

{#if data.wifiName}
	<h1>You are creating a profile on {data.wifiName}</h1>
	<form onsubmit={handleUserCreation}>
		<input bind:value={user_name} type="text" placeholder="Username" required />
		<button type="submit">Create Profile</button>
	</form>
{:else}
	<h1>Loading WiFi information...</h1>
{/if}

{message}

{#if error}
	<pre>{error.error}</pre>
{/if}
