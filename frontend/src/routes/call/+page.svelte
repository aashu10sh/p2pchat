<script lang="ts">
	import { type Peer } from '$lib/entites/peer';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let pc: RTCPeerConnection | null = null;

	async function generateSignalingData(to: string) {

		const localVideoElement = document.getElementById('live-video') as HTMLVideoElement;

		const stream = await navigator.mediaDevices.getUserMedia({
			video: true,
			audio: true,
		});

		localVideoElement.srcObject = stream;

		pc = new RTCPeerConnection();

		stream.getTracks().forEach(track => pc!.addTrack(track, stream));

		const offer = await pc.createOffer();
		await pc.setLocalDescription(offer);

		// Wait for ICE gathering to complete so candidates are bundled into the SDP
		await new Promise<void>(resolve => {
			pc!.onicegatheringstatechange = () => {
				if (pc!.iceGatheringState === 'complete') resolve();
			};
		});

		console.log(pc.localDescription)

		// const response = await fetch('/call/offer', {
		// 	method: 'POST',
		// 	headers: { 'Content-Type': 'application/json' },
		// 	body: JSON.stringify({ sdp: pc.localDescription, to })
		// });

		// const { answer } = await response.json();

		// await pc.setRemoteDescription(new RTCSessionDescription(answer));
	}
</script>


<div class="container flex flex-col items-center justify-center h-screen">
	<div>
		{#if data.error === null}
			{#if !data.data?.is_online}
				<h1>Sorry, <code>{data.data.username}</code> are not currently online right now, we cannot call them!</h1>
			{:else}
				{@render callPageSnippet(data.data)}
			{/if}
		{:else}
			<h1>Error</h1>
			<p>{data.error}</p>
		{/if}
	</div>
</div>

{#snippet callPageSnippet(peerInfo: Peer )}
<div class="flex flex-col">
	<h1>Would you like to call <code>{peerInfo.username}</code>?</h1>
	<div class="flex flex-row gap-20 justify-center">
		<button onclick={()=>{generateSignalingData(peerInfo.peer_id)}}>Yes</button>
		<button onclick={() => {alert("no")}}>No</button>
	</div>
	<video id="live-video"></video>
</div>
{/snippet}
