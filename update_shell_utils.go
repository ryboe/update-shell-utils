// update-shell-utils runs the following commands in parallel:
//   brew update
//   brew upgrade
//   nvim +PlugUpgrade +PlugUpdate        # upgrade all neovim packages
//   pip3 install --upgrade
//   poetry self:update
//   rustup update
//   softwareupdate -ia
//   subl --command update_check          # upgrade sublime text
//   subl --command upgrade_all_packages  # upgrade all sublime [ackages

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	const numWorkers = 5
	errc := make(chan error, numWorkers)

	go func() {
		errc <- brewUpgrade()
	}()

	go func() {
		errc <- macOSUpdate()
	}()

	go func() {
		errc <- pipUpgrade()
	}()

	go func() {
		errc <- rustupUpdate()
	}()

	go func() {
		errc <- sublPkgUpgrade()
	}()

	for i := 0; i < numWorkers; i++ {
		if err := <-errc; err != nil {
			fmt.Println(err)
		}
	}
}

func brewUpgrade() error {
	err := run("brew", "update")
	if err != nil {
		return err
	}

	return run("brew", "upgrade")
}

func pipUpgrade() error {
	if err := run("pip3", "install", "--upgrade", "pip", "setuptools", "wheel"); err != nil {
		return err
	}

	return run("pip3", "install", "--user", "--upgrade", "poetry")
}

func rustupUpdate() error {
	return run("rustup", "update")
}

func sublPkgUpgrade() error {
	return run("subl", "--command", "update_check", "--command", "upgrade_all_packages")
}

func macOSUpdate() error {
	// -i install updates
	// -a install *all* updates
	return run("sudo", "softwareupdate", "-ia")
}

func run(cmd string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute) // macOS updates can take a while
	defer cancel()

	command := exec.CommandContext(ctx, cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
