import { err, ok, Result } from 'neverthrow';
import type { Error } from '$lib/entites/error';
import type { FileTransfer } from '$lib/entites/fileTransfer';
import { writable } from 'svelte/store';

export const fileTransfers = writable<FileTransfer[]>([]);

export default class FileService {
	async sendFile(toPeerId: string, filePath: string): Promise<Result<void, Error>> {
		try {
			const response = await fetch('http://localhost:8000/api/files/send', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					to_peer_id: toPeerId,
					file_path: filePath
				})
			});

			if (response.ok) {
				return ok(undefined);
			} else {
				const error = (await response.json()) as Error;
				return err(error);
			}
		} catch (e) {
			return err({ error: 'Failed to send file: ' + String(e) });
		}
	}

	async fetchFileTransfers(peerId: string): Promise<Result<FileTransfer[], Error>> {
		try {
			const response = await fetch(`http://localhost:8000/api/files?peer_id=${peerId}`, {
				method: 'GET'
			});

			if (response.ok) {
				const transfers = (await response.json()) as FileTransfer[];
				fileTransfers.set(transfers || []);
				return ok(transfers || []);
			} else {
				const error = (await response.json()) as Error;
				return err(error);
			}
		} catch (e) {
			return err({ error: 'Failed to fetch file transfers: ' + String(e) });
		}
	}
}
