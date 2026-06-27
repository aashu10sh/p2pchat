import type { Message, Chat } from '$lib/entites/message';
import type { Error } from '$lib/entites/error';
import { err, ok, Result } from 'neverthrow';
import { writable } from 'svelte/store';
import { sortMessagesByTimestamp } from '$lib/utils';

export const messages = writable<Message[]>([]);
export const chats = writable<Chat[]>([]);

export default class ChatService {
	async fetchChats(): Promise<Result<Chat[], Error>> {
		try {
			const response = await fetch('http://localhost:8000/api/chats', {
				method: 'GET'
			});

			if (response.ok) {
				const chatList = (await response.json()) as Chat[];
				chats.set(chatList || []);
				return ok(chatList || []);
			} else {
				const error = (await response.json()) as Error;
				return err(error);
			}
		} catch (e) {
			return err({
				error: 'Failed to fetch chats: ' + String(e)
			});
		}
	}

	async fetchMessages(peerID: string, limit: number = 50): Promise<Result<Message[], Error>> {
		try {
			const response = await fetch(
				`http://localhost:8000/api/messages?peer_id=${peerID}&limit=${limit}`,
				{
					method: 'GET'
				}
			);

			if (response.ok) {
				const messageList = (await response.json()) as Message[];
				const sorted = sortMessagesByTimestamp(messageList || []);
				messages.set(sorted);
				return ok(sorted);
			} else {
				const error = (await response.json()) as Error;
				return err(error);
			}
		} catch (e) {
			return err({
				error: 'Failed to fetch messages: ' + String(e)
			});
		}
	}

	async sendMessage(
		toPeerID: string,
		content: string,
		messageType: string = 'TEXT'
	): Promise<Result<{ message_id: string; status: string }, Error>> {
		try {
			const response = await fetch('http://localhost:8000/api/messages', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					to_peer_id: toPeerID,
					content: content,
					message_type: messageType
				})
			});

			if (response.ok) {
				const result = (await response.json()) as { message_id: string; status: string };
				return ok(result);
			} else {
				const error = (await response.json()) as Error;
				return err(error);
			}
		} catch (e) {
			return err({
				error: 'Failed to send message: ' + String(e)
			});
		}
	}
}
