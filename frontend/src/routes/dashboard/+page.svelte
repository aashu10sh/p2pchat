<script lang="ts">
	import type { PageProps } from './$types';
	import type { Peer } from '$lib/entites/peer';
	import PeerSidebar from '$lib/components/PeerSidebar.svelte';
	import ChatView from '$lib/components/ChatView.svelte';

	let { data }: PageProps = $props();
	let activePeer = $state<Peer | null>(null);

	function handlePeerSelect(peer: Peer) {
		activePeer = peer;
	}
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
