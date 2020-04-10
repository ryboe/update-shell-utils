//! Run these commands concurrently to update CLI utils, and macOS.
//! ```sh
//! > brew update && brew upgrade && gcloud components update
//! > softwareupdate -ia
//! > pip install --upgrade pip setuptools wheel
//! > rustup update && cargo install <pkgs>
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

use regex::bytes::RegexBuilder;
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
    let _ = Command::new("brew").arg("upgrade").status()?;

    // After brew upgrades gcloud, upgrade gcloud components.
    Command::new("gcloud")
        .arg("components")
        .arg("update")
        .status()
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
    let _ = Command::new("rustup").arg("update").status()?;

    let output = Command::new("cargo")
        .arg("install")
        .arg("--list")
        .output()?;

    let re = RegexBuilder::new(r"^\S+").multi_line(true).build().unwrap();

    let do_not_update = ["gitprompt", "rustlings", "update-shell-utils"];

    let pkgs_to_update: Vec<String> = re
        .captures_iter(&output.stdout)
        .filter_map(|cap| {
            let pkg = String::from_utf8_lossy(&cap[0]).into_owned();
            if do_not_update.contains(&pkg.as_str()) {
                None
            } else {
                Some(pkg)
            }
        })
        .collect();

    Command::new("cargo")
        .arg("install")
        .args(pkgs_to_update)
        .status()
}
