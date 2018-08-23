// update-shell-utils runs the following commands in parallel:
//   brew update && brew upgrade && brew cleanup && brew prune
//   pip3 install --upgrade && pip2 install --upgrade
//   go get -u <path>
//   rustup update
//   softwareupdate -ia
package main

import (
	"bytes"
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
		errc <- pipUpgrade()
	}()

	go func() {
		errc <- upgradeGoBins()
	}()

	go func() {
		errc <- rustupUpdate()
	}()

	go func() {
		errc <- nvimPlugUpdate()
	}()

	go func() {
		errc <- sublPkgUpgrade()
	}()

	go func() {
		errc <- macOSUpdate()
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
	py3Pkgs, err := outdatedPipPkgs()
	if err != nil {
		return err
	}
	if len(py3Pkgs) == 0 {
		return nil
	}

	args := []string{"install", "--upgrade"}
	args = append(args, py3Pkgs...)
	return run("pip", args...)
}

func outdatedPipPkgs() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pip", "list", "--outdated", "--format=freeze")
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	if err := cmd.Run(); err != nil {
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

func upgradeGoBins() error {
	paths := []string{
		"github.com/dvyukov/go-fuzz/go-fuzz",
		"github.com/dvyukov/go-fuzz/go-fuzz-build",
		"github.com/golangci/golangci-lint/cmd/golangci-lint",
		"github.com/magefile/mage",
		"github.com/motemen/gore",
		"github.com/nsf/gocode",
		"github.com/shurcooL/binstale",
		"github.com/spf13/cobra/cobra",
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
	return run("rustup", "update")
}

func nvimPlugUpdate() error {
	return run("nvim", "+PlugUpgrade", "+PlugUpdate", "+qa")
}

func sublPkgUpgrade() error {
	err := run("subl", "--command", "update_check")
	if err != nil {
		return err
	}

	return run("subl", "--command", "upgrade_all_packages")
}

func macOSUpdate() error {
	// -i install updates
	// -a install *all* updates
	// -v verbose
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
