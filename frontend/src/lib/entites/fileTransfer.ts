export interface FileTransfer {
	id: number;
	from_peer_id: string;
	to_peer_id: string;
	file_name: string;
	file_size: number;
	file_path: string;
	direction: string;
	created_at: string;
}
