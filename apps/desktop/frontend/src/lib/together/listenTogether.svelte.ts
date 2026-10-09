import {
	acceptJoin,
	addFriend,
	declineJoin,
	dismissInvite,
	hostRoom,
	inviteFriend,
	joinRoom,
	leaveRoom,
	onRemotePlayback,
	onTogetherChange,
	removeFriend,
	reportPlayback,
	setFriendTrusted,
	setSharePresence,
	setTogetherName,
	togetherState,
	type NowPlaying,
	type RoomStatus,
	type TogetherFriend,
	type TogetherInvite,
	type TogetherPeer,
	type TogetherRoom,
	type TogetherState
} from '#lib/api/together.ts';
import { describeError } from '#lib/format.ts';
import { player } from '#lib/player/player.svelte.ts';

/** How often the player reports its position; the host's report keeps guests in step and friends see it. */
const positionReportIntervalMs = 5_000;
/** How often live positions of friends' songs move on. */
const clockIntervalMs = 1_000;
/** Coalesces the stream of seeks while the slider is dragged into one report. */
const userActionReportDelayMs = 150;
const userIdGroupLength = 4;

/** Groups a user ID for reading: "K7QD-9MZA-X2PT". */
export function formatUserId(userId: string): string {
	const groups: string[] = [];
	for (let start = 0; start < userId.length; start += userIdGroupLength) {
		groups.push(userId.slice(start, start + userIdGroupLength));
	}
	return groups.join('-');
}

function capitalize(message: string): string {
	return message.charAt(0).toUpperCase() + message.slice(1);
}

/**
 * Where a friend is in their song now: the shared position moved on by the time since, while playing.
 * `nowMs` is a reactive clock, so callers re-render every second.
 */
export function livePositionSeconds(nowPlaying: NowPlaying, nowMs: number): number {
	if (nowPlaying.paused) {
		return nowPlaying.positionSeconds;
	}
	const elapsedSeconds = Math.max(0, nowMs - Date.parse(nowPlaying.updatedAt)) / 1000;
	const position = nowPlaying.positionSeconds + elapsedSeconds;
	const duration = nowPlaying.track.durationSeconds;
	if (duration > 0) {
		return Math.min(position, duration);
	}
	return position;
}

/** Listen Together: the user's ID, friends, invites and the open room, mirrored from the Go core. */
export class ListenTogether {
	state = $state<TogetherState | null>(null);
	/** Ticks every second for live song positions. */
	nowMs = $state(Date.now());
	errorMessage = $state('');
	/** The action in flight, e.g. `invite:K7QD9MZAX2PT`, so its button can show progress. */
	pendingAction = $state('');
	/** The friend whose invite went out last, for a short confirmation. */
	invitedUserId = $state('');

	#started = false;
	#reportTimer: ReturnType<typeof setTimeout> | undefined;

	get room(): TogetherRoom | null {
		if (!this.state) {
			return null;
		}
		return this.state.room;
	}

	get roomStatus(): RoomStatus | null {
		if (!this.room) {
			return null;
		}
		return this.room.status as RoomStatus;
	}

	/** True while the room is open, connected or not yet. */
	get roomOpen(): boolean {
		const openStatuses: (RoomStatus | null)[] = ['waiting', 'connecting', 'awaiting', 'connected'];
		return openStatuses.includes(this.roomStatus);
	}

	/** Guests waiting for the user to let them in. */
	get requests(): TogetherPeer[] {
		const room = this.room;
		if (!room || !room.hosting) {
			return [];
		}
		return room.requests;
	}

	/** Invites and join requests, which the header button points out. */
	get attentionCount(): number {
		return this.invites.length + this.requests.length;
	}

	get sharePresence(): boolean {
		if (!this.state) {
			return false;
		}
		return this.state.sharePresence;
	}

	get connected(): boolean {
		return this.roomStatus === 'connected';
	}

	get friends(): TogetherFriend[] {
		if (!this.state) {
			return [];
		}
		return this.state.friends;
	}

	get invites(): TogetherInvite[] {
		if (!this.state) {
			return [];
		}
		return this.state.invites;
	}

	/** Everyone in the room but the user. */
	get otherPeers(): TogetherPeer[] {
		const room = this.room;
		if (!room) {
			return [];
		}
		return room.peers.filter((peer) => peer.id !== room.selfPeerId);
	}

