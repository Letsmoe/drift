import { describeError } from '#lib/format.ts';

export type ResourceStatus = 'loading' | 'done' | 'failed';

/** One page's data, fetched by id; a newer load discards the result of an older one. */
export class RemoteResource<Data> {
	data = $state<Data | undefined>(undefined);
	status = $state<ResourceStatus>('loading');
	errorMessage = $state('');

	#fetcher: (id: string) => Promise<Data>;
	#currentId = '';
	#latestRequestId = 0;

	constructor(fetcher: (id: string) => Promise<Data>) {
		this.#fetcher = fetcher;
	}

	async load(id: string): Promise<void> {
		this.#currentId = id;
		this.data = undefined;
		this.status = 'loading';
		await this.#fetch(id);
	}

	/** Fetches the current id again without showing the loading state. */
	async reload(): Promise<void> {
		await this.#fetch(this.#currentId);
	}

	async #fetch(id: string): Promise<void> {
		this.#latestRequestId += 1;
		const requestId = this.#latestRequestId;
		try {
			const data = await this.#fetcher(id);
			if (requestId === this.#latestRequestId) {
				this.data = data;
				this.status = 'done';
			}
		} catch (error) {
			if (requestId === this.#latestRequestId) {
				this.errorMessage = describeError(error);
				this.status = 'failed';
			}
		}
	}
}
