// update-shell-utils runs the following commands in parallel:
//   brew update && brew upgrade && brew cleanup && brew prune
//   pip3 install --upgrade && pip2 install --upgrade
//   go get -u <path>
//   rustup update
//   softwareupdate -ia
//   subl --command update_check
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	const numWorkers = 7
	errc := make(chan error, numWorkers)

	go func() {
		errc <- brewUpgrade()
	}()

	go func() {
		errc <- macOSUpdate()
	}()

	go func() {
		errc <- nvimPlugUpdate()
	}()

	go func() {
		errc <- pipUpgradeAll()
	}()

	go func() {
		errc <- upgradeGoBins()
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
	var err error
	for _, subcmd := range []string{"update", "upgrade", "cleanup", "prune"} {
		err = run("brew", subcmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func pipUpgradeAll() error {
	err := pipUpgrade("3")
	if err != nil {
		return err
	}

	return pipUpgrade("2")
}

func pipUpgrade(pipVersion string) error {
	if pipVersion != "2" && pipVersion != "3" {
		errMsg := fmt.Sprintf("%q is not a valid version of the pip package manager (must be 2 or 3)", pipVersion)
		panic(errMsg)
	}

	pipCmd := fmt.Sprintf("pip%s", pipVersion)
	pkgs, err := outdatedPipPkgs(pipCmd)
	if err != nil {
		return err
	}
	if len(pkgs) == 0 {
		return nil
	}

	args := append([]string{"sudo", "-H", pipCmd, "install", "--upgrade"}, pkgs...)
	return run("sudo", args...)
}

func outdatedPipPkgs(pipCmd string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sudo", "-H", pipCmd, "list", "--format=freeze")
	var buf strings.Builder
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return extractPipPkgs(buf.String()), nil
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

func upgradeGoBins() error {
	paths := []string{
		"github.com/golangci/golangci-lint/cmd/golangci-lint",
		"github.com/magefile/mage",
		"golang.org/x/tools/cmd/goimports",
		"golang.org/x/tools/cmd/guru",
	}

	var err error
	for _, path := range paths {
		err = run("go", "get", "-u", path)
	}

	return err
}

func rustupUpdate() error {
	err := run("rustup", "self", "update")
	if err != nil {
		return err
	}

	return run("rustup", "update")
}

func sublPkgUpgrade() error {
	err := run("subl", "--command", "update_check")
	if err != nil {
		return err
	}

	return run("subl", "--command", "upgrade_all_packages")
}

func nvimPlugUpdate() error {
	return run("nvim", "+PlugUpgrade", "+PlugUpdate", "+qa")
}

func macOSUpdate() error {
	// -i install updates
	// -a install *all* updates
	return run("sudo", "softwareupdate", "-ia")
}

func run(cmd string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	command := exec.CommandContext(ctx, cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
