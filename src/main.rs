//! Run these commands concurrently to update CLI utils, and macOS.
//! ```sh
//! > brew update && brew upgrade
//! > softwareupdate -ia
//! > pip install --upgrade pip setuptools wheel
//! > rustup update
//! ```

use std::io;
use std::process::{Command, ExitStatus};
use std::sync::mpsc;
use std::thread;

fn main() {
    let rx = start_workers();

    for recvd in rx {
        if recvd.is_err() {
            println!("error: {:?}", recvd);
        }
    }
}

/// Run all the update commands in separate threads. Return a channel receiver
/// for waiting on the commands to complete.
fn start_workers() -> mpsc::Receiver<io::Result<ExitStatus>> {
    let num_workers = 4;
    let (tx, rx) = mpsc::channel();

    let mut update_funcs: Vec<fn() -> io::Result<ExitStatus>> = Vec::with_capacity(num_workers);
    update_funcs.push(brew_upgrade);
    update_funcs.push(macos_update);
    update_funcs.push(pip_upgrade);
    update_funcs.push(rustup_update);

    for update_func in update_funcs {
        let tx = tx.clone();
        thread::spawn(move || {
            tx.send(update_func()).unwrap();
        });
    }

    rx
}

/// Upgrade all the homebrew utils.
fn brew_upgrade() -> io::Result<ExitStatus> {
    Command::new("brew").arg("update").status()?;
    Command::new("brew").arg("upgrade").status()
}

/// Upgrade macOS itself.
fn macos_update() -> io::Result<ExitStatus> {
    Command::new("sudo")
        .arg("softwareupdate")
        .arg("-ia")
        .status()
}

/// Upgrade pip, setuptools, and wheel with pip.
fn pip_upgrade() -> io::Result<ExitStatus> {
    Command::new("pip")
        .arg("install")
        .arg("--upgrade")
        .args(&["pip", "setuptools", "wheel"])
        .status()
}

/// Upgrade the currently installeed Rust toolchains.
fn rustup_update() -> io::Result<ExitStatus> {
    Command::new("rustup").arg("update").status()
}
