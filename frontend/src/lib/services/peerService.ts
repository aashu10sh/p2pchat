import type { Peer } from '$lib/entites/peer';
import type { Error } from '$lib/entites/error';
import type { Message } from '$lib/entites/message';
import { err, ok, Result } from 'neverthrow';
import { writable } from 'svelte/store';
import { messages } from './chatService';
import {
	handleRemoteOffer,
	handleRemoteAnswer,
	handleRemoteICECandidate,
	handleRemoteHangup
} from './callService';

export const peers = writable<Peer[]>([]);

export default class PeerService {
	private eventSource: EventSource | null = null;
	private messageCallback: ((message: Message) => void) | null = null;

	async fetchPeer(peerId: string): Promise<Result<Peer, Error>> {
		try {
			const response = await fetch(`http://localhost:8000/api/profile?peer_id=${peerId}`, {
				method: 'GET'
			});
			if (response.ok) {
				const peer = (await response.json()) as Peer;
				return ok(peer);
			} else {
				const error = (await response.json()) as Error;
				return err(error);
			}
		} catch (e) {
			return err({
				error: 'Failed to fetch peer' + String(e)
			});
		}
	}

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

		// Video call signaling events
		this.eventSource.addEventListener('video_call_offer', (event) => {
			try {
				const data = JSON.parse(event.data);
				// Resolve username from peers store for the incoming call modal
				let fromUsername = data.from_peer_id.substring(0, 8) + '...';
				const unsub = peers.subscribe((peerList) => {
					const found = peerList.find((p) => p.peer_id === data.from_peer_id);
					if (found) fromUsername = found.username;
				});
				unsub();
				handleRemoteOffer(data.from_peer_id, data.sdp, fromUsername);
			} catch (e) {
				console.error('Failed to parse video call offer:', e);
			}
		});

		this.eventSource.addEventListener('video_call_answer', (event) => {
			try {
				const data = JSON.parse(event.data);
				handleRemoteAnswer(data.sdp);
			} catch (e) {
				console.error('Failed to parse video call answer:', e);
			}
		});

		this.eventSource.addEventListener('video_call_ice_candidate', (event) => {
			try {
				const data = JSON.parse(event.data);
				handleRemoteICECandidate(data.candidate, data.sdp_mid, data.sdp_mline_index);
			} catch (e) {
				console.error('Failed to parse video call ICE candidate:', e);
			}
		});

		this.eventSource.addEventListener('video_call_hangup', (event) => {
			try {
				JSON.parse(event.data); // validate
				handleRemoteHangup();
			} catch (e) {
				console.error('Failed to parse video call hangup:', e);
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
