import {
	TogetherAcceptJoin,
	TogetherAddFriend,
	TogetherDeclineJoin,
	TogetherDismissInvite,
	TogetherHost,
	TogetherInvite,
	TogetherJoin,
	TogetherLeave,
	TogetherRemoveFriend,
	TogetherReportPlayback,
	TogetherSetFriendTrusted,
	TogetherSetName,
	TogetherSetSharePresence,
	TogetherState
} from '#lib/backend/generated/App.ts';
import { onEvent } from '#lib/backend/rpc.ts';
import type { together } from '#lib/backend/generated/models.ts';

export type TogetherState = together.State;
export type TogetherRoom = together.Room;
export type TogetherPeer = together.Peer;
export type TogetherFriend = together.Friend;
export type TogetherInvite = together.Invite;
export type FriendPresence = together.Presence;
export type NowPlaying = together.NowPlaying;
export type PlaybackReport = together.PlaybackReport;
export type RoomStatus =
	'waiting' | 'connecting' | 'awaiting' | 'connected' | 'closed' | 'failed' | 'declined';

/** Someone else's play, pause or seek, or the host's periodic position (core together.Playback). */
export interface RemotePlayback {
	entryId: number;
	paused: boolean;
	positionSeconds: number;
	/** How far the local position may be off before seeking. */
	toleranceSeconds: number;
	/** Who played, paused or seeked; empty for the host's periodic position. */
	from: string;
}

const stateEventName = 'together';
const playbackEventName = 'together-playback';

export function togetherState(): Promise<TogetherState> {
	return TogetherState();
}

/** Subscribes to every change of the identity, friends, invites and room. Returns the unsubscribe function. */
export function onTogetherChange(callback: (state: TogetherState) => void): () => void {
	return onEvent(stateEventName, callback);
}

/** Subscribes to what the local player should follow. Returns the unsubscribe function. */
export function onRemotePlayback(callback: (playback: RemotePlayback) => void): () => void {
	return onEvent(playbackEventName, callback);
}

export function setTogetherName(name: string): Promise<void> {
	return TogetherSetName(name);
}

/** Adds a friend by the ID they shared; the name may be empty. */
export function addFriend(userId: string, name: string): Promise<void> {
	return TogetherAddFriend(userId, name);
}

export function removeFriend(userId: string): Promise<void> {
	return TogetherRemoveFriend(userId);
}

/** Trusted friends join the user's sessions without being asked for. */
export function setFriendTrusted(userId: string, trusted: boolean): Promise<void> {
	return TogetherSetFriendTrusted(userId, trusted);
}

/** Whether friends see the song the user plays and where in it they are. */
export function setSharePresence(share: boolean): Promise<void> {
	return TogetherSetSharePresence(share);
}

/** Lets a guest waiting at the user's session in; with trust they never ask again. */
export function acceptJoin(peerId: string, trust: boolean): Promise<void> {
	return TogetherAcceptJoin(peerId, trust);
}

export function declineJoin(peerId: string): Promise<void> {
	return TogetherDeclineJoin(peerId);
}

/** Opens the user's own room, which friends join by the user's ID. */
export function hostRoom(): Promise<void> {
	return TogetherHost();
}

export function joinRoom(hostUserId: string): Promise<void> {
	return TogetherJoin(hostUserId);
}

export function leaveRoom(): Promise<void> {
	return TogetherLeave();
}

/** Opens the user's room when needed and invites a friend; resolves once a relay took the invite. */
export function inviteFriend(userId: string): Promise<void> {
	return TogetherInvite(userId);
}

export function dismissInvite(userId: string): Promise<void> {
	return TogetherDismissInvite(userId);
}

export function reportPlayback(report: PlaybackReport): Promise<void> {
	return TogetherReportPlayback(report);
}
