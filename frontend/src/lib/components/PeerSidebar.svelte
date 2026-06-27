<script lang="ts">
	import { onMount } from 'svelte';
	import PeerService, { peers } from '$lib/services/peerService';
	import type { Peer } from '$lib/entites/peer';
	import { isRecentlyOnline, formatLastSeen, getInitials } from '$lib/utils';

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
		peerService.fetchPeers();

		const unsubscribe = peers.subscribe((value) => {
			peerList = value;
		});

		return () => {
			unsubscribe();
		};
	});
</script>

<aside class="sidebar">
	<div class="sidebar-header">
		<h2>Contacts</h2>
		<span class="count">{peerList.length}</span>
	</div>

	<div class="peer-list">
		{#if peerList.length === 0}
			<div class="empty-state">
				<div class="empty-icon">
					<svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
						<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
						<circle cx="9" cy="7" r="4"></circle>
						<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
						<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
					</svg>
				</div>
				<p class="empty-title">No peers found</p>
				<p class="empty-sub">Waiting for others to join the network...</p>
			</div>
		{:else}
			{#each peerList as peer (peer.peer_id)}
				{@const online = isRecentlyOnline(peer)}
				<button
					class="peer-item"
					class:online
					class:active={activePeerId === peer.peer_id}
					onclick={() => handlePeerClick(peer)}
				>
					<div class="avatar" class:online>
						{getInitials(peer.username)}
						{#if online}
							<div class="online-dot"></div>
						{/if}
					</div>
					<div class="peer-info">
						<span class="peer-name">{peer.username}</span>
						<span class="peer-status">
							{#if online}
								Active now
							{:else}
								{formatLastSeen(peer.last_seen)}
							{/if}
						</span>
					</div>
				</button>
			{/each}
		{/if}
	</div>
</aside>

<style>
	.sidebar {
		width: 280px;
		height: 100vh;
		background-color: var(--bg-primary);
		display: flex;
		flex-direction: column;
		border-right: 1px solid var(--border-color);
	}

	.sidebar-header {
		padding: 24px 20px 16px;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.sidebar-header h2 {
		font-size: 20px;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
		letter-spacing: -0.02em;
	}

	.count {
		font-size: 13px;
		font-weight: 500;
		color: var(--text-secondary);
		background: var(--bg-secondary);
		padding: 2px 10px;
		border-radius: var(--radius-full);
	}

	.peer-list {
		flex: 1;
		overflow-y: auto;
		padding: 4px 8px;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 48px 24px;
		text-align: center;
	}

	.empty-icon {
		color: var(--text-tertiary);
		margin-bottom: 16px;
	}

	.empty-title {
		font-size: 15px;
		font-weight: 500;
		color: var(--text-secondary);
		margin: 0 0 6px;
	}

	.empty-sub {
		font-size: 13px;
		color: var(--text-tertiary);
		margin: 0;
	}

	.peer-item {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 12px;
		border-radius: var(--radius-md);
		cursor: pointer;
		transition: background-color var(--transition-fast);
		background: none;
		border: none;
		text-align: left;
		font-family: inherit;
	}

	.peer-item:hover {
		background-color: var(--bg-secondary);
	}

	.peer-item.active {
		background-color: var(--accent-light);
	}

	.avatar {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background: var(--bg-tertiary);
		color: var(--text-secondary);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 15px;
		font-weight: 600;
		flex-shrink: 0;
		position: relative;
		transition: background var(--transition-fast);
	}

	.avatar.online {
		background: linear-gradient(135deg, var(--accent), #5856d6);
		color: white;
	}

	.online-dot {
		position: absolute;
		bottom: 0;
		right: 0;
		width: 12px;
		height: 12px;
		background: var(--success);
		border: 2.5px solid var(--bg-primary);
		border-radius: 50%;
	}

	.peer-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.peer-name {
		font-size: 14px;
		font-weight: 500;
		color: var(--text-primary);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.peer-status {
		font-size: 12px;
		color: var(--text-secondary);
	}

	.peer-item.active .peer-name {
		color: var(--accent);
	}

	.peer-item.online .peer-status {
		color: var(--success);
	}
</style>
