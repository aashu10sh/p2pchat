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

	const localUnsub = localStream.subscribe((stream) => {
		if (localVideoEl && stream) {
			localVideoEl.srcObject = stream;
		}
	});

	const remoteUnsub = remoteStream.subscribe((stream) => {
		if (remoteVideoEl && stream) {
			remoteVideoEl.srcObject = stream;
		}
	});

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
				return 'Calling...';
			case 'ringing':
				return 'Ringing...';
			case 'connected':
				return 'Connected';
			case 'ended':
				return 'Call ended';
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
		<!-- Top Status Bar -->
		<div class="status-bar">
			<div class="status-left">
				<span class="status-dot" class:connected={currentState === 'connected'} class:calling={currentState === 'calling' || currentState === 'ringing'}></span>
				<span class="status-text" class:connected={currentState === 'connected'}>
					{getStatusLabel(currentState)}
				</span>
				{#if currentState === 'connected'}
					<span class="duration">{formatDuration(callDuration)}</span>
				{/if}
			</div>
			<span class="peer-name">{peerName}</span>
		</div>

		<!-- Video Area -->
		<div class="video-area">
			{#if currentState === 'calling' || currentState === 'ringing'}
				<div class="connecting-state">
					<div class="connecting-avatar">
						<span>{peerName ? peerName.charAt(0).toUpperCase() : '?'}</span>
						<div class="ring-animation"></div>
					</div>
					<p class="connecting-label">{getStatusLabel(currentState)}</p>
					<p class="connecting-name">{peerName}</p>
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

		<!-- Controls -->
		<div class="controls">
			{#if currentState !== 'ended'}
				<button class="control-btn hangup" onclick={handleHangUp} title="End call">
					<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
						<path d="M12 9c-1.6 0-3.15.25-4.6.72v3.1c0 .39-.23.74-.56.9-.98.49-1.87 1.12-2.66 1.85-.18.18-.43.28-.7.28-.28 0-.53-.11-.71-.29L.29 13.08c-.18-.17-.29-.42-.29-.7 0-.28.11-.53.29-.71C3.34 8.78 7.46 7 12 7s8.66 1.78 11.71 4.67c.18.18.29.43.29.71 0 .28-.11.53-.29.71l-2.48 2.48c-.18.18-.43.29-.71.29-.27 0-.52-.11-.7-.28-.79-.74-1.69-1.36-2.67-1.85-.33-.16-.56-.5-.56-.9v-3.1C15.15 9.25 13.6 9 12 9z"/>
					</svg>
				</button>
			{:else}
				<div class="ended-label">Call ended</div>
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
		background: #000000;
		z-index: 900;
		display: flex;
		flex-direction: column;
		animation: fadeIn 0.3s ease-out;
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	/* ── Status Bar ── */
	.status-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 24px;
		background: rgba(255, 255, 255, 0.06);
		backdrop-filter: blur(20px);
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
	}

	.status-left {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--text-secondary);
	}

	.status-dot.connected {
		background: var(--success);
		box-shadow: 0 0 8px rgba(52, 199, 89, 0.5);
	}

	.status-dot.calling {
		background: var(--warning);
		animation: pulse 1.5s infinite ease-in-out;
	}

	.status-text {
		font-size: 14px;
		font-weight: 500;
		color: rgba(255, 255, 255, 0.7);
	}

	.status-text.connected {
		color: var(--success);
	}

	.duration {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.5);
		font-variant-numeric: tabular-nums;
		font-family: var(--font-mono);
	}

	.peer-name {
		font-size: 14px;
		font-weight: 500;
		color: rgba(255, 255, 255, 0.8);
	}

	/* ── Video Area ── */
	.video-area {
		flex: 1;
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}

	.connecting-state {
		position: absolute;
		z-index: 2;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 16px;
	}

	.connecting-avatar {
		width: 80px;
		height: 80px;
		border-radius: 50%;
		background: linear-gradient(135deg, var(--accent), #5856d6);
		display: flex;
		align-items: center;
		justify-content: center;
		position: relative;
	}

	.connecting-avatar span {
		font-size: 32px;
		font-weight: 600;
		color: white;
	}

	.ring-animation {
		position: absolute;
		inset: -8px;
		border: 2px solid rgba(255, 255, 255, 0.2);
		border-radius: 50%;
		animation: ringExpand 2s infinite ease-out;
	}

	@keyframes ringExpand {
		0% {
			inset: -4px;
			opacity: 1;
		}
		100% {
			inset: -24px;
			opacity: 0;
		}
	}

	.connecting-label {
		font-size: 16px;
		font-weight: 500;
		color: rgba(255, 255, 255, 0.8);
		margin: 0;
	}

	.connecting-name {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.5);
		margin: 0;
	}

	.remote-video {
		width: 100%;
		height: 100%;
		object-fit: contain;
		opacity: 0;
		transition: opacity 0.4s ease;
	}

	.remote-video.visible {
		opacity: 1;
	}

	.local-video {
		position: absolute;
		bottom: 24px;
		right: 24px;
		width: 180px;
		height: 135px;
		object-fit: cover;
		border-radius: var(--radius-lg);
		border: 2px solid rgba(255, 255, 255, 0.1);
		background: #111;
		z-index: 3;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
	}

	/* ── Controls ── */
	.controls {
		padding: 24px;
		display: flex;
		justify-content: center;
		background: rgba(255, 255, 255, 0.04);
		backdrop-filter: blur(20px);
	}

	.control-btn {
		width: 56px;
		height: 56px;
		border-radius: 50%;
		border: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all var(--transition-fast);
	}

	.control-btn.hangup {
		background: var(--danger);
		color: white;
	}

	.control-btn.hangup:hover {
		background: var(--danger-hover);
		transform: scale(1.08);
	}

	.control-btn.hangup:active {
		transform: scale(0.95);
	}

	.ended-label {
		font-size: 15px;
		font-weight: 500;
		color: rgba(255, 255, 255, 0.5);
		padding: 16px 0;
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 0.5;
		}
		50% {
			opacity: 1;
		}
	}
</style>
