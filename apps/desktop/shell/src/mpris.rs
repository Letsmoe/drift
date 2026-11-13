//! MPRIS: the player as seen by media keys, desktop widgets and `playerctl`, on the session D-Bus as
//! `org.mpris.MediaPlayer2.drift`.
//!
//! The UI owns the player, so this only mirrors it. The UI pushes the song and the playback state with
//! `mpris_update`; changes go out as `PropertiesChanged` (and `Seeked` after a seek). What other programs
//! ask for (play, pause, next, seek, volume) comes back to the UI as `mpris` events, which the player acts
//! on; it then pushes the new state. There is no `Raise`: the shell has no handle on the window.

use std::sync::mpsc::{self, Receiver, Sender};
use std::sync::{Arc, Mutex, MutexGuard};
use std::thread;
use std::time::Instant;

use mpris_server::zbus::{self, fdo};
use mpris_server::{
    LoopStatus, Metadata, PlaybackRate, PlaybackStatus, PlayerInterface, Property, RootInterface, Server, Signal, Time,
    TrackId, Volume,
};
use serde::Deserialize;
use serde_json::{Value, json};
use svelte_gpui::Emitter;

const EVENT_NAME: &str = "mpris";
const BUS_NAME_SUFFIX: &str = "drift";
const TRACK_PATH_PREFIX: &str = "/org/drift/track/";

#[derive(Deserialize, Clone, PartialEq)]
#[serde(rename_all = "camelCase")]
pub struct TrackInfo {
    /// Names the entry in the queue; the same song twice in the queue is two tracks.
    track_id: String,
    title: String,
    artists: Vec<String>,
    album: String,
    cover_url: String,
    length_seconds: f64,
}

#[derive(Deserialize, Clone, Copy, PartialEq)]
#[serde(rename_all = "camelCase")]
enum Status {
    Playing,
    Paused,
    Stopped,
}

/// Everything the UI tells about its player. `position_seconds` is where it was when the update was sent.
#[derive(Deserialize, Clone)]
#[serde(rename_all = "camelCase")]
pub struct Update {
    track: Option<TrackInfo>,
    status: Status,
    position_seconds: f64,
    volume: f64,
    can_go_next: bool,
    can_go_previous: bool,
    /// The position changed by a seek, which MPRIS clients are told about with `Seeked`.
    #[serde(default)]
    seeked: bool,
}

impl Update {
    fn nothing_playing() -> Update {
        Update {
            track: None,
            status: Status::Stopped,
            position_seconds: 0.0,
            volume: 1.0,
            can_go_next: false,
            can_go_previous: false,
            seeked: false,
        }
    }
}

/// What the player interface answers from, and where the position was measured.
struct Shared {
    state: Mutex<(Update, Instant)>,
    emitter: Emitter,
}

enum Change {
    Properties(Vec<Property>),
    Seeked(Time),
}

pub struct Mpris {
    shared: Arc<Shared>,
    changes: Sender<Change>,
}

impl Mpris {
    /// Claims the bus name on a thread of its own. Without a session bus the player just goes unseen.
    pub fn start(emitter: Emitter) -> Mpris {
        let shared = Arc::new(Shared { state: Mutex::new((Update::nothing_playing(), Instant::now())), emitter });
        let (changes, receiver) = mpsc::channel();
        let served = Arc::clone(&shared);
        let spawned = thread::Builder::new().name("mpris".to_string()).spawn(move || serve(served, receiver));
        if let Err(error) = spawned {
            eprintln!("drift: could not start MPRIS: {error}");
        }
        Mpris { shared, changes }
    }

    pub fn update(&self, update: Update) {
        for change in self.shared.apply(update) {
            // The server is gone when it could not start; nobody is left to tell.
            let _ = self.changes.send(change);
        }
    }
}

