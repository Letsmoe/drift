import { backendAddress } from '#lib/backend/shell.ts';

// Wire format: see apps/desktop/hub.go.
interface ResponseMessage {
	id: number;
	result?: unknown;
	error?: string;
}

interface EventMessage {
	event: string;
	data: unknown;
}

interface PendingCall {
	resolve: (value: unknown) => void;
	reject: (error: Error) => void;
}

type EventListener = (data: unknown) => void;

const reconnectDelayMs = 1000;

class RpcClient {
	#socket: Promise<WebSocket> | undefined;
	#nextId = 1;
	#pending = new Map<number, PendingCall>();
	#listeners = new Map<string, Set<EventListener>>();

	async call<Result>(method: string, params: unknown[]): Promise<Result> {
		const socket = await this.#connect();
		const id = this.#nextId++;
		const result = new Promise<unknown>((resolve, reject) => {
			this.#pending.set(id, { resolve, reject });
		});
		socket.send(JSON.stringify({ id, method, params }));
		return result as Promise<Result>;
	}

	/** Listens to a backend event; returns the unsubscribe function. */
	on(event: string, listener: EventListener): () => void {
		let listeners = this.#listeners.get(event);
		if (!listeners) {
			listeners = new Set();
			this.#listeners.set(event, listeners);
		}
		listeners.add(listener);
		this.#connect().catch((error) => console.warn('backend connection failed', error));
		return () => listeners.delete(listener);
	}

	#connect(): Promise<WebSocket> {
		if (!this.#socket) {
			this.#socket = this.#open();
		}
		return this.#socket;
	}

	#open(): Promise<WebSocket> {
		const address = backendAddress();
		const url = new URL('/rpc', address.url);
		url.protocol = url.protocol.replace('http', 'ws');
		url.searchParams.set('token', address.token);

		const socket = new WebSocket(url);
		socket.onmessage = (message) => this.#receive(message.data as string);
		return new Promise((resolve, reject) => {
			socket.onopen = () => resolve(socket);
			socket.onclose = () => {
				reject(new Error('Could not reach the backend.'));
				this.#closed();
			};
		});
	}

	#receive(text: string): void {
		const message = JSON.parse(text) as ResponseMessage | EventMessage;
		if ('event' in message) {
			this.#dispatch(message);
			return;
		}
		this.#settle(message);
	}

	#dispatch(message: EventMessage): void {
		const listeners = this.#listeners.get(message.event);
		if (!listeners) {
			return;
		}
		for (const listener of listeners) {
			listener(message.data);
		}
	}

	#settle(message: ResponseMessage): void {
		const pending = this.#pending.get(message.id);
		if (!pending) {
			return;
		}
		this.#pending.delete(message.id);
		if (message.error !== undefined) {
			pending.reject(new Error(message.error));
			return;
		}
		pending.resolve(message.result);
	}

	// Fails the calls in flight; reconnects right away only if events are wanted.
	#closed(): void {
		this.#socket = undefined;
		for (const pending of this.#pending.values()) {
			pending.reject(new Error('The backend connection closed.'));
		}
		this.#pending.clear();
		if (this.#listeners.size > 0) {
			setTimeout(() => this.#connect().catch(() => {}), reconnectDelayMs);
		}
	}
}

const client = new RpcClient();

/** Calls an exported method of the Go App; used by the generated bindings. */
export function call<Result>(method: string, params: unknown[]): Promise<Result> {
	return client.call<Result>(method, params);
}

/** Subscribes to a backend event such as `queue` or `downloads`. */
export function onEvent<Data>(event: string, listener: (data: Data) => void): () => void {
	return client.on(event, listener as EventListener);
}
