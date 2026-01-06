import type { Peer } from '$lib/entites/peer';
import type { Error } from '$lib/entites/error';
import type { Message } from '$lib/entites/message';
import { err, ok, Result } from 'neverthrow';
import { writable } from 'svelte/store';
import { messages } from './chatService';

export const peers = writable<Peer[]>([]);

export default class PeerService {
	private eventSource: EventSource | null = null;
	private messageCallback: ((message: Message) => void) | null = null;

	async fetchPeers(): Promise<Result<Peer[], Error>> {
		try {
			const response = await fetch('http://localhost:8000/api/peers', {
				method: 'GET'
			});

			if (response.ok) {
				const peerList = (await response.json()) as Peer[];
				peers.set(peerList || []);
				return ok(peerList || []);
			} else {
				const error = (await response.json()) as Error;
				return err(error);
			}
		} catch (e) {
			return err({
				error: 'Failed to fetch peers: ' + String(e)
			});
		}
	}

	startEventStream(
		onPeersUpdate?: (peers: Peer[]) => void,
		onMessageReceived?: (message: Message) => void
	) {
		if (this.eventSource) {
			return; // Already connected
		}

		this.messageCallback = onMessageReceived || null;

		this.eventSource = new EventSource('http://localhost:8000/api/events');

		this.eventSource.addEventListener('connected', (event) => {
			console.log('SSE Connected:', event.data);
		});

		this.eventSource.addEventListener('peers_update', (event) => {
			try {
				const peerList = JSON.parse(event.data) as Peer[];
				peers.set(peerList || []);
				if (onPeersUpdate) {
					onPeersUpdate(peerList || []);
				}
			} catch (e) {
				console.error('Failed to parse peers update:', e);
			}
		});

		this.eventSource.addEventListener('message_received', (event) => {
			try {
				const message = JSON.parse(event.data) as Message;
				console.log('New message received:', message);

				// Update messages store
				messages.update((msgs) => [...msgs, message]);

				if (this.messageCallback) {
					this.messageCallback(message);
				}
			} catch (e) {
				console.error('Failed to parse message:', e);
			}
		});

		this.eventSource.addEventListener('message_sent', (event) => {
			try {
				const message = JSON.parse(event.data) as Message;
				console.log('Message sent confirmed:', message);

				// Update messages store
				messages.update((msgs) => [...msgs, message]);
			} catch (e) {
				console.error('Failed to parse sent message:', e);
			}
		});

		this.eventSource.onerror = (error) => {
			console.error('SSE Error:', error);
			// Reconnect logic could be added here
		};
	}

	stopEventStream() {
		if (this.eventSource) {
			this.eventSource.close();
			this.eventSource = null;
		}
	}
}
