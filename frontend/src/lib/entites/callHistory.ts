export interface CallHistory {
	id: number;
	from_peer_id: string;
	to_peer_id: string;
	status: string;
	duration: number;
	direction: string;
	created_at: string;
}
