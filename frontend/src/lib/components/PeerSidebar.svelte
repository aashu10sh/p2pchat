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

	onMount(() => {
		// Initial fetch
		peerService.fetchPeers();

		// Subscribe to store updates (SSE is started at dashboard level)
		const unsubscribe = peers.subscribe((value) => {
			peerList = value;
		});

		return () => {
			unsubscribe();
		};
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

		if (diffMinutes < 1) return 'JUST_NOW';
		if (diffMinutes < 60) return `${diffMinutes}M`;
		if (diffHours < 24) return `${diffHours}H`;
		return `${diffDays}D`;
	}
</script>

<aside class="peer-sidebar">
	<div class="sidebar-header">
		<h2>NODE_LIST</h2>
		<span class="peer-count">[{peerList.length}]</span>
	</div>

	<div class="peer-list">
		{#if peerList.length === 0}
			<div class="no-peers">
				<p>NO_NODES_FOUND</p>
				<small>AWAITING_CONNECTIONS...</small>
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
					<div class="status-col">
						<div class="status-indicator"></div>
					</div>
					<div class="peer-info">
						<div class="peer-username">@{peer.username}</div>
						<div class="peer-status">
							{#if isRecentlyOnline(peer)}
								<span class="online-text">ACTIVE</span>
							{:else}
								<span class="offline-text">DC:{formatLastSeen(peer.last_seen)}</span>
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
		width: 250px;
		height: 100vh;
		background-color: var(--bg-primary);
		display: flex;
		flex-direction: column;
		border-right: 1px solid var(--border-color);
		font-family: var(--font-mono);
	}

	.sidebar-header {
		padding: 20px;
		border-bottom: 1px solid var(--border-color);
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.sidebar-header h2 {
		color: var(--text-primary);
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.1em;
		margin: 0;
	}

	.peer-count {
		color: var(--text-accent);
		font-size: 11px;
	}

	.peer-list {
		flex: 1;
		overflow-y: auto;
		padding: 12px 12px;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.no-peers {
		text-align: center;
		padding: 32px 16px;
		color: var(--text-secondary);
	}

	.no-peers p {
		margin: 0 0 8px 0;
		font-size: 12px;
		letter-spacing: 0.05em;
	}

	.no-peers small {
		font-size: 10px;
		opacity: 0.5;
		letter-spacing: 0.05em;
	}

	.peer-item {
		display: flex;
		align-items: flex-start;
		padding: 10px 12px;
		border-radius: 4px;
		cursor: pointer;
		transition: background-color 0.1s ease;
		background-color: transparent;
	}

	.peer-item:hover {
		background-color: var(--bg-secondary);
	}

	.peer-item.active {
		background-color: var(--bg-tertiary);
		border-left: 2px solid var(--text-accent);
		padding-left: 10px; /* Adjust for border */
	}

	.peer-item:focus {
		outline: 1px solid var(--text-accent);
		outline-offset: -1px;
	}

	.status-col {
		width: 16px;
		display: flex;
		justify-content: center;
		padding-top: 5px;
		margin-right: 8px;
	}

	.status-indicator {
		width: 6px;
		height: 6px;
		border-radius: 0;
		background-color: var(--border-color);
	}

	.peer-item.online .status-indicator {
		background-color: var(--text-accent);
		box-shadow: 0 0 5px var(--text-accent);
	}

	.peer-item.active.online .status-indicator {
		animation: pulse 2s infinite ease-in-out;
	}

	.peer-info {
		flex: 1;
		min-width: 0;
	}

	.peer-username {
		color: var(--text-primary);
		font-size: 13px;
		font-family: var(--font-sans);
		font-weight: 500;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		line-height: 1.2;
	}

	.peer-status {
		font-size: 10px;
		margin-top: 4px;
		letter-spacing: 0.05em;
	}

	.online-text {
		color: var(--text-accent);
	}

	.offline-text {
		color: var(--text-secondary);
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 0.5;
			box-shadow: 0 0 2px var(--text-accent);
		}
		50% {
			opacity: 1;
			box-shadow: 0 0 8px var(--text-accent);
		}
	}
</style>
