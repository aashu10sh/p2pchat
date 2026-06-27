export interface Message {
	ID: number;
	chat_id: number;
	from_peer_id: string;
	to_peer_id: string;
	content: string;
	type: string;
	sent_at: string;
	delivered_at: string | null;
	read_at: string | null;
	is_sent_by_me: boolean;
	CreatedAt: string;
	UpdatedAt: string;
}

export interface Chat {
	ID: number;
	my_peer_id: string;
	their_peer_id: string;
	their_username: string;
	last_message_at: string;
	unread_count: number;
	CreatedAt: string;
	UpdatedAt: string;
}
