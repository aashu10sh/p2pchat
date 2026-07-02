<script lang="ts">
	import { onMount } from 'svelte';
	import type { Peer } from '$lib/entites/peer';
	import type { FileTransfer } from '$lib/entites/fileTransfer';
	import FileService, { fileTransfers } from '$lib/services/fileService';

	interface Props {
		activePeer: Peer | null;
	}
	let { activePeer }: Props = $props();

	let filePathInput = $state('');
	let isSending = $state(false);
	let errorMsg = $state('');
	
	const fileService = new FileService();
	let transfersList = $state<FileTransfer[]>([]);

	$effect(() => {
		if (activePeer) {
			loadTransfers();
		}
	});

	$effect(() => {
		const unsub = fileTransfers.subscribe((list) => {
			if (activePeer) {
				transfersList = list.filter(
					(t) => t.from_peer_id === activePeer?.peer_id || t.to_peer_id === activePeer?.peer_id
				);
			}
		});
		return unsub;
	});

	async function loadTransfers() {
		if (!activePeer) return;
		const result = await fileService.fetchFileTransfers(activePeer.peer_id);
		result.match(
			() => {}, // store handles update
			(err) => console.error('Failed to load file transfers:', err)
		);
	}

	async function handleSend(e: Event) {
		e.preventDefault();
		if (!activePeer || !filePathInput.trim() || isSending) return;

		isSending = true;
		errorMsg = '';
		const result = await fileService.sendFile(activePeer.peer_id, filePathInput.trim());
		
		result.match(
			() => {
				filePathInput = '';
			},
			(err) => {
				errorMsg = err.error;
			}
		);
		isSending = false;
	}

	function formatBytes(bytes: number) {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}
</script>

<div class="file-sharing-panel">
	<div class="send-form">
		<h3 class="title">Send File to {activePeer?.username}</h3>
		<p class="subtitle">Enter the absolute file path on your device to send it via P2P.</p>
		<form onsubmit={handleSend} class="form">
			<input 
				type="text" 
				placeholder="/home/user/Documents/secret.txt" 
				bind:value={filePathInput}
				disabled={isSending}
			/>
			<button type="submit" disabled={isSending || !filePathInput.trim()}>
				{isSending ? 'Sending...' : 'Send'}
			</button>
		</form>
		{#if errorMsg}
			<p class="error">{errorMsg}</p>
		{/if}
	</div>

	<div class="transfers-history">
		<h3 class="title">File History</h3>
		{#if transfersList.length === 0}
			<p class="empty">No file transfers with this peer.</p>
		{:else}
			<div class="list">
				{#each transfersList as ft}
					<div class="transfer-card">
						<div class="card-header">
							<span class="direction" class:sent={ft.direction === 'sent'} class:received={ft.direction === 'received'}>
								{ft.direction === 'sent' ? '↑ Sent' : '↓ Received'}
							</span>
							<span class="time">{new Date(ft.created_at).toLocaleString()}</span>
						</div>
						<div class="filename">{ft.file_name}</div>
						<div class="meta">
							<span class="size">{formatBytes(ft.file_size)}</span>
						</div>
						<div class="path-container">
							<span class="path-label">Location:</span>
							<span class="path-value">{ft.file_path}</span>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<style>
	.file-sharing-panel {
		flex: 1;
		display: flex;
		flex-direction: column;
		background: var(--bg-primary);
		overflow: hidden;
	}

	.send-form {
		padding: 20px;
		border-bottom: 1px solid var(--border-color);
		background: var(--bg-secondary);
	}

	.title {
		font-size: 15px;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 4px;
	}

	.subtitle {
		font-size: 12px;
		color: var(--text-secondary);
		margin-bottom: 16px;
	}

	.form {
		display: flex;
		gap: 12px;
	}

	.form input {
		flex: 1;
		padding: 10px 14px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-color);
		background: var(--bg-primary);
		color: var(--text-primary);
		font-family: var(--font-mono);
		font-size: 13px;
	}

	.form input:focus {
		outline: none;
		border-color: var(--accent);
	}

	.form button {
		padding: 0 20px;
		border-radius: var(--radius-md);
		border: none;
		background: var(--accent);
		color: white;
		font-weight: 500;
		cursor: pointer;
		transition: background-color var(--transition-fast);
	}

	.form button:hover:not(:disabled) {
		background: var(--accent-hover);
	}

	.form button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.error {
		color: var(--danger);
		font-size: 12px;
		margin-top: 10px;
	}

	.transfers-history {
		flex: 1;
		padding: 20px;
		overflow-y: auto;
		background: var(--bg-primary);
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

	.transfer-card {
		padding: 14px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-color);
		background: var(--bg-secondary);
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 8px;
	}

	.direction {
		font-size: 11px;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: var(--radius-full);
	}

	.direction.sent {
		background: rgba(0, 122, 255, 0.1);
		color: var(--accent);
	}

	.direction.received {
		background: rgba(52, 199, 89, 0.1);
		color: var(--success);
	}

	.time {
		font-size: 11px;
		color: var(--text-secondary);
	}

	.filename {
		font-size: 14px;
		font-weight: 500;
		color: var(--text-primary);
		margin-bottom: 4px;
		word-break: break-all;
	}

	.meta {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 12px;
	}

	.size {
		font-size: 12px;
		color: var(--text-secondary);
	}

	.path-container {
		font-size: 11px;
		background: var(--bg-primary);
		padding: 8px 10px;
		border-radius: var(--radius-sm);
		border: 1px solid rgba(0,0,0,0.04);
		font-family: var(--font-mono);
		color: var(--text-secondary);
		display: flex;
		gap: 6px;
		word-break: break-all;
	}

	.path-label {
		color: var(--text-tertiary);
		user-select: none;
	}
	
	.path-value {
		user-select: all;
	}
</style>
