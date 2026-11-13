//! Volume and the crossfade. A slot plays at the user's volume times its own gain. The crossfade moves
//! the gains of the two slots along an equal-power curve on a thread of its own, so it runs the same
//! however busy the UI is.

use std::f64::consts::FRAC_PI_2;
use std::sync::{Arc, MutexGuard};
use std::thread;
use std::time::{Duration, Instant};

use gstreamer::prelude::*;
use serde_json::json;

use super::{AudioEngine, MAIN_SLOT, OUTGOING_SLOT, SLOT_COUNT};

const FADE_STEP: Duration = Duration::from_millis(25);
const MINIMUM_FADE: Duration = Duration::from_millis(100);

pub struct Mixer {
    /// The user's volume with mute applied, from 0 to 1.
    volume: f64,
    /// A multiplier on `volume` for each slot.
    gains: [f64; SLOT_COUNT],
    /// The crossfade that owns the gains right now. A newer or cancelled fade makes an older thread stop.
    running_fade: Option<u64>,
    fades_started: u64,
}

impl Mixer {
    pub fn new() -> Mixer {
        Mixer { volume: 1.0, gains: [1.0; SLOT_COUNT], running_fade: None, fades_started: 0 }
    }

    fn slot_volume(&self, slot_index: usize) -> f64 {
        (self.volume * self.gains[slot_index]).clamp(0.0, 1.0)
    }
}

/// The gains of the song coming in and of the song going out at `progress` (0 to 1) of the crossfade.
/// The squares of the two add up to one, so the loudness stays even while both songs play.
fn equal_power_gains(progress: f64) -> (f64, f64) {
    let angle = progress.clamp(0.0, 1.0) * FRAC_PI_2;
    (angle.sin(), angle.cos())
}

impl AudioEngine {
    /// Sets the user's volume (0 when muted) for every slot.
    pub fn set_volume(&self, volume: f64) {
        let mut mixer = self.lock_mixer();
        mixer.volume = volume.clamp(0.0, 1.0);
        self.apply_volumes(&mixer);
    }

    /// Sets what a slot's volume is multiplied with, apart from a running crossfade.
    pub fn set_gain(&self, slot_index: usize, gain: f64) -> Result<(), String> {
        self.slot(slot_index)?;
        let mut mixer = self.lock_mixer();
        mixer.gains[slot_index] = gain.clamp(0.0, 1.0);
        self.apply_volumes(&mixer);
        Ok(())
    }

    /// Fades the main slot in and the outgoing slot out over `seconds`. The outgoing slot is stopped
    /// at the end, and a `crossfade` event says how the fade ended.
    pub fn start_crossfade(self: &Arc<Self>, seconds: f64) -> Result<(), String> {
        if !self.slot(OUTGOING_SLOT)?.lock().loaded {
            return Err("there is no song to fade out".to_string());
        }
        let duration = Duration::from_secs_f64(seconds.max(0.0)).max(MINIMUM_FADE);
        let fade_id = self.begin_fade();
        let engine = Arc::clone(self);
        thread::Builder::new()
            .name("audio-fade".to_string())
            .spawn(move || engine.run_crossfade(fade_id, duration))
            .map(|_| ())
            .map_err(|error| format!("could not start the crossfade: {error}"))
    }

    /// Ends a running crossfade at once: the outgoing song stops and the main slot plays at full volume.
    pub fn cancel_crossfade(&self) {
        if self.end_fade(None, "cancelled") {
            self.stop_outgoing_slot();
        }
    }

    fn begin_fade(&self) -> u64 {
        let mut mixer = self.lock_mixer();
        mixer.fades_started += 1;
        mixer.running_fade = Some(mixer.fades_started);
        mixer.gains = [0.0, 1.0];
        self.apply_volumes(&mixer);
        mixer.fades_started
    }

    fn run_crossfade(&self, fade_id: u64, duration: Duration) {
        let started = Instant::now();
        loop {
            thread::sleep(FADE_STEP);
            let progress = (started.elapsed().as_secs_f64() / duration.as_secs_f64()).min(1.0);
            if !self.apply_fade_step(fade_id, progress) {
                return;
            }
            if progress >= 1.0 {
                break;
            }
        }
        if self.end_fade(Some(fade_id), "finished") {
            self.stop_outgoing_slot();
        }
    }

    /// Moves the gains to `progress`; false when this fade is no longer the running one.
    fn apply_fade_step(&self, fade_id: u64, progress: f64) -> bool {
        let mut mixer = self.lock_mixer();
        if mixer.running_fade != Some(fade_id) {
            return false;
        }
        let (incoming, outgoing) = equal_power_gains(progress);
        mixer.gains = [incoming, outgoing];
        self.apply_volumes(&mixer);
        true
    }

    /// Ends the running fade (only the given one, when an id is given), leaving the main slot at full
    /// gain, and reports how it ended. False when no such fade ran.
    fn end_fade(&self, only_fade_id: Option<u64>, how: &str) -> bool {
        {
            let mut mixer = self.lock_mixer();
            let Some(running) = mixer.running_fade else {
                return false;
            };
            if only_fade_id.is_some_and(|fade_id| fade_id != running) {
                return false;
            }
            mixer.running_fade = None;
            mixer.gains = [1.0, 0.0];
            self.apply_volumes(&mixer);
        }
        self.emit_without_slot(json!({ "kind": "crossfade", "how": how }));
        true
    }

    fn stop_outgoing_slot(&self) {
        if let Err(message) = self.stop(OUTGOING_SLOT) {
            eprintln!("audio crossfade could not stop the outgoing song: {message}");
        }
    }

    fn apply_volumes(&self, mixer: &Mixer) {
        for (slot_index, slot) in self.slots.iter().enumerate() {
            let volume = mixer.slot_volume(slot_index);
            slot.pipeline.set_property("volume", volume);
            self.log_event(&json!({ "kind": "volume", "slot": slot_index, "volume": volume }));
        }
    }

    fn lock_mixer(&self) -> MutexGuard<'_, Mixer> {
        self.mixer.lock().unwrap_or_else(|poisoned| poisoned.into_inner())
    }

    /// A user action on the main slot ends a crossfade; the outgoing slot is the crossfade's own.
    pub(super) fn cancel_crossfade_for(&self, slot_index: usize) {
        if slot_index == MAIN_SLOT {
            self.cancel_crossfade();
        }
    }
}

#[cfg(test)]
mod tests {
    use super::equal_power_gains;

    #[test]
    fn the_curve_starts_on_the_outgoing_song_and_ends_on_the_incoming_one() {
        assert_eq!(equal_power_gains(0.0), (0.0, 1.0));
        let (incoming, outgoing) = equal_power_gains(1.0);
        assert!((incoming - 1.0).abs() < 1e-12);
        assert!(outgoing.abs() < 1e-12);
    }

    #[test]
    fn the_power_stays_even_all_the_way() {
        for step in 0..=100 {
            let (incoming, outgoing) = equal_power_gains(f64::from(step) / 100.0);
            assert!((incoming * incoming + outgoing * outgoing - 1.0).abs() < 1e-12);
        }
    }
}