impl Shared {
    fn lock(&self) -> MutexGuard<'_, (Update, Instant)> {
        self.state.lock().unwrap_or_else(|poisoned| poisoned.into_inner())
    }

    /// Takes the update over and says what the bus has to hear about it.
    fn apply(&self, update: Update) -> Vec<Change> {
        let mut state = self.lock();
        let mut changes = Vec::new();
        let properties = changed_properties(&state.0, &update);
        if !properties.is_empty() {
            changes.push(Change::Properties(properties));
        }
        if update.seeked {
            changes.push(Change::Seeked(time_of(update.position_seconds)));
        }
        *state = (update, Instant::now());
        changes
    }

    fn request(&self, request: Value) {
        self.emitter.emit(EVENT_NAME, request);
    }

    fn position(&self) -> Time {
        let state = self.lock();
        let mut seconds = state.0.position_seconds;
        if state.0.status == Status::Playing {
            seconds += state.1.elapsed().as_secs_f64();
        }
        time_of(seconds)
    }

    fn current_track_id(&self) -> Option<TrackId> {
        self.lock().0.track.as_ref().map(track_id_of)
    }
}

fn time_of(seconds: f64) -> Time {
    Time::from_micros((seconds.max(0.0) * 1_000_000.0) as i64)
}

/// The D-Bus object path of a track: letters and digits stay, anything else becomes `_` and its hex code.
fn track_id_of(track: &TrackInfo) -> TrackId {
    let mut path = TRACK_PATH_PREFIX.to_string();
    for character in track.track_id.chars() {
        if character.is_ascii_alphanumeric() {
            path.push(character);
        } else {
            path.push_str(&format!("_{:x}", u32::from(character)));
        }
    }
    TrackId::try_from(path).unwrap_or(TrackId::NO_TRACK)
}

fn metadata_of(track: Option<&TrackInfo>) -> Metadata {
    let Some(track) = track else {
        return Metadata::builder().trackid(TrackId::NO_TRACK).build();
    };
    let mut metadata = Metadata::builder().trackid(track_id_of(track)).title(track.title.clone()).build();
    if !track.artists.is_empty() {
        metadata.set_artist(Some(track.artists.clone()));
    }
    if !track.album.is_empty() {
        metadata.set_album(Some(track.album.clone()));
    }
    if !track.cover_url.is_empty() {
        metadata.set_art_url(Some(track.cover_url.clone()));
    }
    if track.length_seconds > 0.0 {
        metadata.set_length(Some(time_of(track.length_seconds)));
    }
    metadata
}

fn playback_status_of(status: Status) -> PlaybackStatus {
    match status {
        Status::Playing => PlaybackStatus::Playing,
        Status::Paused => PlaybackStatus::Paused,
        Status::Stopped => PlaybackStatus::Stopped,
    }
}

/// The properties whose value differs between two updates. A position is not one of them: clients ask for it.
fn changed_properties(old: &Update, new: &Update) -> Vec<Property> {
    let mut properties = Vec::new();
    if old.track != new.track {
        properties.push(Property::Metadata(metadata_of(new.track.as_ref())));
    }
    if old.track.is_some() != new.track.is_some() {
        properties.push(Property::CanPlay(new.track.is_some()));
        properties.push(Property::CanPause(new.track.is_some()));
        properties.push(Property::CanSeek(new.track.is_some()));
    }
    if old.status != new.status {
        properties.push(Property::PlaybackStatus(playback_status_of(new.status)));
    }
    if old.volume != new.volume {
        properties.push(Property::Volume(new.volume));
    }
    if old.can_go_next != new.can_go_next {
        properties.push(Property::CanGoNext(new.can_go_next));
    }
    if old.can_go_previous != new.can_go_previous {
        properties.push(Property::CanGoPrevious(new.can_go_previous));
    }
    properties
}

