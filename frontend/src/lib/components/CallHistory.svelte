<script lang="ts">
	import type { Peer } from '$lib/entites/peer';
	import type { CallHistory } from '$lib/entites/callHistory';
	import CallHistoryService, { callHistories } from '$lib/services/callHistoryService';

	interface Props {
		activePeer: Peer | null;
	}
	let { activePeer }: Props = $props();

	const callHistoryService = new CallHistoryService();
	let historyList = $state<CallHistory[]>([]);

	$effect(() => {
		if (activePeer) {
			loadHistories();
		}
	});

	$effect(() => {
		const unsub = callHistories.subscribe((list) => {
			if (activePeer) {
				historyList = list.filter(
					(h) => h.from_peer_id === activePeer?.peer_id || h.to_peer_id === activePeer?.peer_id
				);
			}
		});
		return unsub;
	});

	async function loadHistories() {
		if (!activePeer) return;
		const result = await callHistoryService.fetchCallHistories(activePeer.peer_id);
		result.match(
			() => {}, // store handles update
			(err) => console.error('Failed to load call histories:', err)
		);
	}

	function formatDuration(seconds: number) {
		if (seconds === 0) return '0s';
		const m = Math.floor(seconds / 60);
		const s = seconds % 60;
		if (m > 0) {
			return `${m}m ${s}s`;
		}
		return `${s}s`;
	}
</script>

<div class="call-history-panel">
	<div class="history-list-container">
		<h3 class="title">Call History with {activePeer?.username}</h3>
		{#if historyList.length === 0}
			<p class="empty">No call history with this peer.</p>
		{:else}
			<div class="list">
				{#each historyList as ch}
					<div class="history-card" class:missed={ch.status === 'missed' || ch.status === 'rejected'}>
						<div class="card-header">
							<span class="direction" class:outgoing={ch.direction === 'outgoing'} class:incoming={ch.direction === 'incoming'}>
								{ch.direction === 'outgoing' ? '↗ Outgoing' : '↙ Incoming'}
							</span>
							<span class="time">{new Date(ch.created_at).toLocaleString()}</span>
						</div>
						<div class="meta">
							<span class="status-badge" class:completed={ch.status === 'completed'} class:rejected={ch.status === 'rejected' || ch.status === 'missed'}>
								{ch.status.charAt(0).toUpperCase() + ch.status.slice(1)}
							</span>
							{#if ch.status === 'completed'}
								<span class="duration">Duration: {formatDuration(ch.duration)}</span>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<style>
	.call-history-panel {
		flex: 1;
		display: flex;
		flex-direction: column;
		background: var(--bg-primary);
		overflow: hidden;
	}

	.history-list-container {
		flex: 1;
		padding: 20px;
		overflow-y: auto;
		background: var(--bg-primary);
	}

	.title {
		font-size: 15px;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 4px;
	}

	.empty {
		font-size: 13px;
		color: var(--text-tertiary);
		text-align: center;
		padding: 40px 0;
	}

	.list {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin-top: 16px;
	}

	.history-card {
		padding: 14px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-color);
		background: var(--bg-secondary);
	}

	.history-card.missed {
		border-left: 3px solid var(--danger);
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}

	.direction {
		font-size: 12px;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.direction.outgoing {
		color: var(--text-primary);
	}

	.direction.incoming {
		color: var(--accent);
	}

	.time {
		font-size: 11px;
		color: var(--text-secondary);
	}

	.meta {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.status-badge {
		font-size: 11px;
		font-weight: 600;
		padding: 3px 8px;
		border-radius: var(--radius-sm);
	}

	.status-badge.completed {
		background: rgba(52, 199, 89, 0.1);
		color: var(--success);
	}

	.status-badge.rejected {
		background: rgba(255, 59, 48, 0.1);
		color: var(--danger);
	}

	.duration {
		font-size: 12px;
		color: var(--text-secondary);
		font-family: var(--font-mono);
	}
</style>
