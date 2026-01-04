import type { PageLoad } from './$types';
import ProfileService from '$lib/services/profileService';

export const load: PageLoad = async () => {
	const profileService = new ProfileService();
	const wifi = await profileService.getCurrentWifi();
	return {
		wifiName: wifi
	};
};
