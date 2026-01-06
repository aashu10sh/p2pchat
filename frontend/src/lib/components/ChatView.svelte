<script lang="ts">
	import { onMount } from 'svelte';
	import ChatService, { messages } from '$lib/services/chatService';
	import { peers } from '$lib/services/peerService';
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
	let chatContainer: HTMLDivElement;
	let currentPeer = $state<Peer | null>(activePeer);

	const chatService = new ChatService();

	// Update currentPeer reactively when peers store changes
	$effect(() => {
		if (activePeer) {
			peers.subscribe((peerList) => {
				const updated = peerList.find(p => p.peer_id === activePeer.peer_id);
				if (updated) {
					currentPeer = updated;
				}
			});
		} else {
			currentPeer = null;
		}
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
			},
			(error) => {
				console.error('Failed to send message:', error);
				messageInput = content; // Restore message on error
				isSending = false;
			}
		);
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
			return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
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
			const today = new Date();
			const messageDate = new Date(date);

			if (messageDate.toDateString() === today.toDateString()) {
				return 'Today';
			}

			const yesterday = new Date(today);
			yesterday.setDate(yesterday.getDate() - 1);
			if (messageDate.toDateString() === yesterday.toDateString()) {
				return 'Yesterday';
			}

			return messageDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
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
</script>

