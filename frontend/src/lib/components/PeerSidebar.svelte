<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import PeerService, { peers } from '$lib/services/peerService';
	import type { Peer } from '$lib/entites/peer';

	interface Props {
		onPeerSelect?: (peer: Peer) => void;
		activePeerId?: string | null;
	}

	let { onPeerSelect, activePeerId = null }: Props = $props();

	const peerService = new PeerService();
	let peerList = $state<Peer[]>([]);

	function handlePeerClick(peer: Peer) {
		if (onPeerSelect) {
			onPeerSelect(peer);
		}
	}

	onMount(async () => {
		// Initial fetch
		await peerService.fetchPeers();

		// Start SSE connection
		peerService.startEventStream();

		// Subscribe to store updates
		const unsubscribe = peers.subscribe((value) => {
			peerList = value;
		});

		return () => {
			unsubscribe();
		};
	});

	onDestroy(() => {
		peerService.stopEventStream();
	});

	function isRecentlyOnline(peer: Peer): boolean {
		if (!peer.last_seen) return false;
		const lastSeen = new Date(peer.last_seen);
		const now = new Date();
		const diffMinutes = (now.getTime() - lastSeen.getTime()) / 1000 / 60;
		return diffMinutes < 1; // Online if seen in last minute
	}

	function formatLastSeen(lastSeen: string): string {
		const date = new Date(lastSeen);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMinutes = Math.floor(diffMs / 1000 / 60);
		const diffHours = Math.floor(diffMinutes / 60);
		const diffDays = Math.floor(diffHours / 24);

		if (diffMinutes < 1) return 'Just now';
		if (diffMinutes < 60) return `${diffMinutes}m ago`;
		if (diffHours < 24) return `${diffHours}h ago`;
		return `${diffDays}d ago`;
	}
</script>

<aside class="peer-sidebar">
	<div class="sidebar-header">
		<h2>Peers</h2>
		<span class="peer-count">{peerList.length}</span>
	</div>

	<div class="peer-list">
		{#if peerList.length === 0}
			<div class="no-peers">
				<p>No peers discovered yet</p>
				<small>Waiting for peers on the network...</small>
			</div>
		{:else}
			{#each peerList as peer (peer.peer_id)}
				<div
					class="peer-item"
					class:online={isRecentlyOnline(peer)}
					class:active={activePeerId === peer.peer_id}
					role="button"
					tabindex="0"
					onclick={() => handlePeerClick(peer)}
					onkeydown={(e) => e.key === 'Enter' && handlePeerClick(peer)}
				>
					<div class="peer-avatar">
						{#if peer.image_url}
							<img src={peer.image_url} alt={peer.username} />
						{:else}
							<div class="avatar-placeholder">
								{peer.username.charAt(0).toUpperCase()}
							</div>
						{/if}
						<div class="status-indicator" class:online={isRecentlyOnline(peer)}></div>
					</div>
					<div class="peer-info">
						<div class="peer-username">{peer.username}</div>
						<div class="peer-status">
							{#if isRecentlyOnline(peer)}
								<span class="online-text">Online</span>
							{:else}
								<span class="offline-text">{formatLastSeen(peer.last_seen)}</span>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		{/if}
	</div>
</aside>

<style>
	.peer-sidebar {
		width: 240px;
		height: 100vh;
		background-color: #2f3136;
		display: flex;
		flex-direction: column;
		border-right: 1px solid #202225;
	}

	.sidebar-header {
		padding: 16px;
		border-bottom: 1px solid #202225;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.sidebar-header h2 {
		color: #fff;
		font-size: 16px;
		font-weight: 600;
		margin: 0;
	}

	.peer-count {
		background-color: #5865f2;
		color: #fff;
		font-size: 12px;
		padding: 2px 8px;
		border-radius: 12px;
		font-weight: 600;
	}

	.peer-list {
		flex: 1;
		overflow-y: auto;
		padding: 8px;
	}

	.no-peers {
		text-align: center;
		padding: 32px 16px;
		color: #b9bbbe;
	}

	.no-peers p {
		margin: 0 0 8px 0;
		font-size: 14px;
	}

	.no-peers small {
		font-size: 12px;
		color: #72767d;
	}

	.peer-item {
		display: flex;
		align-items: center;
		padding: 8px;
		border-radius: 8px;
		margin-bottom: 4px;
		cursor: pointer;
		transition: background-color 0.15s ease;
	}

	.peer-item:hover {
		background-color: #393c43;
	}

	.peer-item.active {
		background-color: #404449;
	}

	.peer-item:focus {
		outline: 2px solid #5865f2;
		outline-offset: -2px;
	}

	.peer-avatar {
		position: relative;
		margin-right: 12px;
	}

	.peer-avatar img,
	.avatar-placeholder {
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

	.status-indicator {
		position: absolute;
		bottom: 0;
		right: 0;
		width: 12px;
		height: 12px;
		border-radius: 50%;
		background-color: #747f8d;
		border: 3px solid #2f3136;
	}

	.status-indicator.online {
		background-color: #3ba55d;
	}

	.peer-info {
		flex: 1;
		min-width: 0;
	}

	.peer-username {
		color: #fff;
		font-size: 14px;
		font-weight: 500;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.peer-status {
		font-size: 12px;
		margin-top: 2px;
	}

	.online-text {
		color: #3ba55d;
	}

	.offline-text {
		color: #72767d;
	}

	/* Scrollbar styling */
	.peer-list::-webkit-scrollbar {
		width: 8px;
	}

	.peer-list::-webkit-scrollbar-track {
		background: #2f3136;
	}

	.peer-list::-webkit-scrollbar-thumb {
		background: #202225;
		border-radius: 4px;
	}

	.peer-list::-webkit-scrollbar-thumb:hover {
		background: #1a1c1e;
	}
</style>
