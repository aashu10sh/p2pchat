import type { Peer } from '$lib/entites/peer';
import type { Message } from '$lib/entites/message';

/**
 * Check if a peer was seen within the last 60 seconds.
 */
export function isRecentlyOnline(peer: Peer): boolean {
	if (!peer.last_seen) return false;
	const lastSeen = new Date(peer.last_seen);
	const now = new Date();
	const diffMinutes = (now.getTime() - lastSeen.getTime()) / 1000 / 60;
	return diffMinutes < 1;
}

/**
 * Format a timestamp string into HH:MM format.
 */
export function formatTime(timestamp: string): string {
	if (!timestamp) return '';
	try {
		const date = new Date(timestamp);
		if (isNaN(date.getTime())) return '';
		return date.toLocaleTimeString('en-US', {
			hour12: false,
			hour: '2-digit',
			minute: '2-digit'
		});
	} catch {
		return '';
	}
}

/**
 * Format a timestamp string into a readable date label.
 * Returns "Today", "Yesterday", or a formatted date string.
 */
export function formatDate(timestamp: string): string {
	if (!timestamp) return '';
	try {
		const date = new Date(timestamp);
		if (isNaN(date.getTime())) return '';

		const now = new Date();
		const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
		const msgDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());
		const diffDays = Math.floor((today.getTime() - msgDate.getTime()) / (1000 * 60 * 60 * 24));

		if (diffDays === 0) return 'Today';
		if (diffDays === 1) return 'Yesterday';
		return date.toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
		});
	} catch {
		return '';
	}
}

/**
 * Format a peer's last_seen into a human-readable relative time.
 */
export function formatLastSeen(lastSeen: string): string {
	if (!lastSeen) return '';
	const date = new Date(lastSeen);
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	const diffMinutes = Math.floor(diffMs / 1000 / 60);
	const diffHours = Math.floor(diffMinutes / 60);
	const diffDays = Math.floor(diffHours / 24);

	if (diffMinutes < 1) return 'just now';
	if (diffMinutes < 60) return `${diffMinutes}m ago`;
	if (diffHours < 24) return `${diffHours}h ago`;
	return `${diffDays}d ago`;
}

/**
 * Sort messages by sent_at timestamp (ascending — oldest first).
 * Uses millisecond precision to prevent race conditions.
 */
export function sortMessagesByTimestamp(messages: Message[]): Message[] {
	return [...messages].sort((a, b) => {
		const timeA = new Date(a.sent_at).getTime();
		const timeB = new Date(b.sent_at).getTime();
		if (timeA !== timeB) return timeA - timeB;
		// Fallback to ID for same-millisecond messages
		return a.ID - b.ID;
	});
}

/**
 * Get the initials from a username for avatar display.
 */
export function getInitials(username: string): string {
	if (!username) return '?';
	return username.charAt(0).toUpperCase();
}
