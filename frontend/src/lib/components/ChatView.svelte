<script lang="ts">
	import { onMount } from 'svelte';
	import ChatService, { messages } from '$lib/services/chatService';
	import { peers } from '$lib/services/peerService';
	import { startCall } from '$lib/services/callService';
	import FileSharing from '$lib/components/FileSharing.svelte';
	import type { Message } from '$lib/entites/message';
	import type { Peer } from '$lib/entites/peer';
	import {
		isRecentlyOnline,
		formatTime,
		formatDate,
		sortMessagesByTimestamp,
		getInitials
	} from '$lib/utils';

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
	let currentTab = $state<'chat' | 'files'>('chat');

	const chatService = new ChatService();

	// Update currentPeer reactively when peers store changes
	$effect(() => {
		if (!activePeer) {
			currentPeer = null;
			return;
		}

		currentPeer = activePeer;

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
				messageList = sortMessagesByTimestamp(msgs);
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
				isSending = false;
				setTimeout(() => {
					if (inputRef) inputRef.focus();
				}, 10);
			},
			(error) => {
				console.error('Failed to send message:', error);
				messageInput = content;
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

	/**
	 * Robustly determines if a message was sent by the current user.
	 * Primary: compare from_peer_id against myPeerId (reliable).
	 * Fallback: is_sent_by_me flag from server.
	 */
	function isMine(message: Message): boolean {
		if (myPeerId && message.from_peer_id) {
			return message.from_peer_id === myPeerId;
		}
		return message.is_sent_by_me;
	}

	onMount(() => {
		const unsubscribe = messages.subscribe((value) => {
			if (activePeer && value.length > 0) {
				messageList = sortMessagesByTimestamp(value);
				setTimeout(scrollToBottom, 50);
			}
		});

		return () => {
			unsubscribe();
		};
	});
</script>

{#if !activePeer}
	<div class="empty-chat">
		<div class="empty-content">
			<div class="empty-icon">
				<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
				</svg>
			</div>
			<h3>Select a conversation</h3>
			<p>Choose a contact from the sidebar to begin chatting.</p>
		</div>
	</div>
{:else}
	<div class="chat-view">
		<!-- Header -->
		<div class="chat-header">
			<div class="header-left">
				<div class="header-avatar" class:online={currentPeer && isRecentlyOnline(currentPeer)}>
					{getInitials(currentPeer?.username || '')}
				</div>
				<div class="header-info">
					<h2>{currentPeer?.username}</h2>
					<span class="header-status" class:online={currentPeer && isRecentlyOnline(currentPeer)}>
						{currentPeer && isRecentlyOnline(currentPeer) ? 'Active now' : 'Offline'}
					</span>
				</div>
			</div>
			
			<div class="header-tabs">
				<button 
					class="tab-btn" 
					class:active={currentTab === 'chat'} 
					onclick={() => currentTab = 'chat'}
				>
					Messages
				</button>
				<button 
					class="tab-btn" 
					class:active={currentTab === 'files'} 
					onclick={() => currentTab = 'files'}
				>
					Files
				</button>
			</div>

			<div class="header-actions">
				{#if currentPeer && isRecentlyOnline(currentPeer)}
					<button class="action-btn call-btn" onclick={handleCallPeer} title="Start video call">
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<polygon points="23 7 16 12 23 17 23 7"></polygon>
							<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
						</svg>
					</button>
				{/if}
			</div>
		</div>

		{#if currentTab === 'chat'}
			<!-- Messages -->
			<div class="messages-area" bind:this={chatContainer}>
				{#if isLoading}
					<div class="loading-state">
						<div class="skeleton-group">
							{#each [1, 2, 3] as _}
								<div class="skeleton-row left">
									<div class="skeleton-bubble" style="width: {140 + Math.random() * 80}px"></div>
								</div>
							{/each}
							<div class="skeleton-row right">
								<div class="skeleton-bubble" style="width: {100 + Math.random() * 60}px"></div>
							</div>
							<div class="skeleton-row left">
								<div class="skeleton-bubble" style="width: {160 + Math.random() * 60}px"></div>
							</div>
						</div>
					</div>
				{:else if messageList.length === 0}
					<div class="no-messages">
						<div class="no-msg-icon">👋</div>
						<p class="no-msg-title">No messages yet</p>
						<p class="no-msg-sub">Send a message to start the conversation.</p>
					</div>
				{:else}
					{#each messageList as message, index (message.ID)}
						<!-- Date separator -->
						{#if index === 0 || formatDate(messageList[index - 1].sent_at) !== formatDate(message.sent_at)}
							<div class="date-divider">
								<span>{formatDate(message.sent_at)}</span>
							</div>
						{/if}

						<!-- Message bubble -->
						{@const mine = isMine(message)}
						<div
							class="message-row"
							class:sent={mine}
							class:received={!mine}
						>
							<div class="bubble">
								<p class="bubble-text">{message.content}</p>
								<div class="bubble-meta">
									<span class="bubble-time">{formatTime(message.sent_at)}</span>
									{#if mine}
										{#if message.delivered_at}
											<svg class="check-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
												<polyline points="20 6 9 17 4 12"></polyline>
											</svg>
										{:else}
											<svg class="check-icon pending" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
												<circle cx="12" cy="12" r="10"></circle>
											</svg>
										{/if}
									{/if}
								</div>
							</div>
						</div>
					{/each}
				{/if}
			</div>

			<!-- Input -->
			<form class="input-bar" onsubmit={handleSendMessage}>
				<div class="input-container" class:focused={isFocused}>
					<input
						bind:this={inputRef}
						bind:value={messageInput}
						type="text"
						placeholder="Type a message..."
						autocomplete="off"
						spellcheck="false"
						disabled={isSending}
						onfocus={() => (isFocused = true)}
						onblur={() => (isFocused = false)}
					/>
					<button
						type="submit"
						class="send-btn"
						disabled={!messageInput.trim() || isSending}
						title="Send message"
					>
						{#if isSending}
							<div class="send-spinner"></div>
						{:else}
							<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
								<path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"></path>
							</svg>
						{/if}
					</button>
				</div>
			</form>
		{:else}
			<FileSharing {activePeer} />
		{/if}
	</div>
{/if}

<style>
	/* ── Empty State ── */
	.empty-chat {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-primary);
	}

	.empty-content {
		text-align: center;
	}

	.empty-icon {
		color: var(--text-tertiary);
		margin-bottom: 16px;
		opacity: 0.6;
	}

	.empty-content h3 {
		font-size: 17px;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 6px;
	}

	.empty-content p {
		font-size: 14px;
		color: var(--text-secondary);
		margin: 0;
	}

	/* ── Chat View ── */
	.chat-view {
		flex: 1;
		display: flex;
		flex-direction: column;
		background: var(--bg-primary);
	}

	/* ── Header ── */
	.chat-header {
		padding: 14px 20px;
		border-bottom: 1px solid var(--border-color);
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: var(--glass-bg-heavy);
		backdrop-filter: blur(var(--glass-blur));
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.header-avatar {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		background: var(--bg-tertiary);
		color: var(--text-secondary);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 14px;
		font-weight: 600;
	}

	.header-avatar.online {
		background: linear-gradient(135deg, var(--accent), #5856d6);
		color: white;
	}

	.header-info h2 {
		font-size: 15px;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
		letter-spacing: -0.01em;
	}

	.header-status {
		font-size: 12px;
		color: var(--text-secondary);
	}

	.header-status.online {
		color: var(--success);
	}

	.header-actions {
		display: flex;
		gap: 8px;
	}

	.header-tabs {
		display: flex;
		background: var(--bg-tertiary);
		border-radius: var(--radius-full);
		padding: 2px;
		margin-left: 20px;
	}

	.tab-btn {
		border: none;
		background: transparent;
		padding: 4px 14px;
		font-size: 13px;
		font-weight: 500;
		color: var(--text-secondary);
		border-radius: var(--radius-full);
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.tab-btn:hover {
		color: var(--text-primary);
	}

	.tab-btn.active {
		background: var(--bg-primary);
		color: var(--text-primary);
		box-shadow: 0 1px 3px rgba(0,0,0,0.1);
	}

	.action-btn {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		border: none;
		background: var(--bg-secondary);
		color: var(--text-secondary);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all var(--transition-fast);
	}

	.action-btn:hover {
		background: var(--accent-light);
		color: var(--accent);
	}

	.action-btn.call-btn:hover {
		background: var(--accent);
		color: white;
	}

	/* ── Messages Area ── */
	.messages-area {
		flex: 1;
		overflow-y: auto;
		padding: 16px 20px;
		display: flex;
		flex-direction: column;
		gap: 4px;
		background: var(--bg-secondary);
	}

	/* ── Loading Skeleton ── */
	.loading-state {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 32px;
	}

	.skeleton-group {
		width: 100%;
		max-width: 500px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.skeleton-row {
		display: flex;
	}

	.skeleton-row.left {
		justify-content: flex-start;
	}

	.skeleton-row.right {
		justify-content: flex-end;
	}

	.skeleton-bubble {
		height: 36px;
		border-radius: 18px;
		background: linear-gradient(90deg, var(--bg-tertiary) 25%, #e0e0e5 50%, var(--bg-tertiary) 75%);
		background-size: 200% 100%;
		animation: shimmer 1.5s ease-in-out infinite;
	}

	@keyframes shimmer {
		0% {
			background-position: 200% 0;
		}
		100% {
			background-position: -200% 0;
		}
	}

	/* ── No Messages ── */
	.no-messages {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		padding: 40px;
	}

	.no-msg-icon {
		font-size: 40px;
		margin-bottom: 12px;
	}

	.no-msg-title {
		font-size: 16px;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 4px;
	}

	.no-msg-sub {
		font-size: 13px;
		color: var(--text-secondary);
		margin: 0;
	}

	/* ── Date Divider ── */
	.date-divider {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 12px 0;
	}

	.date-divider span {
		font-size: 12px;
		font-weight: 500;
		color: var(--text-secondary);
		background: var(--bg-primary);
		padding: 4px 14px;
		border-radius: var(--radius-full);
		box-shadow: var(--shadow-sm);
	}

	/* ── Message Rows ── */
	.message-row {
		display: flex;
		animation: messageSlideIn var(--transition-spring) both;
	}

	.message-row.sent {
		justify-content: flex-end;
	}

	.message-row.received {
		justify-content: flex-start;
	}

	@keyframes messageSlideIn {
		from {
			opacity: 0;
			transform: translateY(12px) scale(0.97);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}

	/* ── Bubble ── */
	.bubble {
		max-width: 75%;
		padding: 10px 14px;
		border-radius: 18px;
		position: relative;
		word-wrap: break-word;
	}

	.message-row.sent .bubble {
		background: var(--accent);
		color: white;
		border-bottom-right-radius: 6px;
	}

	.message-row.received .bubble {
		background: var(--bg-primary);
		color: var(--text-primary);
		border-bottom-left-radius: 6px;
		box-shadow: var(--shadow-sm);
	}

	.bubble-text {
		margin: 0;
		font-size: 14px;
		line-height: 1.45;
	}

	.bubble-meta {
		display: flex;
		align-items: center;
		gap: 4px;
		justify-content: flex-end;
		margin-top: 4px;
	}

	.bubble-time {
		font-size: 11px;
		opacity: 0.6;
		font-variant-numeric: tabular-nums;
	}

	.message-row.sent .bubble-time {
		color: rgba(255, 255, 255, 0.7);
	}

	.message-row.received .bubble-time {
		color: var(--text-secondary);
	}

	.check-icon {
		opacity: 0.7;
	}

	.message-row.sent .check-icon {
		color: rgba(255, 255, 255, 0.8);
	}

	.check-icon.pending {
		opacity: 0.4;
	}

	/* ── Input Bar ── */
	.input-bar {
		padding: 12px 16px;
		background: var(--bg-primary);
		border-top: 1px solid var(--border-color);
	}

	.input-container {
		display: flex;
		align-items: center;
		background: var(--bg-secondary);
		border-radius: var(--radius-full);
		border: 1.5px solid transparent;
		padding: 4px 4px 4px 18px;
		transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
	}

	.input-container.focused {
		border-color: var(--accent);
		box-shadow: 0 0 0 3px var(--accent-light);
		background: var(--bg-primary);
	}

	.input-container input {
		flex: 1;
		background: transparent;
		border: none;
		color: var(--text-primary);
		font-family: var(--font-sans);
		font-size: 14px;
		outline: none;
		padding: 8px 0;
	}

	.input-container input::placeholder {
		color: var(--text-tertiary);
	}

	.input-container input:disabled {
		opacity: 0.5;
	}

	.send-btn {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		border: none;
		background: var(--accent);
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all var(--transition-fast);
		flex-shrink: 0;
	}

	.send-btn:hover:not(:disabled) {
		background: var(--accent-hover);
		transform: scale(1.05);
	}

	.send-btn:active:not(:disabled) {
		transform: scale(0.95);
	}

	.send-btn:disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}

	.send-spinner {
		width: 16px;
		height: 16px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top-color: white;
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	/* ── Scrollbar ── */
	.messages-area::-webkit-scrollbar {
		width: 6px;
	}

	.messages-area::-webkit-scrollbar-track {
		background: transparent;
	}

	.messages-area::-webkit-scrollbar-thumb {
		background: rgba(0, 0, 0, 0.12);
		border-radius: var(--radius-full);
	}

	.messages-area::-webkit-scrollbar-thumb:hover {
		background: rgba(0, 0, 0, 0.2);
	}
</style>
