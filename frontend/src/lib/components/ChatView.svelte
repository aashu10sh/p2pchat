<script lang="ts">
	import { onMount } from 'svelte';
	import ChatService, { messages } from '$lib/services/chatService';
	import { peers } from '$lib/services/peerService';
	import { startCall } from '$lib/services/callService';
	import type { Message } from '$lib/entites/message';
	import type { Peer } from '$lib/entites/peer';

	interface Props {
		activePeer: Peer | null;
		myPeerId: string;
	}

	let { activePeer, myPeerId }: Props = $props();
	let messageList = $state<Message[]>([]);
	let messageInput = $state('');
	let isLoading = $state(false);
	let isSending = $state(false);
	let chatContainer = $state<HTMLDivElement | null>(null);
	let currentPeer = $state<Peer | null>(null);
	let inputRef = $state<HTMLInputElement | null>(null);
	let isFocused = $state(false);

	const chatService = new ChatService();

	// Update currentPeer reactively when peers store changes
	$effect(() => {
		if (!activePeer) {
			currentPeer = null;
			return;
		}

		currentPeer = activePeer; // Set initially

		const unsubscribe = peers.subscribe((peerList) => {
			const updated = peerList.find((p) => p.peer_id === activePeer.peer_id);
			if (updated) {
				currentPeer = updated;
			}
		});

		return () => {
			unsubscribe();
		};
	});

	$effect(() => {
		if (activePeer) {
			loadMessages();
		}
	});

	async function loadMessages() {
		if (!activePeer) return;

		isLoading = true;
		const result = await chatService.fetchMessages(activePeer.peer_id);
		result.match(
			(msgs) => {
				// Don't reverse - keep chronological order (oldest first, newest last)
				messageList = msgs;
				setTimeout(scrollToBottom, 100);
			},
			(error) => {
				console.error('Failed to load messages:', error);
			}
		);
		isLoading = false;
	}

	async function handleSendMessage(e: Event) {
		e.preventDefault();
		if (!messageInput.trim() || !activePeer || isSending) return;

		const content = messageInput;
		messageInput = '';
		isSending = true;

		const result = await chatService.sendMessage(activePeer.peer_id, content);
		result.match(
			() => {
				// Message sent successfully
				isSending = false;
				setTimeout(() => {
					if (inputRef) inputRef.focus();
				}, 10);
			},
			(error) => {
				console.error('Failed to send message:', error);
				messageInput = content; // Restore message on error
				isSending = false;
			}
		);
	}

	function handleCallPeer() {
		if (!currentPeer) return;
		startCall(currentPeer.peer_id, currentPeer.username);
	}

	function scrollToBottom() {
		if (chatContainer) {
			chatContainer.scrollTop = chatContainer.scrollHeight;
		}
	}

	function isRecentlyOnline(peer: Peer): boolean {
		if (!peer.last_seen) return false;
		const lastSeen = new Date(peer.last_seen);
		const now = new Date();
		const diffMinutes = (now.getTime() - lastSeen.getTime()) / 1000 / 60;
		return diffMinutes < 1; // Online if seen in last minute
	}

	function formatTime(timestamp: string): string {
		if (!timestamp) return '';
		try {
			const date = new Date(timestamp);
			if (isNaN(date.getTime())) {
				return '';
			}
			return date.toLocaleTimeString('en-US', {
				hour12: false,
				hour: '2-digit',
				minute: '2-digit',
				second: '2-digit'
			});
		} catch {
			return '';
		}
	}

	function formatDate(timestamp: string): string {
		if (!timestamp) return '';
		try {
			const date = new Date(timestamp);
			if (isNaN(date.getTime())) {
				return '';
			}
			return date.toISOString().split('T')[0];
		} catch {
			return '';
		}
	}

	onMount(() => {
		// Subscribe to messages store for real-time updates
		const unsubscribe = messages.subscribe((value) => {
			if (activePeer && value.length > 0) {
				// Don't reverse - keep chronological order
				messageList = value;
				setTimeout(scrollToBottom, 50);
			}
		});

		return () => {
			unsubscribe();
		};
	});

	function handleContainerClick() {
		if (inputRef) inputRef.focus();
	}
