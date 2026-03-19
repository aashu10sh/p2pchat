<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { PageProps } from './$types';
	import type { Peer } from '$lib/entites/peer';
	import PeerSidebar from '$lib/components/PeerSidebar.svelte';
	import ChatView from '$lib/components/ChatView.svelte';
	import IncomingCallModal from '$lib/components/IncomingCallModal.svelte';
	import ActiveCallView from '$lib/components/ActiveCallView.svelte';
	import PeerService from '$lib/services/peerService';

	let { data }: PageProps = $props();
	let activePeer = $state<Peer | null>(null);

	const peerService = new PeerService();

	function handlePeerSelect(peer: Peer) {
		activePeer = peer;
	}

	onMount(() => {
		peerService.startEventStream();
	});

	onDestroy(() => {
		peerService.stopEventStream();
	});
</script>

<div class="dashboard">
	<PeerSidebar onPeerSelect={handlePeerSelect} activePeerId={activePeer?.peer_id} />
	<ChatView {activePeer} myPeerId={data.user.peer_id} />

	<!-- Call Overlays -->
	<IncomingCallModal />
	<ActiveCallView />
</div>

<style>
	.dashboard {
		display: flex;
		height: 100vh;
		background-color: var(--bg-secondary);
	}
</style>
