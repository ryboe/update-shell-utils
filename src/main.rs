//! Run these commands concurrently to update CLI utils, and macOS.
//! ```sh
//! > brew update && brew upgrade
//! > softwareupdate -ia
//! > pip install --upgrade pip setuptools wheel
//! > rustup update
//! ```

#![deny(
    bare_trait_objects,
    missing_copy_implementations,
    missing_debug_implementations,
    missing_docs,
    trivial_casts,
    trivial_numeric_casts,
    unsafe_code,
    unused_extern_crates,
    unused_import_braces,
    unused_qualifications,
    unused_results
)]

use std::io;
use std::process::{Command, ExitStatus};
use std::sync::mpsc::{channel, Receiver};
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
fn start_workers() -> Receiver<io::Result<ExitStatus>> {
    let (tx, rx) = channel();

    let update_funcs: Vec<fn() -> io::Result<ExitStatus>> =
        vec![brew_upgrade, macos_update, pip_upgrade, rustup_update];

    for update_func in update_funcs {
        let tx = tx.clone();
        let _ = thread::spawn(move || {
            tx.send(update_func()).unwrap();
        });
    }

    rx
}

/// Upgrade all the homebrew utils.
fn brew_upgrade() -> io::Result<ExitStatus> {
    let _ = Command::new("brew").arg("update").status()?;
    Command::new("brew").arg("upgrade").status()
}

/// Upgrade macOS itself.
fn macos_update() -> io::Result<ExitStatus> {
    Command::new("softwareupdate").arg("-ia").status()
}

/// Upgrade pip, setuptools, and wheel with pip.
fn pip_upgrade() -> io::Result<ExitStatus> {
    let _ = Command::new("python3")
        .args(&["-m", "pip", "install", "--upgrade", "pip"])
        .status()?;

    Command::new("python3")
        .args(&["-m", "pip", "install", "--upgrade", "setuptools", "wheel"])
        .status()
}

/// Upgrade the currently installeed Rust toolchains.
fn rustup_update() -> io::Result<ExitStatus> {
    Command::new("rustup").arg("update").status()
}
