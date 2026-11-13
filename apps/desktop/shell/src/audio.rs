//! The audio engine: GStreamer `playbin` pipelines that play the backend's `/stream/{videoId}` URLs.
//!
//! There are two slots. Slot 0 plays the song; slot 1 exists for the crossfade, where it plays the
//! ending song while slot 0 moves on to the next one. Each slot reports what happens to it as
//! `audio` events: position (a few times a second, only while playing), duration, playing or
//! paused, buffering, ended and error.
//!
//! `DRIFT_AUDIO_SINK` names another GStreamer sink (`fakesink` plays nothing, in real time), for
//! tests and machines without a sound server.

use std::io::Write;
use std::sync::{Arc, Mutex, MutexGuard};
use std::thread;
use std::time::{Duration, Instant};

use gst::prelude::*;
use gstreamer as gst;
use serde_json::{Value, json};
use svelte_gpui::Emitter;

const SLOT_COUNT: usize = 2;
const EVENT_NAME: &str = "audio";
const POLL_INTERVAL: Duration = Duration::from_millis(40);
const IDLE_POLL_INTERVAL: Duration = Duration::from_millis(250);
const POSITION_INTERVAL: Duration = Duration::from_millis(250);
/// A resume position this close to the start is not worth a seek.
const MINIMUM_SEEK_SECONDS: f64 = 0.05;

pub struct AudioEngine {
    slots: Vec<Slot>,
    emitter: Emitter,
    /// With `DRIFT_AUDIO_LOG` naming a file, every event is also appended there, for tests.
    event_log: Option<Mutex<std::fs::File>>,
}

struct Slot {
    index: usize,
    pipeline: gst::Element,
    state: Mutex<SlotState>,
}

#[derive(Default)]
struct SlotState {
    /// A song is loaded and has not ended or failed.
    loaded: bool,
    /// The pipeline has prerolled, so it can seek and play.
    prerolled: bool,
    /// The song played to its end; a seek brings it back.
    ended: bool,
    /// What the app asked for; the pipeline follows once it can.
    wanted_playing: bool,
    /// The pipeline was paused by buffering, not by the app.
    paused_by_buffering: bool,
    buffering: bool,
    /// Where to seek once prerolled: a resume position or an early seek.
    pending_seek_seconds: Option<f64>,
    announced_duration: Option<f64>,
    last_position_event: Option<Instant>,
}

impl AudioEngine {
    pub fn start(emitter: Emitter) -> Result<Arc<AudioEngine>, String> {
        gst::init().map_err(|error| format!("could not start GStreamer: {error}"))?;
        let mut slots = Vec::new();
        for index in 0..SLOT_COUNT {
            slots.push(Slot::new(index)?);
        }
        let engine = Arc::new(AudioEngine { slots, emitter, event_log: open_event_log() });
        let poller = Arc::clone(&engine);
        thread::Builder::new()
            .name("audio-bus".to_string())
            .spawn(move || poller.poll_forever())
            .map_err(|error| format!("could not start the audio thread: {error}"))?;
        Ok(engine)
    }

    pub fn load(&self, slot_index: usize, url: &str, position_seconds: f64, playing: bool) -> Result<(), String> {
        let slot = self.slot(slot_index)?;
        let mut state = slot.lock();
        reset_pipeline(slot)?;
        *state = SlotState { loaded: true, wanted_playing: playing, ..SlotState::default() };
        if position_seconds > MINIMUM_SEEK_SECONDS {
            state.pending_seek_seconds = Some(position_seconds);
        }
        slot.pipeline.set_property("uri", url);
        set_pipeline_state(slot, gst::State::Paused)
    }

    pub fn play(&self, slot_index: usize) -> Result<(), String> {
        let slot = self.slot(slot_index)?;
        let mut state = slot.lock();
        state.wanted_playing = true;
        if state.prerolled && !state.paused_by_buffering {
            return set_pipeline_state(slot, gst::State::Playing);
        }
        Ok(())
    }

    pub fn pause(&self, slot_index: usize) -> Result<(), String> {
        let slot = self.slot(slot_index)?;
        let mut state = slot.lock();
        state.wanted_playing = false;
        state.paused_by_buffering = false;
        if state.loaded {
            return set_pipeline_state(slot, gst::State::Paused);
        }
        Ok(())
    }

    pub fn seek(&self, slot_index: usize, position_seconds: f64) -> Result<(), String> {
        let slot = self.slot(slot_index)?;
        let mut state = slot.lock();
        if !state.loaded {
            return Ok(());
        }
        if !state.prerolled {
            state.pending_seek_seconds = Some(position_seconds);
            return Ok(());
        }
        seek_pipeline(slot, position_seconds)?;
        state.ended = false;
        state.last_position_event = None;
        Ok(())
    }

