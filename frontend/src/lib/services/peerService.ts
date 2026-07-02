import type { Peer } from '$lib/entites/peer';
import type { Error } from '$lib/entites/error';
import type { Message } from '$lib/entites/message';
import { err, ok, Result } from 'neverthrow';
import { writable } from 'svelte/store';
import { messages } from './chatService';
import { sortMessagesByTimestamp } from '$lib/utils';
import { fileTransfers } from './fileService';
import type { FileTransfer } from '$lib/entites/fileTransfer';
import { callHistories } from './callHistoryService';
import type { CallHistory } from '$lib/entites/callHistory';
import {
	handleRemoteOffer,
	handleRemoteAnswer,
	handleRemoteICECandidate,
	handleRemoteHangup
} from './callService';

export const peers = writable<Peer[]>([]);

/** Maximum number of SSE reconnect attempts before giving up */
const MAX_RECONNECT_ATTEMPTS = 10;
const BASE_RECONNECT_DELAY_MS = 1000;

export default class PeerService {
	private eventSource: EventSource | null = null;
	private messageCallback: ((message: Message) => void) | null = null;
	private reconnectAttempts = 0;
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

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
		this.connectSSE(onPeersUpdate);
	}

	private connectSSE(onPeersUpdate?: (peers: Peer[]) => void) {
		this.eventSource = new EventSource('http://localhost:8000/api/events');

		this.eventSource.addEventListener('connected', (event) => {
			console.log('SSE Connected:', event.data);
			this.reconnectAttempts = 0; // Reset on successful connection
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

				// Update messages store with sorting to maintain order
				messages.update((msgs) => sortMessagesByTimestamp([...msgs, message]));

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

				// Update messages store with sorting to maintain order
				messages.update((msgs) => sortMessagesByTimestamp([...msgs, message]));
			} catch (e) {
				console.error('Failed to parse sent message:', e);
			}
		});

		this.eventSource.addEventListener('file_received', (event) => {
			try {
				const ft = JSON.parse(event.data) as FileTransfer;
				fileTransfers.update((fts) => [ft, ...fts]);
			} catch (e) {
				console.error('Failed to parse file received event:', e);
			}
		});

		this.eventSource.addEventListener('file_sent', (event) => {
			try {
				const ft = JSON.parse(event.data) as FileTransfer;
				fileTransfers.update((fts) => [ft, ...fts]);
			} catch (e) {
				console.error('Failed to parse file sent event:', e);
			}
		});

		this.eventSource.addEventListener('call_history_saved', (event) => {
			try {
				const ch = JSON.parse(event.data) as CallHistory;
				callHistories.update((chs) => [ch, ...chs]);
			} catch (e) {
				console.error('Failed to parse call history event:', e);
			}
		});

		// Video call signaling events
		this.eventSource.addEventListener('video_call_offer', (event) => {
			try {
				const data = JSON.parse(event.data);
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

		this.eventSource.onerror = () => {
			console.warn('SSE connection error, attempting reconnect...');
			this.handleReconnect(onPeersUpdate);
		};
	}

	private handleReconnect(onPeersUpdate?: (peers: Peer[]) => void) {
		// Close the broken connection
		if (this.eventSource) {
			this.eventSource.close();
			this.eventSource = null;
		}

		if (this.reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
			console.error('SSE: Max reconnect attempts reached. Giving up.');
			return;
		}

		// Exponential backoff: 1s, 2s, 4s, 8s... capped at 30s
		const delay = Math.min(
			BASE_RECONNECT_DELAY_MS * Math.pow(2, this.reconnectAttempts),
			30000
		);
		this.reconnectAttempts++;

		console.log(`SSE: Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})`);

		this.reconnectTimer = setTimeout(() => {
			this.connectSSE(onPeersUpdate);
		}, delay);
	}

	stopEventStream() {
		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}
		if (this.eventSource) {
			this.eventSource.close();
			this.eventSource = null;
		}
		this.reconnectAttempts = 0;
	}
}
