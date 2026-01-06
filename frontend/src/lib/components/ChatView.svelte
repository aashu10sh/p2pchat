<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import ChatService, { messages } from '$lib/services/chatService';
	import PeerService from '$lib/services/peerService';
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
	let chatContainer: HTMLDivElement;

	const chatService = new ChatService();
	const peerService = new PeerService();

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
				messageList = msgs.reverse(); // Show oldest first
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
		if (!messageInput.trim() || !activePeer) return;

		const content = messageInput;
		messageInput = '';

		const result = await chatService.sendMessage(activePeer.peer_id, content);
		result.match(
			() => {
				// Message sent, will be updated via SSE
			},
			(error) => {
				console.error('Failed to send message:', error);
				messageInput = content; // Restore message on error
			}
		);
	}

	function scrollToBottom() {
		if (chatContainer) {
			chatContainer.scrollTop = chatContainer.scrollHeight;
		}
	}

	function formatTime(timestamp: string): string {
		const date = new Date(timestamp);
		return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
	}

	onMount(() => {
		// Subscribe to messages store
		const unsubscribe = messages.subscribe((value) => {
			if (activePeer && value.length > 0) {
				messageList = [...value].reverse();
				setTimeout(scrollToBottom, 50);
			}
		});

		// Subscribe to SSE for real-time updates
		peerService.startEventStream((peers) => {
			// Handled by peer service
		});

		return () => {
			unsubscribe();
		};
	});

	onDestroy(() => {
		peerService.stopEventStream();
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
					{#if activePeer.image_url}
						<img src={activePeer.image_url} alt={activePeer.username} />
					{:else}
						<div class="avatar-placeholder">
							{activePeer.username.charAt(0).toUpperCase()}
						</div>
					{/if}
				</div>
				<div class="peer-details">
					<h2>{activePeer.username}</h2>
					<span class="peer-status" class:online={activePeer.is_online}>
						{activePeer.is_online ? 'Online' : 'Offline'}
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
				{#each messageList as message (message.ID)}
					<div class="message" class:own={message.is_sent_by_me}>
						<div class="message-content">
							<p>{message.content}</p>
							<span class="message-time">{formatTime(message.sent_at)}</span>
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
			/>
			<button type="submit" class="send-button" disabled={!messageInput.trim()}>
				<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
					<path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
				</svg>
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
		gap: 8px;
	}

	.loading,
	.no-messages {
		text-align: center;
		color: #72767d;
		padding: 32px;
	}

	.message {
		display: flex;
		margin-bottom: 4px;
	}

	.message.own {
		justify-content: flex-end;
	}

	.message-content {
		max-width: 60%;
		padding: 10px 14px;
		border-radius: 18px;
		background-color: #40444b;
		color: #dcddde;
	}

	.message.own .message-content {
		background-color: #5865f2;
		color: #fff;
	}

	.message-content p {
		margin: 0 0 4px 0;
		word-wrap: break-word;
	}

	.message-time {
		font-size: 10px;
		opacity: 0.7;
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