/// Owns the bus name and forwards changes to the bus until the shell exits.
fn serve(shared: Arc<Shared>, changes: Receiver<Change>) {
    let server = match zbus::block_on(claim_bus_name(&shared)) {
        Ok(server) => server,
        Err(error) => {
            eprintln!("drift: MPRIS is not available: {error}");
            return;
        }
    };
    for change in changes {
        if let Err(error) = zbus::block_on(publish(&server, change)) {
            eprintln!("drift: MPRIS could not send a change: {error}");
        }
    }
}

/// `drift` when it is free; another running instance holds it, so this one is `drift.instance<pid>`.
async fn claim_bus_name(shared: &Arc<Shared>) -> zbus::Result<Server<MprisPlayer>> {
    let first = Server::new(BUS_NAME_SUFFIX, MprisPlayer { shared: Arc::clone(shared) }).await;
    if first.is_ok() {
        return first;
    }
    let suffix = format!("{BUS_NAME_SUFFIX}.instance{}", std::process::id());
    Server::new(&suffix, MprisPlayer { shared: Arc::clone(shared) }).await
}

async fn publish(server: &Server<MprisPlayer>, change: Change) -> zbus::Result<()> {
    match change {
        Change::Properties(properties) => server.properties_changed(properties).await,
        Change::Seeked(position) => server.emit(Signal::Seeked { position }).await,
    }
}

struct MprisPlayer {
    shared: Arc<Shared>,
}

impl MprisPlayer {
    fn request(&self, kind: &str) -> fdo::Result<()> {
        self.shared.request(json!({ "kind": kind }));
        Ok(())
    }

    fn has_track(&self) -> bool {
        self.shared.lock().0.track.is_some()
    }
}

impl RootInterface for MprisPlayer {
    async fn raise(&self) -> fdo::Result<()> {
        Ok(())
    }

    async fn quit(&self) -> fdo::Result<()> {
        Ok(())
    }

    async fn can_quit(&self) -> fdo::Result<bool> {
        Ok(false)
    }

    async fn fullscreen(&self) -> fdo::Result<bool> {
        Ok(false)
    }

    async fn set_fullscreen(&self, _fullscreen: bool) -> zbus::Result<()> {
        Ok(())
    }

    async fn can_set_fullscreen(&self) -> fdo::Result<bool> {
        Ok(false)
    }

    async fn can_raise(&self) -> fdo::Result<bool> {
        Ok(false)
    }

    async fn has_track_list(&self) -> fdo::Result<bool> {
        Ok(false)
    }

    async fn identity(&self) -> fdo::Result<String> {
        Ok("Drift".to_string())
    }

    async fn desktop_entry(&self) -> fdo::Result<String> {
        Ok("drift".to_string())
    }

    async fn supported_uri_schemes(&self) -> fdo::Result<Vec<String>> {
        Ok(Vec::new())
    }

    async fn supported_mime_types(&self) -> fdo::Result<Vec<String>> {
        Ok(Vec::new())
    }
}

impl PlayerInterface for MprisPlayer {
    async fn next(&self) -> fdo::Result<()> {
        self.request("next")
    }

    async fn previous(&self) -> fdo::Result<()> {
        self.request("previous")
    }

    async fn pause(&self) -> fdo::Result<()> {
        self.request("pause")
    }

    async fn play_pause(&self) -> fdo::Result<()> {
        self.request("playPause")
    }

    async fn stop(&self) -> fdo::Result<()> {
        self.request("stop")
    }

    async fn play(&self) -> fdo::Result<()> {
        self.request("play")
    }

    async fn seek(&self, offset: Time) -> fdo::Result<()> {
        let offset_seconds = offset.as_micros() as f64 / 1_000_000.0;
        self.shared.request(json!({ "kind": "seek", "offsetSeconds": offset_seconds }));
        Ok(())
    }

    // A position for a song that is no longer the current one is ignored, as the specification asks.
    async fn set_position(&self, track_id: TrackId, position: Time) -> fdo::Result<()> {
        if self.shared.current_track_id() != Some(track_id) {
            return Ok(());
        }
        let position_seconds = position.as_micros() as f64 / 1_000_000.0;
        self.shared.request(json!({ "kind": "setPosition", "positionSeconds": position_seconds }));
        Ok(())
    }