    pub fn stop(&self, slot_index: usize) -> Result<(), String> {
        let slot = self.slot(slot_index)?;
        let mut state = slot.lock();
        reset_pipeline(slot)?;
        *state = SlotState::default();
        Ok(())
    }

    pub fn set_volume(&self, slot_index: usize, volume: f64) -> Result<(), String> {
        let slot = self.slot(slot_index)?;
        let volume = volume.clamp(0.0, 1.0);
        slot.pipeline.set_property("volume", volume);
        self.log_event(&json!({ "kind": "volume", "slot": slot_index, "volume": volume }));
        Ok(())
    }

    fn slot(&self, index: usize) -> Result<&Slot, String> {
        self.slots.get(index).ok_or_else(|| format!("there is no audio slot {index}"))
    }

    fn emit(&self, slot: &Slot, mut event: Value) {
        event["slot"] = json!(slot.index);
        self.log_event(&event);
        self.emitter.emit(EVENT_NAME, event);
    }

    fn log_event(&self, event: &Value) {
        let Some(log) = &self.event_log else {
            return;
        };
        if let Ok(mut file) = log.lock() {
            let _ = writeln!(file, "{event}");
        }
    }

    fn poll_forever(&self) {
        loop {
            let mut anything_loaded = false;
            for slot in &self.slots {
                self.drain_bus(slot);
                anything_loaded |= self.report_position(slot);
            }
            if anything_loaded {
                thread::sleep(POLL_INTERVAL);
            } else {
                thread::sleep(IDLE_POLL_INTERVAL);
            }
        }
    }

    fn drain_bus(&self, slot: &Slot) {
        let Some(bus) = slot.pipeline.bus() else {
            return;
        };
        while let Some(message) = bus.pop() {
            self.handle_message(slot, &message);
        }
    }

    fn handle_message(&self, slot: &Slot, message: &gst::Message) {
        use gst::MessageView;
        match message.view() {
            MessageView::Eos(_) => self.handle_end(slot),
            MessageView::Error(error) => self.handle_error(slot, error.error().to_string()),
            MessageView::AsyncDone(_) => self.handle_preroll(slot),
            MessageView::Buffering(buffering) => self.handle_buffering(slot, buffering.percent()),
            MessageView::DurationChanged(_) => self.announce_duration(slot),
            MessageView::StateChanged(changed) => {
                if message.src() == Some(slot.pipeline.upcast_ref::<gst::Object>()) {
                    self.handle_state_change(slot, changed.current());
                }
            }
            _ => {}
        }
    }

    // The song stays loaded, so the app can still seek back into it (repeat one, or the end of the queue).
    fn handle_end(&self, slot: &Slot) {
        slot.lock().ended = true;
        self.emit(slot, json!({ "kind": "ended" }));
    }

    fn handle_error(&self, slot: &Slot, message: String) {
        {
            let mut state = slot.lock();
            state.loaded = false;
            state.wanted_playing = false;
        }
        let _ = slot.pipeline.set_state(gst::State::Null);
        self.emit(slot, json!({ "kind": "error", "message": message }));
    }

    // The pipeline has data to show: do the seek that was waiting, then start if asked to.
    fn handle_preroll(&self, slot: &Slot) {
        let mut state = slot.lock();
        if !state.loaded {
            return;
        }
        state.prerolled = true;
        if let Some(seconds) = state.pending_seek_seconds.take() {
            if let Err(message) = seek_pipeline(slot, seconds) {
                eprintln!("audio seek failed: {message}");
            }
            return;
        }
        drop(state);
        self.announce_duration(slot);
        self.start_if_wanted(slot);
    }

    fn start_if_wanted(&self, slot: &Slot) {
        let state = slot.lock();
        if state.wanted_playing && !state.paused_by_buffering {
            if let Err(message) = set_pipeline_state(slot, gst::State::Playing) {
                eprintln!("audio play failed: {message}");
            }
        }
    }

    fn handle_buffering(&self, slot: &Slot, percent: i32) {
        let mut state = slot.lock();
        if !state.loaded {
            return;
        }
        let was_buffering = state.buffering;
        state.buffering = percent < 100;
        let resume = was_buffering && !state.buffering && state.paused_by_buffering;
        let pause = !was_buffering && state.buffering && state.wanted_playing && state.prerolled;
        if pause {
            state.paused_by_buffering = true;
            let _ = slot.pipeline.set_state(gst::State::Paused);
        }
        if resume {
            state.paused_by_buffering = false;
            let _ = slot.pipeline.set_state(gst::State::Playing);
        }
        let buffering = state.buffering;
        drop(state);
        if was_buffering != buffering {
            self.emit(slot, json!({ "kind": "buffering", "buffering": buffering }));
        }
    }

