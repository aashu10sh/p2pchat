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

	import type { IncomingCallData } from '$lib/services/callService';
	let incomingCallValue = $state<IncomingCallData | null>(null);

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
	<div class="overlay">
		<div class="modal">
			<div class="caller-section">
				<div class="caller-avatar">
					<span>{callerName ? callerName.charAt(0).toUpperCase() : '?'}</span>
					<div class="pulse-ring"></div>
					<div class="pulse-ring delay"></div>
				</div>
				<h3 class="caller-name">{callerName}</h3>
				<p class="caller-sub">Incoming video call</p>
			</div>

			<div class="actions">
				<button class="action-btn decline" onclick={handleReject} title="Decline">
					<svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
						<path d="M12 9c-1.6 0-3.15.25-4.6.72v3.1c0 .39-.23.74-.56.9-.98.49-1.87 1.12-2.66 1.85-.18.18-.43.28-.7.28-.28 0-.53-.11-.71-.29L.29 13.08c-.18-.17-.29-.42-.29-.7 0-.28.11-.53.29-.71C3.34 8.78 7.46 7 12 7s8.66 1.78 11.71 4.67c.18.18.29.43.29.71 0 .28-.11.53-.29.71l-2.48 2.48c-.18.18-.43.29-.71.29-.27 0-.52-.11-.7-.28-.79-.74-1.69-1.36-2.67-1.85-.33-.16-.56-.5-.56-.9v-3.1C15.15 9.25 13.6 9 12 9z"/>
					</svg>
					<span>Decline</span>
				</button>
				<button class="action-btn accept" onclick={handleAccept} title="Accept">
					<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<polygon points="23 7 16 12 23 17 23 7"></polygon>
						<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
					</svg>
					<span>Accept</span>
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		background: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(var(--glass-blur-heavy));
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		animation: overlayIn 0.3s ease-out;
	}

	@keyframes overlayIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	.modal {
		background: var(--glass-bg-heavy);
		backdrop-filter: blur(var(--glass-blur-heavy));
		border: 1px solid var(--border-color);
		border-radius: var(--radius-2xl);
		padding: 40px 48px;
		min-width: 320px;
		display: flex;
		flex-direction: column;
		gap: 32px;
		box-shadow: var(--shadow-xl);
		animation: modalIn 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
	}

	@keyframes modalIn {
		from {
			opacity: 0;
			transform: scale(0.92) translateY(12px);
		}
		to {
			opacity: 1;
			transform: scale(1) translateY(0);
		}
	}

	.caller-section {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
	}

	.caller-avatar {
		width: 72px;
		height: 72px;
		border-radius: 50%;
		background: linear-gradient(135deg, var(--accent), #5856d6);
		display: flex;
		align-items: center;
		justify-content: center;
		position: relative;
		margin-bottom: 4px;
	}

	.caller-avatar span {
		font-size: 28px;
		font-weight: 600;
		color: white;
	}

	.pulse-ring {
		position: absolute;
		inset: -6px;
		border: 2px solid var(--accent);
		border-radius: 50%;
		animation: ringPulse 2s infinite ease-out;
	}

	.pulse-ring.delay {
		animation-delay: 0.7s;
	}

	@keyframes ringPulse {
		0% {
			inset: -4px;
			opacity: 0.6;
		}
		100% {
			inset: -20px;
			opacity: 0;
		}
	}

	.caller-name {
		font-size: 20px;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
		letter-spacing: -0.02em;
	}

	.caller-sub {
		font-size: 14px;
		color: var(--text-secondary);
		margin: 0;
	}

	.actions {
		display: flex;
		gap: 16px;
		justify-content: center;
	}

	.action-btn {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		padding: 16px 28px;
		border-radius: var(--radius-xl);
		border: none;
		cursor: pointer;
		font-family: var(--font-sans);
		font-size: 13px;
		font-weight: 500;
		transition: all var(--transition-fast);
	}

	.action-btn.accept {
		background: var(--success);
		color: white;
	}

	.action-btn.accept:hover {
		background: #2db84e;
		transform: scale(1.04);
	}

	.action-btn.decline {
		background: var(--danger);
		color: white;
	}

	.action-btn.decline:hover {
		background: var(--danger-hover);
		transform: scale(1.04);
	}

	.action-btn:active {
		transform: scale(0.96);
	}
</style>
