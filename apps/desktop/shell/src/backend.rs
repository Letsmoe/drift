//! The Go backend: started as a child process, which announces where it listens on stdout.

use std::env;
use std::io::{BufRead, BufReader};
use std::path::PathBuf;
use std::process::{Child, ChildStdin, Command, Stdio};
use std::sync::{Arc, Condvar, Mutex};
use std::thread;
use std::time::Duration;

use serde_json::{Value, json};

/// The stdout line that carries the address; see `apps/desktop/main.go`.
const ANNOUNCE_PREFIX: &str = "drift-backend ";
const START_TIMEOUT: Duration = Duration::from_secs(30);

#[cfg(windows)]
const BINARY_NAME: &str = "drift-backend.exe";
#[cfg(not(windows))]
const BINARY_NAME: &str = "drift-backend";

/// Where the backend listens and the token its RPC socket requires.
#[derive(Clone)]
pub struct Address {
    pub url: String,
    pub token: String,
}

impl Address {
    pub fn to_json(&self) -> Value {
        json!({ "url": self.url, "token": self.token })
    }
}

type Announcement = Option<Result<Address, String>>;

/// A running backend. It shuts down when its stdin closes, which happens when this process ends,
/// however it ends; dropping this stops it too.
pub struct Backend {
    announcement: Arc<(Mutex<Announcement>, Condvar)>,
    child: Child,
    // Held so the pipe stays open for as long as the shell runs.
    _stdin: Option<ChildStdin>,
}

impl Backend {
    pub fn start() -> Result<Backend, String> {
        let binary = find_binary()?;
        let mut child = Command::new(&binary)
            .arg("-exit-with-stdin")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::inherit())
            .spawn()
            .map_err(|error| format!("could not start the backend at {}: {error}", binary.display()))?;
        let stdin = child.stdin.take();
        let stdout = child.stdout.take().ok_or("the backend has no stdout")?;

        let announcement = Arc::new((Mutex::new(None), Condvar::new()));
        let reader_announcement = Arc::clone(&announcement);
        thread::spawn(move || read_announcement(stdout, reader_announcement));
        Ok(Backend { announcement, child, _stdin: stdin })
    }

    /// Waits for the backend to say where it listens.
    pub fn address(&self) -> Result<Address, String> {
        let (lock, ready) = &*self.announcement;
        let guard = lock.lock().map_err(|_| "the backend state is poisoned")?;
        let (guard, _) = ready
            .wait_timeout_while(guard, START_TIMEOUT, |announcement| announcement.is_none())
            .map_err(|_| "the backend state is poisoned")?;
        match &*guard {
            Some(result) => result.clone(),
            None => Err(format!("the backend did not start within {} s", START_TIMEOUT.as_secs())),
        }
    }
}

impl Drop for Backend {
    fn drop(&mut self) {
        let _ = self.child.kill();
        let _ = self.child.wait();
    }
}

fn read_announcement(stdout: impl std::io::Read, announcement: Arc<(Mutex<Announcement>, Condvar)>) {
    let mut announced = false;
    for line in BufReader::new(stdout).lines().map_while(Result::ok) {
        if announced {
            eprintln!("{line}");
            continue;
        }
        match line.strip_prefix(ANNOUNCE_PREFIX) {
            Some(text) => {
                announced = true;
                publish(&announcement, parse_address(text));
            }
            None => eprintln!("{line}"),
        }
    }
    if !announced {
        publish(&announcement, Err("the backend exited before it started listening".to_string()));
    }
}

fn parse_address(text: &str) -> Result<Address, String> {
    let value: Value = serde_json::from_str(text).map_err(|error| format!("bad backend announcement: {error}"))?;
    let url = value["url"].as_str().ok_or("the backend announced no url")?;
    let token = value["token"].as_str().ok_or("the backend announced no token")?;
    Ok(Address { url: url.to_string(), token: token.to_string() })
}

fn publish(announcement: &(Mutex<Announcement>, Condvar), result: Result<Address, String>) {
    let (lock, ready) = announcement;
    if let Ok(mut guard) = lock.lock() {
        *guard = Some(result);
        ready.notify_all();
    }
}

/// `DRIFT_BACKEND_BINARY`, else next to this executable (or under `backend/`, as packaged),
/// else the development build of `apps/desktop/build/bin`.
fn find_binary() -> Result<PathBuf, String> {
    if let Some(path) = env::var_os("DRIFT_BACKEND_BINARY") {
        return Ok(PathBuf::from(path));
    }
    let mut candidates = Vec::new();
    if let Some(directory) = env::current_exe().ok().and_then(|path| path.parent().map(PathBuf::from)) {
        candidates.push(directory.join(BINARY_NAME));
        candidates.push(directory.join("backend").join(BINARY_NAME));
    }
    candidates.push(PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("../build/bin").join(BINARY_NAME));
    candidates
        .iter()
        .find(|path| path.is_file())
        .cloned()
        .ok_or_else(|| format!("no drift-backend found; looked in {candidates:?}"))
}
