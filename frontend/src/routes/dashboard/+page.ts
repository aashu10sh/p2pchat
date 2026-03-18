import { goto } from '$app/navigation';
import ProfileService from '$lib/services/profileService';
import type { PageLoad } from './$types';
// going to need, user, knownPeers, onlinePeers

export const load: PageLoad = async () => {
	const profileService = new ProfileService();
	const userResult = await profileService.getCurrentUser();

	if (userResult.isErr()) {
		goto('/new');
	}

	let user = userResult._unsafeUnwrap();

	return {
		user: user
	};
};
