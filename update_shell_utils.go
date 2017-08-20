// update-shell-utils runs the following commands in parallel:
//   brew update && brew upgrade && brew cleanup && brew prune
//   pip3 install --upgrade && pip2 install --upgrade
//   gcloud components update
//   upgrade_oh_my_zsh
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	const numWorkers = 4
	errc := make(chan error, numWorkers)

	go func() {
		errc <- brewUpgrade()
	}()

	go func() {
		errc <- pipUpgrade()
	}()

	go func() {
		errc <- run("gcloud", "components", "update", "--quiet")
	}()

	go func() {
		errc <- upgradeOhMyZSH()
	}()

	for i := 0; i < numWorkers; i++ {
		if err := <-errc; err != nil {
			fmt.Println(err)
		}
	}
}

func brewUpgrade() error {
	var err error
	for _, subcmd := range []string{"update", "upgrade", "cleanup", "prune"} {
		err = run("brew", subcmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func pipUpgrade() error {
	py3Pkgs, err := outdatedPipPkgs("3")
	if err != nil {
		return err
	}

	args := []string{"install", "--upgrade"}
	err = run("pip3", append(args, py3Pkgs...)...)
	if err != nil {
		return err
	}

	py2Pkgs, err := outdatedPipPkgs("2")
	if err != nil {
		return err
	}

	return run("pip2", append(args, py2Pkgs...)...)
}

func outdatedPipPkgs(pyVersion string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	command := exec.CommandContext(ctx, "pip"+pyVersion, "list", "--outdated", "--format=freeze")
	var outBuf bytes.Buffer
	command.Stdout = &outBuf

	if err := command.Run(); err != nil {
		return nil, err
	}

	return extractPipPkgs(outBuf.String()), nil
}

func extractPipPkgs(output string) []string {
	lines := strings.Split(output, "\n")
	pkgs := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		chunks := strings.Split(line, "==")
		if len(chunks) < 2 {
			continue // line doesn't contain "=="
		}

		pkg := strings.TrimSpace(chunks[0])
		pkgs = append(pkgs, pkg)
	}

	return pkgs
}

func upgradeOhMyZSH() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	command := exec.CommandContext(ctx, "git", "pull", "--rebase", "--stat", "origin", "master")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	zshDir := os.Getenv("ZSH")
	if zshDir == "" {
		zshDir = filepath.Join(os.Getenv("HOME"), ".oh-my-zsh")
	}
	command.Dir = filepath.Join(zshDir, "tools")

	return command.Run()
}

func run(cmd string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	command := exec.CommandContext(ctx, cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