    async fn open_uri(&self, _uri: String) -> fdo::Result<()> {
        Ok(())
    }

    async fn playback_status(&self) -> fdo::Result<PlaybackStatus> {
        Ok(playback_status_of(self.shared.lock().0.status))
    }

    async fn loop_status(&self) -> fdo::Result<LoopStatus> {
        Ok(LoopStatus::None)
    }

    async fn set_loop_status(&self, _loop_status: LoopStatus) -> zbus::Result<()> {
        Ok(())
    }

    async fn rate(&self) -> fdo::Result<PlaybackRate> {
        Ok(PlaybackRate::default())
    }

    async fn set_rate(&self, _rate: PlaybackRate) -> zbus::Result<()> {
        Ok(())
    }

    async fn shuffle(&self) -> fdo::Result<bool> {
        Ok(false)
    }

    async fn set_shuffle(&self, _shuffle: bool) -> zbus::Result<()> {
        Ok(())
    }

    async fn metadata(&self) -> fdo::Result<Metadata> {
        Ok(metadata_of(self.shared.lock().0.track.as_ref()))
    }

    async fn volume(&self) -> fdo::Result<Volume> {
        Ok(self.shared.lock().0.volume)
    }

    async fn set_volume(&self, volume: Volume) -> zbus::Result<()> {
        self.shared.request(json!({ "kind": "setVolume", "volume": volume.clamp(0.0, 1.0) }));
        Ok(())
    }

    async fn position(&self) -> fdo::Result<Time> {
        Ok(self.shared.position())
    }

    async fn minimum_rate(&self) -> fdo::Result<PlaybackRate> {
        Ok(PlaybackRate::default())
    }

    async fn maximum_rate(&self) -> fdo::Result<PlaybackRate> {
        Ok(PlaybackRate::default())
    }

    async fn can_go_next(&self) -> fdo::Result<bool> {
        Ok(self.shared.lock().0.can_go_next)
    }

    async fn can_go_previous(&self) -> fdo::Result<bool> {
        Ok(self.shared.lock().0.can_go_previous)
    }

    async fn can_play(&self) -> fdo::Result<bool> {
        Ok(self.has_track())
    }

    async fn can_pause(&self) -> fdo::Result<bool> {
        Ok(self.has_track())
    }

    async fn can_seek(&self) -> fdo::Result<bool> {
        Ok(self.has_track())
    }

    async fn can_control(&self) -> fdo::Result<bool> {
        Ok(true)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn song(track_id: &str) -> TrackInfo {
        TrackInfo {
            track_id: track_id.to_string(),
            title: "Title".to_string(),
            artists: vec!["Artist".to_string()],
            album: String::new(),
            cover_url: String::new(),
            length_seconds: 200.0,
        }
    }

    #[test]
    fn a_track_id_is_a_valid_object_path() {
        let id = track_id_of(&song("12-ab"));
        assert_eq!(id.as_str(), "/org/drift/track/12_2dab");
    }

    #[test]
    fn nothing_changed_tells_the_bus_nothing() {
        let update = Update::nothing_playing();
        assert!(changed_properties(&update, &update).is_empty());
    }

    #[test]
    fn a_new_song_changes_metadata_and_what_can_be_done() {
        let old = Update::nothing_playing();
        let new = Update { track: Some(song("1")), status: Status::Playing, ..Update::nothing_playing() };
        let properties = changed_properties(&old, &new);
        let has = |wanted: fn(&Property) -> bool| properties.iter().any(wanted);
        assert!(has(|property| matches!(property, Property::Metadata(_))));
        assert!(has(|property| matches!(property, Property::CanPlay(true))));
        assert!(has(|property| matches!(property, Property::PlaybackStatus(PlaybackStatus::Playing))));
        assert!(!has(|property| matches!(property, Property::Volume(_))));
    }
}