{#if !activePeer}
	<div class="no-chat-selected">
		<div class="placeholder">
			<svg width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="currentColor">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
				/>
			</svg>
			<h3>Select a peer to start chatting</h3>
			<p>Choose a peer from the sidebar to begin messaging</p>
		</div>
	</div>
{:else}
	<div class="chat-view">
		<div class="chat-header">
			<div class="peer-info">
				<div class="peer-avatar">
					{#if currentPeer?.image_url}
						<img src={currentPeer.image_url} alt={currentPeer.username} />
					{:else if currentPeer}
						<div class="avatar-placeholder">
							{currentPeer.username.charAt(0).toUpperCase()}
						</div>
					{/if}
				</div>
				<div class="peer-details">
					<h2>{currentPeer?.username}</h2>
					<span class="peer-status" class:online={currentPeer && isRecentlyOnline(currentPeer)}>
						{currentPeer && isRecentlyOnline(currentPeer) ? 'Online' : 'Offline'}
					</span>
				</div>
			</div>
		</div>

		<div class="messages-container" bind:this={chatContainer}>
			{#if isLoading}
				<div class="loading">Loading messages...</div>
			{:else if messageList.length === 0}
				<div class="no-messages">
					<p>No messages yet. Start the conversation!</p>
				</div>
			{:else}
				{#each messageList as message, index (message.ID)}
					{#if index === 0 || formatDate(messageList[index - 1].sent_at) !== formatDate(message.sent_at)}
						<div class="date-separator">
							<span>{formatDate(message.sent_at)}</span>
						</div>
					{/if}
					<div class="message" class:sent={message.is_sent_by_me} class:received={!message.is_sent_by_me}>
						<div class="message-bubble">
							<p class="message-text">{message.content}</p>
							<div class="message-footer">
								<span class="message-time">{formatTime(message.sent_at)}</span>
								{#if message.is_sent_by_me}
									<span class="message-status">
										{#if message.delivered_at}
											<svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
												<path d="M13.854 3.646a.5.5 0 0 1 0 .708l-7 7a.5.5 0 0 1-.708 0l-3.5-3.5a.5.5 0 1 1 .708-.708L6.5 10.293l6.646-6.647a.5.5 0 0 1 .708 0z"/>
												<path d="M10.354 3.646a.5.5 0 0 1 0 .708l-7 7a.5.5 0 0 1-.708 0l-1-1a.5.5 0 0 1 .708-.708l.646.647 6.646-6.647a.5.5 0 0 1 .708 0z"/>
											</svg>
										{:else}
											<svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
												<path d="M10.97 4.97a.75.75 0 0 1 1.07 1.05l-3.99 4.99a.75.75 0 0 1-1.08.02L4.324 8.384a.75.75 0 1 1 1.06-1.06l2.094 2.093 3.473-4.425a.267.267 0 0 1 .02-.022z"/>
											</svg>
										{/if}
									</span>
								{/if}
							</div>
						</div>
					</div>
				{/each}
			{/if}
		</div>

		<form class="message-input-container" onsubmit={handleSendMessage}>
			<input
				bind:value={messageInput}
				type="text"
				placeholder="Type a message..."
				class="message-input"
				autocomplete="off"
				disabled={isSending}
			/>
			<button type="submit" class="send-button" disabled={!messageInput.trim() || isSending}>
				{#if isSending}
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" class="spinner">
						<circle cx="12" cy="12" r="10" stroke-width="3"/>
					</svg>
				{:else}
					<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
						<path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
					</svg>
				{/if}
			</button>
		</form>
	</div>
{/if}

<style>
	.no-chat-selected {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		background-color: #36393f;
	}

	.placeholder {
		text-align: center;
		color: #72767d;
	}

	.placeholder svg {
		margin-bottom: 16px;
	}

	.placeholder h3 {
		color: #dcddde;
		font-size: 20px;
		margin-bottom: 8px;
	}

	.placeholder p {
		font-size: 14px;
	}

	.chat-view {
		flex: 1;
		display: flex;
		flex-direction: column;
		background-color: #36393f;
	}

	.chat-header {
		padding: 16px;
		border-bottom: 1px solid #202225;
		background-color: #2f3136;
	}

	.peer-info {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.peer-avatar img,
	.peer-avatar .avatar-placeholder {
		width: 40px;
		height: 40px;
		border-radius: 50%;
	}

	.avatar-placeholder {
		background: linear-gradient(135deg, #5865f2 0%, #7289da 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		color: #fff;
		font-size: 18px;
		font-weight: 600;
	}

	.peer-details h2 {
		color: #fff;
		font-size: 18px;
		margin: 0 0 4px 0;
		font-weight: 600;
	}

	.peer-status {
		font-size: 12px;
		color: #72767d;
	}

	.peer-status.online {
		color: #3ba55d;
	}

	.messages-container {
		flex: 1;
		overflow-y: auto;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.loading,
	.no-messages {
		text-align: center;
		color: #72767d;
		padding: 32px;
	}

	.date-separator {
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 16px 0;
	}

	.date-separator span {
		background-color: #2f3136;
		color: #72767d;
		padding: 4px 12px;
		border-radius: 12px;
		font-size: 12px;
		font-weight: 500;
	}

	.message {
		display: flex;
		margin-bottom: 2px;
		animation: slideIn 0.2s ease-out;
	}

	@keyframes slideIn {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.message.sent {
		justify-content: flex-end;
	}

	.message.received {
		justify-content: flex-start;
	}

	.message-bubble {
		max-width: 60%;
		padding: 8px 12px;
		border-radius: 18px;
		position: relative;
	}

	.message.received .message-bubble {
		background-color: #40444b;
		color: #dcddde;
		border-bottom-left-radius: 4px;
	}

	.message.sent .message-bubble {
		background-color: #5865f2;
		color: #fff;
		border-bottom-right-radius: 4px;
	}

	.message-text {
		margin: 0 0 4px 0;
		word-wrap: break-word;
		font-size: 15px;
		line-height: 1.4;
	}

	.message-footer {
		display: flex;
		align-items: center;
		gap: 4px;
		justify-content: flex-end;
	}

	.message-time {
		font-size: 11px;
		opacity: 0.7;
	}

	.message-status {
		display: flex;
		align-items: center;
		opacity: 0.8;
	}

	.message-status svg {
		width: 14px;
		height: 14px;
	}

	.message-input-container {
		padding: 16px;
		background-color: #2f3136;
		border-top: 1px solid #202225;
		display: flex;
		gap: 8px;
	}

	.message-input {
		flex: 1;
		padding: 12px 16px;
		background-color: #40444b;
		border: none;
		border-radius: 8px;
		color: #dcddde;
		font-size: 14px;
	}

	.message-input:focus {
		outline: none;
		background-color: #484c52;
	}

	.message-input:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.send-button {
		padding: 12px 20px;
		background-color: #5865f2;
		border: none;
		border-radius: 8px;
		color: #fff;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: background-color 0.15s ease;
	}

	.send-button:hover:not(:disabled) {
		background-color: #4752c4;
	}

	.send-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.spinner {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}

	/* Scrollbar styling */
	.messages-container::-webkit-scrollbar {
		width: 8px;
	}

	.messages-container::-webkit-scrollbar-track {
		background: #2f3136;
	}

	.messages-container::-webkit-scrollbar-thumb {
		background: #202225;
		border-radius: 4px;
	}

	.messages-container::-webkit-scrollbar-thumb:hover {
		background: #1a1c1e;
	}
</style>
