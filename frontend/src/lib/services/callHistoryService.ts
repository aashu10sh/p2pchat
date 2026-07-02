import { err, ok, Result } from 'neverthrow';
import type { Error } from '$lib/entites/error';
import type { CallHistory } from '$lib/entites/callHistory';
import { writable } from 'svelte/store';

export const callHistories = writable<CallHistory[]>([]);

export default class CallHistoryService {
	async fetchCallHistories(peerId: string): Promise<Result<CallHistory[], Error>> {
		try {
			const response = await fetch(`http://localhost:8000/api/call/history?peer_id=${peerId}`, {
				method: 'GET'
			});

			if (response.ok) {
				const histories = (await response.json()) as CallHistory[];
				callHistories.set(histories || []);
				return ok(histories || []);
			} else {
				const error = (await response.json()) as Error;
				return err(error);
			}
		} catch (e) {
			return err({ error: 'Failed to fetch call histories: ' + String(e) });
		}
	}
}
