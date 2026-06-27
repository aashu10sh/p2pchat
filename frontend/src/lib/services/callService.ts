import { writable, get } from 'svelte/store';

// Call states
export type CallState = 'idle' | 'calling' | 'ringing' | 'connected' | 'ended';

export interface IncomingCallData {
	fromPeerId: string;
	fromUsername: string;
	sdp: string;
}

export const callState = writable<CallState>('idle');
export const remotePeerId = writable<string>('');
export const remotePeerName = writable<string>('');
export const localStream = writable<MediaStream | null>(null);
export const remoteStream = writable<MediaStream | null>(null);
export const incomingCall = writable<IncomingCallData | null>(null);
export const mediaError = writable<string | null>(null);

let pc: RTCPeerConnection | null = null;

// Queues ICE candidates received before remote description is set
let pendingICECandidates: RTCIceCandidateInit[] = [];

// No STUN/TURN — LAN only, host candidates are sufficient
const rtcConfig: RTCConfiguration = {
	iceServers: [],
	iceTransportPolicy: 'all'
};

async function getLocalMedia(): Promise<MediaStream> {
	try {
		mediaError.set(null);
		const stream = await navigator.mediaDevices.getUserMedia({
			video: true,
			audio: true
		});
		localStream.set(stream);
		return stream;
	} catch (err) {
		const errorMessage =
			err instanceof DOMException
				? err.name === 'NotAllowedError'
					? 'Camera and microphone access denied. Please allow permissions.'
					: err.name === 'NotFoundError'
						? 'No camera or microphone found on this device.'
						: `Media error: ${err.message}`
				: 'Unable to access camera and microphone.';

		mediaError.set(errorMessage);
		console.error('getLocalMedia failed:', err);
		throw err;
	}
}

function createPeerConnection(): RTCPeerConnection {
	pc = new RTCPeerConnection(rtcConfig);

	pc.ontrack = (event) => {
		remoteStream.set(event.streams[0]);
	};

	pc.onicecandidate = (event) => {
		if (!event.candidate) return;

		// Only forward host candidates (LAN-only, no relay/srflx)
		if (event.candidate.type && event.candidate.type !== 'host') return;

		const toPeerId = get(remotePeerId);
		if (!toPeerId) return;

		fetch('http://localhost:8000/api/call/ice-candidate', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				to_peer_id: toPeerId,
				candidate: event.candidate.candidate,
				sdp_mid: event.candidate.sdpMid,
				sdp_mline_index: event.candidate.sdpMLineIndex
			})
		}).catch((err) => console.error('Failed to send ICE candidate:', err));
	};

	pc.oniceconnectionstatechange = () => {
		if (!pc) return;
		switch (pc.iceConnectionState) {
			case 'connected':
			case 'completed':
				callState.set('connected');
				break;
			case 'disconnected':
			case 'failed':
			case 'closed':
				// Remote peer dropped unexpectedly
				cleanup();
				callState.set('ended');
				setTimeout(() => callState.set('idle'), 3000);
				break;
		}
	};

	return pc;
}

// Caller initiates a call
export async function startCall(peerId: string, peerName: string): Promise<void> {
	if (get(callState) !== 'idle') return;

	remotePeerId.set(peerId);
	remotePeerName.set(peerName);
	callState.set('calling');

	try {
		const stream = await getLocalMedia();
		createPeerConnection();

		stream.getTracks().forEach((track) => pc!.addTrack(track, stream));

		const offer = await pc!.createOffer();
		await pc!.setLocalDescription(offer);

		const response = await fetch('http://localhost:8000/api/call/offer', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				to_peer_id: peerId,
				sdp: offer.sdp
			})
		});

		if (!response.ok) {
			throw new Error('Failed to send offer');
		}
	} catch (err) {
		console.error('Failed to start call:', err);
		cleanup();
		callState.set('ended');
		setTimeout(() => callState.set('idle'), 3000);
	}
}