	/** The name of the room's host, as a friend or as they call themselves. */
	get hostName(): string {
		const room = this.room;
		if (!room) {
			return '';
		}
		const host = room.peers.find((peer) => peer.userId === room.hostUserId);
		if (host) {
			return host.name;
		}
		return this.nameOf(room.hostUserId);
	}

	isFriend(userId: string): boolean {
		return this.friends.some((friend) => friend.userId === userId);
	}

	/** A friend's name, else their formatted ID. */
	nameOf(userId: string): string {
		const friend = this.friends.find((candidate) => candidate.userId === userId);
		if (friend && friend.name !== '') {
			return friend.name;
		}
		return formatUserId(userId);
	}

	isPending(action: string): boolean {
		return this.pendingAction === action;
	}

	/** Subscribes to the core and links the room to the player; runs once. */
	start(): void {
		if (this.#started) {
			return;
		}
		this.#started = true;
		onTogetherChange((state) => (this.state = state));
		onRemotePlayback((playback) => player.followRemote(playback));
		player.onUserAction(() => this.#scheduleUserActionReport());
		setInterval(() => this.#reportPosition(), positionReportIntervalMs);
		setInterval(() => (this.nowMs = Date.now()), clockIntervalMs);
		togetherState().then(
			(state) => (this.state = state),
			(error) => (this.errorMessage = capitalize(describeError(error)))
		);
	}

	host(): Promise<boolean> {
		return this.#run('host', () => hostRoom());
	}

	join(hostUserId: string): Promise<boolean> {
		return this.#run(`join:${hostUserId}`, () => joinRoom(hostUserId));
	}

	leave(): Promise<boolean> {
		return this.#run('leave', () => leaveRoom());
	}

	async invite(userId: string): Promise<boolean> {
		const sent = await this.#run(`invite:${userId}`, () => inviteFriend(userId));
		if (sent) {
			this.invitedUserId = userId;
		}
		return sent;
	}

	dismiss(userId: string): Promise<boolean> {
		return this.#run(`dismiss:${userId}`, () => dismissInvite(userId));
	}

	addFriend(userId: string, name: string): Promise<boolean> {
		return this.#run('add-friend', () => addFriend(userId, name));
	}

	removeFriend(userId: string): Promise<boolean> {
		return this.#run(`remove:${userId}`, () => removeFriend(userId));
	}

	setTrusted(userId: string, trusted: boolean): Promise<boolean> {
		return this.#run(`trust:${userId}`, () => setFriendTrusted(userId, trusted));
	}

	setSharePresence(share: boolean): Promise<boolean> {
		return this.#run('share-presence', () => setSharePresence(share));
	}

	/** Lets a waiting guest in; `trust` also lets them in without asking from now on. */
	acceptJoin(peerId: string, trust: boolean): Promise<boolean> {
		return this.#run(`accept:${peerId}`, () => acceptJoin(peerId, trust));
	}

	declineJoin(peerId: string): Promise<boolean> {
		return this.#run(`decline:${peerId}`, () => declineJoin(peerId));
	}

	rename(name: string): Promise<boolean> {
		return this.#run('rename', () => setTogetherName(name));
	}

	clearError(): void {
		this.errorMessage = '';
	}

	/** Runs an action, keeping its key pending; returns true when it worked. */
	async #run(action: string, request: () => Promise<void>): Promise<boolean> {
		this.errorMessage = '';
		this.invitedUserId = '';
		this.pendingAction = action;
		try {
			await request();
			return true;
		} catch (error) {
			this.errorMessage = capitalize(describeError(error));
			return false;
		} finally {
			if (this.pendingAction === action) {
				this.pendingAction = '';
			}
		}
	}

	#scheduleUserActionReport(): void {
		if (this.#reportTimer !== undefined) {
			return;
		}
		this.#reportTimer = setTimeout(() => {
			this.#reportTimer = undefined;
			this.#report(true);
		}, userActionReportDelayMs);
	}

	// Reported outside rooms too: friends see the song and position, and a joining guest starts in step.
	#reportPosition(): void {
		this.#report(false);
	}

	#report(userAction: boolean): void {
		const entry = player.entry;
		if (!entry) {
			return;
		}
		reportPlayback({
			entryId: entry.id,
			paused: player.paused,
			positionSeconds: player.currentTime,
			userAction
		}).catch((error) => console.warn('reportPlayback failed', error));
	}
}

export const listenTogether = new ListenTogether();