</script>

{#if !activePeer}
	<div class="no-chat-selected">
		<div class="placeholder">
			<div class="cursor pulse">_</div>
			<h3>AWAITING_CONNECTION</h3>
			<p>Select a node from the registry to initiate transmission sequence.</p>
		</div>
	</div>
{:else}
	<div class="chat-view">
		<div class="chat-header">
			<div class="peer-details">
				<span class="label">TARGET:</span>
				<h2>@{currentPeer?.username}</h2>
				<span class="peer-status" class:online={currentPeer && isRecentlyOnline(currentPeer)}>
					[{currentPeer && isRecentlyOnline(currentPeer) ? 'LINK_ACTIVE' : 'OFFLINE'}]
				</span>
			</div>
			<div class="header-meta">
				<span class="label">ID:</span>
				{currentPeer?.peer_id.substring(0, 8)}...
				{#if currentPeer && isRecentlyOnline(currentPeer)}
					<button class="call-btn" onclick={handleCallPeer}>[CALL]</button>
				{/if}
			</div>
		</div>

		<div class="messages-container" bind:this={chatContainer}>
			{#if isLoading}
				<div class="loading">
					<span class="pulse">Decrypting transmission logs...</span>
				</div>
			{:else if messageList.length === 0}
				<div class="no-messages">
					<p>>> CONNECTION ESTABLISHED. LOG IS EMPTY.</p>
				</div>
			{:else}
				<div class="log-start">--- BEGIN_LOG ---</div>
				{#each messageList as message, index (message.ID)}
					{#if index === 0 || formatDate(messageList[index - 1].sent_at) !== formatDate(message.sent_at)}
						<div class="date-separator">
							<span>{formatDate(message.sent_at)}</span>
						</div>
					{/if}
					<div
						class="message"
						class:sent={message.is_sent_by_me}
						class:received={!message.is_sent_by_me}
					>
						<div class="message-meta">
							<span class="message-time">[{formatTime(message.sent_at)}]</span>
							<span class="message-author">{message.is_sent_by_me ? 'SYS' : 'RCV'}</span>
							{#if message.is_sent_by_me}
								<span class="message-status">
									{#if message.delivered_at}
										<span class="ack">[ACK]</span>
									{:else}
										<span class="pend">[...]</span>
									{/if}
								</span>
							{/if}
						</div>
						<div class="message-content">
							<span class="prompt-arrow">{message.is_sent_by_me ? '>' : '<'}</span>
							<p class="message-text">{message.content}</p>
						</div>
					</div>
				{/each}
				<div class="log-end">--- END_LOG ---</div>
			{/if}
		</div>

		<form class="message-input-container" onsubmit={handleSendMessage}>
			<button type="button" class="input-wrapper-btn" onclick={handleContainerClick} tabindex="-1">
				<div class="input-wrapper">
					<span class="prompt-symbol {isFocused ? 'active' : ''}">>_</span>
					<input
						bind:this={inputRef}
						bind:value={messageInput}
						type="text"
						placeholder="Transmit packet..."
						class="message-input"
						autocomplete="off"
						spellcheck="false"
						disabled={isSending}
						onfocus={() => (isFocused = true)}
						onblur={() => (isFocused = false)}
					/>
				</div>
			</button>
			{#if isSending}
				<div class="sending-indicator pulse">[TX...]</div>
			{/if}
		</form>
	</div>
{/if}

<style>
	.no-chat-selected {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		background-color: var(--bg-primary);
		font-family: var(--font-mono);
	}

	.placeholder {
		text-align: center;
		color: var(--text-secondary);
	}

	.placeholder .cursor {
		font-size: 3rem;
		color: var(--text-accent);
		margin-bottom: 20px;
	}

	.placeholder h3 {
		color: var(--text-primary);
		font-size: 16px;
		letter-spacing: 0.1em;
		margin-bottom: 12px;
	}

	.placeholder p {
		font-size: 12px;
		opacity: 0.7;
		letter-spacing: 0.05em;
	}

	.chat-view {
		flex: 1;
		display: flex;
		flex-direction: column;
		background-color: var(--bg-primary);
		font-family: var(--font-mono);
	}

	.chat-header {
		padding: 16px 24px;
		border-bottom: 1px solid var(--border-color);
		background-color: var(--bg-secondary);
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.peer-details {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.label {
		color: var(--text-secondary);
		font-size: 10px;
		letter-spacing: 0.1em;
	}

	.peer-details h2 {
		color: var(--text-primary);
		font-size: 14px;
		margin: 0;
		font-weight: 500;
		letter-spacing: 0.05em;
	}

	.peer-status {
		font-size: 10px;
		color: var(--text-secondary);
		letter-spacing: 0.05em;
	}

	.peer-status.online {
		color: var(--text-accent);
	}

	.header-meta {
		font-size: 10px;
		color: var(--text-secondary);
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.call-btn {
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.1em;
		padding: 4px 12px;
		border: 1px solid var(--text-accent);
		color: var(--text-accent);
		background: transparent;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.call-btn:hover {
		background: var(--text-accent);
		color: var(--bg-primary);
	}

	.messages-container {
		flex: 1;
		overflow-y: auto;
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.loading,
	.no-messages {
		padding: 32px;
		color: var(--text-secondary);
		font-size: 12px;
		letter-spacing: 0.05em;
	}

	.log-start,
	.log-end {
		color: var(--border-hover);
		font-size: 10px;
		text-align: center;
		margin: 16px 0;
		letter-spacing: 0.1em;
	}

	.date-separator {
		display: flex;
		margin: 24px 0 16px 0;
	}

	.date-separator span {
		color: var(--text-secondary);
		font-size: 10px;
		letter-spacing: 0.1em;
		border-bottom: 1px solid var(--border-hover);
		padding-bottom: 2px;
	}

	.message {
		display: flex;
		flex-direction: column;
		margin-bottom: 12px;
		animation: slideIn 0.15s ease-out;
	}

	@keyframes slideIn {
		from {
			opacity: 0;
			transform: translateY(4px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.message-meta {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 10px;
		color: var(--text-secondary);
		margin-bottom: 4px;
	}

	.message.sent .message-meta {
		color: var(--text-accent);
		opacity: 0.7;
	}

	.message-author {
		font-weight: bold;
	}

	.message-status {
		margin-left: 4px;
	}

	.message-status .ack {
		color: var(--text-accent);
	}

	.message-content {
		display: flex;
		align-items: flex-start;
		gap: 8px;
	}

	.prompt-arrow {
		color: var(--text-secondary);
		margin-top: 1px;
	}

	.message.sent .prompt-arrow {
		color: var(--text-accent);
	}

	.message-text {
		margin: 0;
		word-wrap: break-word;
		font-size: 13px;
		line-height: 1.5;
		color: var(--text-primary);
		font-family: var(--font-sans);
	}

	.message.received .message-text {
		color: #e0e0e0;
	}

	.message-input-container {
		padding: 20px 24px;
		background-color: var(--bg-primary);
		border-top: 1px solid var(--border-color);
		display: flex;
		align-items: center;
		gap: 16px;
		cursor: text;
	}

	.input-wrapper {
		flex: 1;
		display: flex;
		align-items: center;
	}

	.prompt-symbol {
		color: var(--text-secondary);
		margin-right: 12px;
		font-weight: bold;
		font-size: 14px;
		transition: color 0.15s ease;
	}

	.prompt-symbol.active {
		color: var(--text-accent);
	}

	.message-input {
		flex: 1;
		background: transparent;
		border: none;
		color: var(--text-primary);
		font-family: var(--font-mono);
		font-size: 13px;
		outline: none;
		letter-spacing: 0.05em;
	}

	.message-input:disabled {
		opacity: 0.5;
	}

	.sending-indicator {
		color: var(--text-accent);
		font-size: 11px;
		letter-spacing: 0.1em;
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

	/* Scrollbar styling for chat container */
	.messages-container::-webkit-scrollbar {
		width: 6px;
	}

	.messages-container::-webkit-scrollbar-track {
		background: transparent;
	}

	.messages-container::-webkit-scrollbar-thumb {
		background: var(--border-color);
	}

	.messages-container::-webkit-scrollbar-thumb:hover {
		background: var(--border-hover);
	}
</style>
