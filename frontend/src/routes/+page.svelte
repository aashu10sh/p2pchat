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

{#if isFetching}
	<div class="loading">
		<h1>Loading...</h1>
	</div>
{:else if error}
	<div class="error">
		<!-- <h1>Error</h1> -->
		<pre>{error.error}</pre>
		<pre>Redirecting!</pre>
	</div>
{:else if profile}
	<div class="success">
		<h1>Hi {profile.user_name}!</h1>
		<h2>Redirecting you to the dashboard...</h2>
	</div>
{/if}