// Callee accepts an incoming call
export async function acceptCall(): Promise<void> {
	const incoming = get(incomingCall);
	if (!incoming || !pc) return;

	callState.set('connected');

	try {
		const answer = await pc.createAnswer();
		await pc.setLocalDescription(answer);

		await fetch('http://localhost:8000/api/call/answer', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				to_peer_id: incoming.fromPeerId,
				sdp: answer.sdp
			})
		});

		// Drain any ICE candidates that arrived before we set remote description
		for (const candidate of pendingICECandidates) {
			await pc.addIceCandidate(candidate);
		}
		pendingICECandidates = [];

		incomingCall.set(null);
	} catch (err) {
		console.error('Failed to accept call:', err);
		cleanup();
		callState.set('ended');
		setTimeout(() => callState.set('idle'), 3000);
	}
}

// Callee rejects an incoming call
export async function rejectCall(): Promise<void> {
	const incoming = get(incomingCall);
	if (!incoming) return;

	await fetch('http://localhost:8000/api/call/hangup', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			to_peer_id: incoming.fromPeerId,
			reason: 'rejected'
		})
	}).catch((err) => console.error('Failed to send reject:', err));

	incomingCall.set(null);
	cleanup();
	callState.set('idle');
}

// Either side hangs up
export async function hangUp(): Promise<void> {
	const peerId = get(remotePeerId);
	if (peerId) {
		await fetch('http://localhost:8000/api/call/hangup', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				to_peer_id: peerId,
				reason: 'hangup'
			})
		}).catch((err) => console.error('Failed to send hangup:', err));
	}

	cleanup();
	callState.set('ended');
	setTimeout(() => callState.set('idle'), 3000);
}

// SSE event handlers — called from peerService when signaling events arrive

export async function handleRemoteOffer(
	fromPeerId: string,
	sdp: string,
	fromUsername: string
): Promise<void> {
	// If already in a call, auto-reject the new one
	if (get(callState) !== 'idle') {
		await fetch('http://localhost:8000/api/call/hangup', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				to_peer_id: fromPeerId,
				reason: 'busy'
			})
		}).catch(() => {});
		return;
	}

	remotePeerId.set(fromPeerId);
	remotePeerName.set(fromUsername);
	callState.set('ringing');

	try {
		const stream = await getLocalMedia();
		createPeerConnection();

		stream.getTracks().forEach((track) => pc!.addTrack(track, stream));

		await pc!.setRemoteDescription(
			new RTCSessionDescription({
				type: 'offer',
				sdp: sdp
			})
		);

		incomingCall.set({ fromPeerId, fromUsername, sdp });
	} catch (err) {
		console.error('Failed to handle remote offer:', err);
		cleanup();
		callState.set('idle');
	}
}

export async function handleRemoteAnswer(sdp: string): Promise<void> {
	if (!pc) return;

	try {
		await pc.setRemoteDescription(
			new RTCSessionDescription({
				type: 'answer',
				sdp: sdp
			})
		);

		// Drain any ICE candidates that arrived before remote description was set
		for (const candidate of pendingICECandidates) {
			await pc.addIceCandidate(candidate);
		}
		pendingICECandidates = [];

		callState.set('connected');
	} catch (err) {
		console.error('Failed to handle remote answer:', err);
	}
}

export async function handleRemoteICECandidate(
	candidate: string,
	sdpMid: string,
	sdpMLineIndex: number
): Promise<void> {
	if (!pc) return;

	const iceCandidate: RTCIceCandidateInit = {
		candidate: candidate,
		sdpMid: sdpMid,
		sdpMLineIndex: sdpMLineIndex
	};

	// Queue candidates if remote description not yet set
	if (!pc.remoteDescription) {
		pendingICECandidates.push(iceCandidate);
		return;
	}

	try {
		await pc.addIceCandidate(iceCandidate);
	} catch (err) {
		console.error('Failed to add ICE candidate:', err);
	}
}

export function handleRemoteHangup(): void {
	incomingCall.set(null);
	cleanup();
	callState.set('ended');
	setTimeout(() => callState.set('idle'), 3000);
}

function cleanup(): void {
	if (pc) {
		pc.close();
		pc = null;
	}

	const stream = get(localStream);
	if (stream) {
		stream.getTracks().forEach((track) => track.stop());
		localStream.set(null);
	}

	remoteStream.set(null);
	mediaError.set(null);
	pendingICECandidates = [];
}
