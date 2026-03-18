import PeerService from '$lib/services/peerService.js';

export async function load({ url }) {
	const peerId = url.searchParams.get('peer_id');
	if (peerId == null) {
		return {
			error: 'need peer_id',
			data: null
		};
	}
	const peerService = new PeerService();
	const peerResult = await peerService.fetchPeer(peerId);

	return peerResult.match(
		(data) => {
			return {
				data: data,
				error: null
			};
		},
		(err) => {
			return {
				error: err.error,
				data: null
			};
		}
	);
	// Return the data to the page component as props
}
