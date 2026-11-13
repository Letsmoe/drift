//! Opens links in the default browser.

use std::process::{Command, Stdio};

/// Opens an http(s) link; anything else is refused, so the UI cannot launch arbitrary files or programs.
pub fn open(url: &str) -> Result<(), String> {
    if !url.starts_with("https://") && !url.starts_with("http://") {
        return Err(format!("only http and https links can be opened, not {url}"));
    }
    let mut command = opener(url);
    command
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .spawn()
        .map(|_| ())
        .map_err(|error| format!("could not open {url}: {error}"))
}

#[cfg(target_os = "macos")]
fn opener(url: &str) -> Command {
    let mut command = Command::new("open");
    command.arg(url);
    command
}

#[cfg(windows)]
fn opener(url: &str) -> Command {
    // Not `cmd /c start`: cmd would read an `&` in the URL as another command.
    let mut command = Command::new("rundll32");
    command.args(["url.dll,FileProtocolHandler", url]);
    command
}

#[cfg(not(any(target_os = "macos", windows)))]
fn opener(url: &str) -> Command {
    let mut command = Command::new("xdg-open");
    command.arg(url);
    command
}
