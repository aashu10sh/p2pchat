export interface Peer {
	ID: number;
	peer_id: string;
	username: string;
	wifiname: string;
	address: string;
	is_online: boolean;
	image_url: string;
	last_seen: string;
	CreatedAt: string;
	UpdatedAt: string;
	DeletedAt: string | null;
}