    fn handle_state_change(&self, slot: &Slot, current: gst::State) {
        let paused_by_buffering = slot.lock().paused_by_buffering;
        if paused_by_buffering {
            return;
        }
        match current {
            gst::State::Playing => self.emit(slot, json!({ "kind": "state", "playing": true })),
            gst::State::Paused => self.emit(slot, json!({ "kind": "state", "playing": false })),
            _ => {}
        }
    }

    fn announce_duration(&self, slot: &Slot) {
        let Some(duration) = slot.pipeline.query_duration::<gst::ClockTime>() else {
            return;
        };
        let seconds = duration.seconds_f64();
        {
            let mut state = slot.lock();
            if state.announced_duration == Some(seconds) {
                return;
            }
            state.announced_duration = Some(seconds);
        }
        self.emit(slot, json!({ "kind": "duration", "durationSeconds": seconds }));
    }

    /// Emits the position while the slot plays; says whether the slot has a song at all.
    fn report_position(&self, slot: &Slot) -> bool {
        let due = {
            let mut state = slot.lock();
            if !state.loaded {
                return false;
            }
            let playing = state.wanted_playing && state.prerolled && !state.paused_by_buffering && !state.ended;
            let elapsed = state.last_position_event.map(|moment| moment.elapsed());
            let overdue = elapsed.is_none_or(|duration| duration >= POSITION_INTERVAL);
            if playing && overdue {
                state.last_position_event = Some(Instant::now());
            }
            playing && overdue
        };
        if due && self.emitter.has_listeners(EVENT_NAME) {
            self.emit_position(slot);
        }
        true
    }

    fn emit_position(&self, slot: &Slot) {
        if let Some(position) = slot.pipeline.query_position::<gst::ClockTime>() {
            self.emit(slot, json!({ "kind": "position", "positionSeconds": position.seconds_f64() }));
        }
    }
}

impl Slot {
    fn new(index: usize) -> Result<Slot, String> {
        let pipeline = gst::ElementFactory::make("playbin")
            .build()
            .map_err(|error| format!("GStreamer has no playbin: {error}"))?;
        // Only audio is wanted, so a stream that carries a picture does not open a window.
        let video_sink = gst::ElementFactory::make("fakesink")
            .build()
            .map_err(|error| format!("GStreamer has no fakesink: {error}"))?;
        pipeline.set_property("video-sink", &video_sink);
        if let Some(sink) = audio_sink_override()? {
            pipeline.set_property("audio-sink", &sink);
        }
        Ok(Slot { index, pipeline, state: Mutex::new(SlotState::default()) })
    }

    fn lock(&self) -> MutexGuard<'_, SlotState> {
        self.state.lock().unwrap_or_else(|poisoned| poisoned.into_inner())
    }
}

fn open_event_log() -> Option<Mutex<std::fs::File>> {
    let path = std::env::var_os("DRIFT_AUDIO_LOG")?;
    let file = std::fs::OpenOptions::new().create(true).append(true).open(path).ok()?;
    Some(Mutex::new(file))
}

fn audio_sink_override() -> Result<Option<gst::Element>, String> {
    let Ok(name) = std::env::var("DRIFT_AUDIO_SINK") else {
        return Ok(None);
    };
    let sink = gst::ElementFactory::make(&name).build().map_err(|error| format!("DRIFT_AUDIO_SINK {name}: {error}"))?;
    // A fakesink does not follow the clock unless told to, and would play a song in a blink.
    if sink.has_property("sync") {
        sink.set_property("sync", true);
    }
    Ok(Some(sink))
}

// Back to a clean pipeline, with the old song's messages dropped.
fn reset_pipeline(slot: &Slot) -> Result<(), String> {
    set_pipeline_state(slot, gst::State::Null)?;
    if let Some(bus) = slot.pipeline.bus() {
        bus.set_flushing(true);
        bus.set_flushing(false);
    }
    Ok(())
}

fn set_pipeline_state(slot: &Slot, state: gst::State) -> Result<(), String> {
    slot.pipeline
        .set_state(state)
        .map(|_| ())
        .map_err(|error| format!("audio slot {} could not change state: {error}", slot.index))
}

fn seek_pipeline(slot: &Slot, seconds: f64) -> Result<(), String> {
    let position = gst::ClockTime::from_nseconds((seconds.max(0.0) * 1_000_000_000.0) as u64);
    // ACCURATE decodes from the container's seek point to the exact position; snapping to a key unit
    // would land seconds early in a stream whose clusters are that long.
    slot.pipeline
        .seek_simple(gst::SeekFlags::FLUSH | gst::SeekFlags::ACCURATE, position)
        .map_err(|error| format!("audio slot {} could not seek: {error}", slot.index))
}
