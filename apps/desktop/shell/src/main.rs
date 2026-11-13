//! The Drift desktop shell. It starts the Go backend, plays audio, and runs the UI bundle built from
//! `apps/desktop/frontend` with svelte-gpui.
//!
//! Commands the UI calls (typed in `frontend/src/lib/backend/commands.ts`):
//! - `backend_address` answers `{ url, token }` once the backend listens.
//! - `open_external` opens an http(s) link in the default browser.
//! - `audio_load`, `audio_play`, `audio_pause`, `audio_seek`, `audio_stop`, `audio_set_volume`,
//!   `audio_set_gain`, `audio_crossfade` and `audio_cancel_crossfade` drive the audio engine; what it
//!   reports arrives as `audio` events (see `audio.rs`).
//!
//! Run it against a build of the UI:
//!
//! ```text
//! cd apps/desktop/frontend && bun build.mjs
//! cd apps/desktop/shell && cargo run --release -- --bundle ../frontend/dist/drift.js
//! ```

mod audio;
mod backend;
mod external;

use std::sync::Arc;

use audio::AudioEngine;
use backend::Backend;
use serde_json::{Value, json};
use svelte_gpui::{CommandError, Commands, Options};

/// A part that failed to start stays in place as its error, which the commands that need it report to the UI.
type Started<T> = Arc<Result<T, String>>;

fn main() {
    let commands = Commands::new();
    let backend: Started<Backend> = Arc::new(Backend::start().inspect_err(report));
    let engine: Started<Arc<AudioEngine>> = Arc::new(AudioEngine::start(commands.emitter()).inspect_err(report));
    publish_address_for_tools(&backend);
    let commands = register_shell_commands(commands, backend);
    let commands = register_audio_commands(commands, engine);
    svelte_gpui::run(options().commands(commands));
}

fn report(message: &String) {
    eprintln!("drift: {message}");
}

/// With `DRIFT_BACKEND_ADDRESS_FILE` set, writes the backend's address there once it listens, so a
/// script (a test driving the same backend as the UI) can connect to it.
fn publish_address_for_tools(backend: &Started<Backend>) {
    let Some(path) = std::env::var_os("DRIFT_BACKEND_ADDRESS_FILE") else {
        return;
    };
    let backend = Arc::clone(backend);
    std::thread::spawn(move || {
        let Ok(started) = backend.as_ref() else {
            return;
        };
        if let Ok(address) = started.address() {
            let _ = std::fs::write(path, address.to_json().to_string());
        }
    });
}

fn register_shell_commands(commands: Commands, backend: Started<Backend>) -> Commands {
    commands
        .command("backend_address", move |_| {
            let started = backend.as_ref().as_ref().map_err(|message| message.clone())?;
            Ok::<_, CommandError>(started.address()?.to_json())
        })
        .command("open_external", |args| {
            let url = args["url"].as_str().ok_or("open_external needs a url")?;
            external::open(url)?;
            Ok::<_, CommandError>(Value::Null)
        })
}

fn register_audio_commands(commands: Commands, engine: Started<Arc<AudioEngine>>) -> Commands {
    commands
        .command(
            "audio_load",
            with_engine(&engine, |engine, args| {
                let url = args["url"].as_str().ok_or("audio_load needs a url")?;
                let position = args["positionSeconds"].as_f64().unwrap_or(0.0);
                let playing = args["playing"].as_bool().unwrap_or(true);
                engine.load(slot_of(args)?, url, position, playing)
            }),
        )
        .command("audio_play", with_engine(&engine, |engine, args| engine.play(slot_of(args)?)))
        .command("audio_pause", with_engine(&engine, |engine, args| engine.pause(slot_of(args)?)))
        .command("audio_stop", with_engine(&engine, |engine, args| engine.stop(slot_of(args)?)))
        .command(
            "audio_seek",
            with_engine(&engine, |engine, args| {
                let position = args["positionSeconds"].as_f64().ok_or("audio_seek needs positionSeconds")?;
                engine.seek(slot_of(args)?, position)
            }),
        )
        .command(
            "audio_set_volume",
            with_engine(&engine, |engine, args| {
                let volume = args["volume"].as_f64().ok_or("audio_set_volume needs a volume")?;
                engine.set_volume(volume);
                Ok(())
            }),
        )
        .command(
            "audio_set_gain",
            with_engine(&engine, |engine, args| {
                let gain = args["gain"].as_f64().ok_or("audio_set_gain needs a gain")?;
                engine.set_gain(slot_of(args)?, gain)
            }),
        )
        .command(
            "audio_crossfade",
            with_engine(&engine, |engine, args| {
                let seconds = args["seconds"].as_f64().ok_or("audio_crossfade needs seconds")?;
                engine.start_crossfade(seconds)
            }),
        )
        .command(
            "audio_cancel_crossfade",
            with_engine(&engine, |engine, _| {
                engine.cancel_crossfade();
                Ok(())
            }),
        )
}

// Wraps an engine call as a command that fails with the reason when GStreamer did not start.
fn with_engine(
    engine: &Started<Arc<AudioEngine>>,
    call: impl Fn(&Arc<AudioEngine>, &Value) -> Result<(), String> + Send + Sync + 'static,
) -> impl Fn(Value) -> Result<Value, CommandError> + Send + Sync + 'static {
    let engine = Arc::clone(engine);
    move |args| {
        let running = engine.as_ref().as_ref().map_err(|message| message.clone())?;
        call(running, &args)?;
        Ok(json!(null))
    }
}

fn slot_of(args: &Value) -> Result<usize, String> {
    args["slot"].as_u64().map(|slot| slot as usize).ok_or_else(|| "the call needs a slot".to_string())
}

#[cfg(feature = "package")]
fn options() -> Options {
    Options::embedded(svelte_gpui::embed!())
}

#[cfg(not(feature = "package"))]
fn options() -> Options {
    Options::from_command_line().unwrap_or_else(|| {
        eprintln!("usage: drift --bundle apps/desktop/frontend/dist/drift.js");
        std::process::exit(2);
    })
}
