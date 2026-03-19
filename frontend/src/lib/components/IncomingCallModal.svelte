<script lang="ts">
	import {
		incomingCall,
		callState,
		acceptCall,
		rejectCall,
		remotePeerName
	} from '$lib/services/callService';
	import { peers } from '$lib/services/peerService';

	let callerName = $state('');
	let visible = $state(false);

	$effect(() => {
		const incoming = incomingCallValue;

		if (incoming) {
			visible = true;
			// Try to resolve username from peers store
			const unsubscribe = peers.subscribe((peerList) => {
				const found = peerList.find((p) => p.peer_id === incoming.fromPeerId);
				if (found) {
					callerName = found.username;
				} else {
					callerName = incoming.fromPeerId.substring(0, 8) + '...';
				}
			});
			unsubscribe();
		} else {
			visible = false;
		}
	});

	// Local state bound to store
	import type { IncomingCallData } from '$lib/services/callService';
	let incomingCallValue = $state<IncomingCallData | null>(null);

	// Sync local state with store
	$effect(() => {
		const unsub = incomingCall.subscribe((val) => {
			incomingCallValue = val;
		});
		return unsub;
	});

	async function handleAccept() {
		await acceptCall();
	}

	async function handleReject() {
		await rejectCall();
	}
</script>

{#if incomingCallValue}
	<div class="modal-overlay">
		<div class="modal-container">
			<div class="modal-header">
				<span class="label">INCOMING_CALL</span>
			</div>

			<div class="caller-info">
				<div class="ring-indicator">
					<div class="ring-dot"></div>
				</div>
				<div class="caller-name">@{callerName}</div>
				<div class="caller-status">is requesting a video link...</div>
			</div>

			<div class="modal-actions">
				<button class="btn accept" onclick={handleAccept}>
					[ACCEPT]
				</button>
				<button class="btn reject" onclick={handleReject}>
					[REJECT]
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		background: rgba(0, 0, 0, 0.85);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
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

	.modal-container {
		border: 1px solid var(--text-accent);
		background: var(--bg-primary);
		padding: 32px 40px;
		min-width: 360px;
		display: flex;
		flex-direction: column;
		gap: 24px;
		box-shadow: 0 0 30px rgba(0, 255, 65, 0.1);
	}

	.modal-header .label {
		color: var(--text-accent);
		font-size: 11px;
		letter-spacing: 0.15em;
	}

	.caller-info {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		padding: 16px 0;
	}

	.ring-indicator {
		width: 48px;
		height: 48px;
		border: 2px solid var(--text-accent);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		animation: ringPulse 1.5s infinite ease-in-out;
	}

	.ring-dot {
		width: 12px;
		height: 12px;
		background: var(--text-accent);
		border-radius: 50%;
	}

	@keyframes ringPulse {
		0%,
		100% {
			box-shadow: 0 0 5px rgba(0, 255, 65, 0.3);
			transform: scale(1);
		}
		50% {
			box-shadow: 0 0 20px rgba(0, 255, 65, 0.6);
			transform: scale(1.05);
		}
	}

	.caller-name {
		color: var(--text-primary);
		font-size: 18px;
		font-weight: 500;
	}

	.caller-status {
		color: var(--text-secondary);
		font-size: 12px;
		letter-spacing: 0.05em;
	}

	.modal-actions {
		display: flex;
		gap: 16px;
		justify-content: center;
	}

	.btn {
		font-family: var(--font-mono);
		font-size: 12px;
		letter-spacing: 0.1em;
		padding: 10px 24px;
		border: 1px solid;
		cursor: pointer;
		background: transparent;
		transition: all 0.15s ease;
	}

	.btn.accept {
		color: var(--text-accent);
		border-color: var(--text-accent);
	}

	.btn.accept:hover {
		background: var(--text-accent);
		color: var(--bg-primary);
	}

	.btn.reject {
		color: #ff3333;
		border-color: #ff3333;
	}

	.btn.reject:hover {
		background: #ff3333;
		color: var(--bg-primary);
	}
</style>
