package main

import (
	"errors"
	"os/user"
	"strings"

	"github.com/letsmoe/drift/core/together"
)

const (
	// togetherEvent carries every new together.State.
	togetherEvent = "together"
	// togetherPlaybackEvent asks the player to follow someone else's play, pause or seek.
	togetherPlaybackEvent = "together-playback"
)

var errTogetherUnavailable = errors.New("listen together needs the library, which could not be opened")

// openTogether loads the Listen Together identity and friends; it needs the library.
func (app *App) openTogether() {
	if app.library == nil {
		return
	}
	opened, err := together.New(app.ctx, app.library, app.queue,
		together.WithLogger(app.logger),
		together.WithDefaultName(defaultDisplayName()),
		together.WithOnState(func(state together.State) { app.emit(togetherEvent, state) }),
		together.WithOnPlayback(func(playback together.Playback) { app.emit(togetherPlaybackEvent, playback) }),
	)
	if err != nil {
		app.logger.Error("listen together failed to open", "error", err)
		return
	}
	app.together = opened
}

// defaultDisplayName is the account's full name, else the login name.
func defaultDisplayName() string {
	current, err := user.Current()
	if err != nil {
		return ""
	}
	// GECOS fields list room and phone after the name, separated by commas.
	fullName, _, _ := strings.Cut(current.Name, ",")
	if strings.TrimSpace(fullName) != "" {
		return fullName
	}
	return current.Username
}

func (app *App) closeTogether() {
	if app.together != nil {
		app.together.Close()
	}
}

// following reports whether the user is a guest in someone's room.
func (app *App) following() bool {
	return app.together != nil && app.together.Following()
}

// TogetherState returns the user's ID and name, friends, invites and the open room.
func (app *App) TogetherState() (together.State, error) {
	if app.together == nil {
		return together.State{}, errTogetherUnavailable
	}
	return app.together.State(), nil
}

// TogetherSetName changes the name others see.
func (app *App) TogetherSetName(name string) error {
	if app.together == nil {
		return errTogetherUnavailable
	}
	return app.together.SetName(app.ctx, name)
}

// TogetherAddFriend adds a friend by the ID they shared; name may be empty.
func (app *App) TogetherAddFriend(userID, name string) error {
	if app.together == nil {
		return errTogetherUnavailable
	}
	return app.together.AddFriend(app.ctx, userID, name)
}

// TogetherRemoveFriend removes a friend.
func (app *App) TogetherRemoveFriend(userID string) error {
	if app.together == nil {
		return errTogetherUnavailable
	}
	return app.together.RemoveFriend(app.ctx, userID)
}

// TogetherHost opens the user's own room for friends to join.
func (app *App) TogetherHost() error {
	if app.together == nil {
		return errTogetherUnavailable
	}
	return app.together.Host()
}

// TogetherJoin joins the room of the person with userID.
func (app *App) TogetherJoin(userID string) error {
	if app.together == nil {
		return errTogetherUnavailable
	}
	return app.together.Join(userID)
}

// TogetherLeave leaves the open room.
func (app *App) TogetherLeave() {
	if app.together != nil {
		app.together.Leave()
	}
}

// TogetherInvite hosts the user's room when needed and invites a friend into it.
func (app *App) TogetherInvite(userID string) error {
	if app.together == nil {
		return errTogetherUnavailable
	}
	return app.together.Invite(userID)
}

// TogetherDismissInvite drops the invite from userID.
func (app *App) TogetherDismissInvite(userID string) {
	if app.together != nil {
		app.together.DismissInvite(userID)
	}
}

// TogetherSetSharePresence turns sharing the playing song with friends on or off.
func (app *App) TogetherSetSharePresence(share bool) error {
	if app.together == nil {
		return errTogetherUnavailable
	}
	return app.together.SetSharePresence(app.ctx, share)
}

// TogetherSetFriendTrusted decides whether a friend joins the user's sessions without asking.
func (app *App) TogetherSetFriendTrusted(userID string, trusted bool) error {
	if app.together == nil {
		return errTogetherUnavailable
	}
	return app.together.SetFriendTrusted(app.ctx, userID, trusted)
}

// TogetherAcceptJoin lets a waiting guest in; with trust they never need to ask again.
func (app *App) TogetherAcceptJoin(peerID string, trust bool) error {
	if app.together == nil {
		return errTogetherUnavailable
	}
	return app.together.AcceptJoin(app.ctx, peerID, trust)
}

// TogetherDeclineJoin turns a waiting guest away.
func (app *App) TogetherDeclineJoin(peerID string) {
	if app.together != nil {
		app.together.DeclineJoin(peerID)
	}
}

// TogetherReportPlayback passes the player's state to the room and to friends:
// on play, pause and seek by the user, and periodically.
func (app *App) TogetherReportPlayback(report together.PlaybackReport) {
	if app.together != nil {
		app.together.ReportPlayback(report)
	}
}
