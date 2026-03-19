<script lang="ts">
	import { onDestroy } from 'svelte';
	import {
		callState,
		localStream,
		remoteStream,
		remotePeerName,
		hangUp
	} from '$lib/services/callService';

	let localVideoEl = $state<HTMLVideoElement | null>(null);
	let remoteVideoEl = $state<HTMLVideoElement | null>(null);
	let currentState = $state('idle');
	let peerName = $state('');
	let callDuration = $state(0);
	let durationInterval: ReturnType<typeof setInterval> | null = null;

	const stateUnsub = callState.subscribe((val) => {
		currentState = val;
		if (val === 'connected' && !durationInterval) {
			callDuration = 0;
			durationInterval = setInterval(() => {
				callDuration++;
			}, 1000);
		}
		if (val === 'idle' || val === 'ended') {
			if (durationInterval) {
				clearInterval(durationInterval);
				durationInterval = null;
			}
		}
	});

	const nameUnsub = remotePeerName.subscribe((val) => {
		peerName = val;
	});

	// Bind local stream to video element
	const localUnsub = localStream.subscribe((stream) => {
		if (localVideoEl && stream) {
			localVideoEl.srcObject = stream;
		}
	});

	// Bind remote stream to video element
	const remoteUnsub = remoteStream.subscribe((stream) => {
		if (remoteVideoEl && stream) {
			remoteVideoEl.srcObject = stream;
		}
	});

	// Re-bind when video elements mount
	$effect(() => {
		if (localVideoEl) {
			const unsub = localStream.subscribe((stream) => {
				if (localVideoEl && stream) {
					localVideoEl.srcObject = stream;
				}
			});
			return unsub;
		}
	});

	$effect(() => {
		if (remoteVideoEl) {
			const unsub = remoteStream.subscribe((stream) => {
				if (remoteVideoEl && stream) {
					remoteVideoEl.srcObject = stream;
				}
			});
			return unsub;
		}
	});

	onDestroy(() => {
		stateUnsub();
		nameUnsub();
		localUnsub();
		remoteUnsub();
		if (durationInterval) clearInterval(durationInterval);
	});

	function formatDuration(seconds: number): string {
		const m = Math.floor(seconds / 60)
			.toString()
			.padStart(2, '0');
		const s = (seconds % 60).toString().padStart(2, '0');
		return `${m}:${s}`;
	}

	function getStatusLabel(state: string): string {
		switch (state) {
			case 'calling':
				return 'INITIATING_LINK';
			case 'ringing':
				return 'AWAITING_RESPONSE';
			case 'connected':
				return 'LINK_ACTIVE';
			case 'ended':
				return 'LINK_TERMINATED';
			default:
				return '';
		}
	}

	async function handleHangUp() {
		await hangUp();
	}
</script>

{#if currentState !== 'idle'}
	<div class="call-overlay">
		<div class="call-status-bar">
			<span class="status-label" class:connected={currentState === 'connected'}>
				[{getStatusLabel(currentState)}]
			</span>
			{#if currentState === 'connected'}
				<span class="duration">{formatDuration(callDuration)}</span>
			{/if}
			<span class="peer-label">TARGET: @{peerName}</span>
		</div>

		<div class="video-container">
			{#if currentState === 'calling' || currentState === 'ringing'}
				<div class="connecting-indicator">
					<div class="connecting-text pulse">{getStatusLabel(currentState)}...</div>
					<div class="connecting-sub">@{peerName}</div>
				</div>
			{/if}

			<!-- svelte-ignore a11y_media_has_caption -->
			<video
				class="remote-video"
				bind:this={remoteVideoEl}
				autoplay
				playsinline
				class:visible={currentState === 'connected'}
			></video>

			<!-- svelte-ignore a11y_media_has_caption -->
			<video
				class="local-video"
				bind:this={localVideoEl}
				autoplay
				playsinline
				muted
			></video>
		</div>

		<div class="call-controls">
			{#if currentState !== 'ended'}
				<button class="hangup-btn" onclick={handleHangUp}>
					[TERMINATE]
				</button>
			{:else}
				<div class="ended-text">LINK_CLOSED</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	.call-overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		background: var(--bg-primary);
		z-index: 900;
		display: flex;
		flex-direction: column;
		font-family: var(--font-mono);
		animation: fadeIn 0.2s ease-out;
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	.call-status-bar {
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 16px 24px;
		border-bottom: 1px solid var(--border-color);
		background: var(--bg-secondary);
		font-size: 11px;
		letter-spacing: 0.1em;
	}

	.status-label {
		color: var(--text-secondary);
	}

	.status-label.connected {
		color: var(--text-accent);
	}

	.duration {
		color: var(--text-primary);
		font-variant-numeric: tabular-nums;
	}

	.peer-label {
		color: var(--text-secondary);
		margin-left: auto;
	}

	.video-container {
		flex: 1;
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		background: #050505;
	}

	.connecting-indicator {
		text-align: center;
		position: absolute;
		z-index: 2;
	}

	.connecting-text {
		color: var(--text-accent);
		font-size: 14px;
		letter-spacing: 0.1em;
		margin-bottom: 8px;
	}

	.connecting-sub {
		color: var(--text-secondary);
		font-size: 12px;
	}

	.pulse {
		animation: pulse 1.5s infinite ease-in-out;
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 0.4;
		}
		50% {
			opacity: 1;
		}
	}

	.remote-video {
		width: 100%;
		height: 100%;
		object-fit: contain;
		background: #050505;
		opacity: 0;
		transition: opacity 0.3s ease;
	}

	.remote-video.visible {
		opacity: 1;
	}

	.local-video {
		position: absolute;
		bottom: 24px;
		right: 24px;
		width: 200px;
		height: 150px;
		object-fit: cover;
		border: 1px solid var(--border-color);
		background: #111;
		z-index: 3;
	}

	.call-controls {
		padding: 20px 24px;
		display: flex;
		justify-content: center;
		border-top: 1px solid var(--border-color);
		background: var(--bg-secondary);
	}

	.hangup-btn {
		font-family: var(--font-mono);
		font-size: 12px;
		letter-spacing: 0.1em;
		padding: 12px 32px;
		border: 1px solid #ff3333;
		color: #ff3333;
		background: transparent;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.hangup-btn:hover {
		background: #ff3333;
		color: var(--bg-primary);
	}

	.ended-text {
		color: var(--text-secondary);
		font-size: 12px;
		letter-spacing: 0.1em;
	}
</style>
