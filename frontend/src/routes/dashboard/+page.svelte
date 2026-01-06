<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { PageProps } from './$types';
	import type { Peer } from '$lib/entites/peer';
	import PeerSidebar from '$lib/components/PeerSidebar.svelte';
	import ChatView from '$lib/components/ChatView.svelte';
	import PeerService from '$lib/services/peerService';

	let { data }: PageProps = $props();
	let activePeer = $state<Peer | null>(null);

	const peerService = new PeerService();

	function handlePeerSelect(peer: Peer) {
		activePeer = peer;
	}

	onMount(() => {
		// Start SSE connection once at the dashboard level
		peerService.startEventStream();
	});

	onDestroy(() => {
		// Stop SSE when leaving dashboard
		peerService.stopEventStream();
	});
</script>

<div class="dashboard-container">
	<PeerSidebar onPeerSelect={handlePeerSelect} activePeerId={activePeer?.peer_id} />
	<ChatView activePeer={activePeer} myPeerId={data.user.peer_id} />
</div>

<style>
	.dashboard-container {
		display: flex;
		height: 100vh;
		background-color: #36393f;
	}
</style>
