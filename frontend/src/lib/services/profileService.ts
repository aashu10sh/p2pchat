import type { Profile } from '$lib/entites/profile';
import type { Error } from '$lib/entites/error';
import { err, ok, Result } from 'neverthrow';

export default class UserService {
	async getCurrentUser(): Promise<Result<Profile, Error>> {
		try {
			const response = await fetch('http://localhost:8080/api/profile/check', {
				method: 'GET'
			});

			if (response.ok) {
				const profile = (await response.json()) as Profile;
				return ok(profile);
			} else {
				const error = (await response.json()) as Error;
				return err(error);
			}
		} catch (e) {
			const error = {
				error: 'an unknown error happened in the post request'
			};

			return err(error);
		}
	}

	async createNewProfile(userName: string): Promise<Result<Profile, Error>> {
		const response = await fetch('http://localhost:8080/api/profile', {
			method: 'POST',
			body: JSON.stringify({
				user_name: userName
			})
		});

		if (response.ok) {
			const data = (await response.json()) as Profile;
			return ok(data);
		} else {
			const error = (await response.json()) as Error;

			if (error.error.toLowerCase().includes('constraint')) {
				error.error = 'A Profile already exists for this wifi! Redirecting to dashboard!';
			}

			return err(error);
		}
	}

	async getCurrentWifi(): Promise<string> {
		const response = await fetch('http://localhost:8080/api/current-wifi', {
			method: 'GET'
		});

		const jsonData = await response.json();
		const wifiName = jsonData.wifiName;
		return wifiName;
	}
}
