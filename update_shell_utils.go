// update-shell-utils runs the following commands in parallel:
//   brew update && brew upgrade && brew cleanup && brew prune
//   pip3 install --upgrade && pip2 install --upgrade
//   go get -u <path>
//   rustup update
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
	const numWorkers = 4
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

func upgradeGoBins() error {
	paths := []string{
		"github.com/client9/misspell/cmd/misspell",
		"github.com/dvyukov/go-fuzz/go-fuzz",
		"github.com/dvyukov/go-fuzz/go-fuzz-build",
		"github.com/FiloSottile/gvt",
		"github.com/fzipp/gocyclo",
		"github.com/golang/dep/cmd/dep",
		"github.com/golang/lint/golint",
		"github.com/gordonklaus/ineffassign",
		"github.com/magefile/mage",
		"github.com/nsf/gocode",
		"github.com/shurcooL/binstale",
		"github.com/spf13/cobra/cobra",
		"golang.org/x/tools/cmd/goimports",
		"golang.org/x/tools/cmd/guru",
		"honnef.co/go/tools/cmd/megacheck",
		"mvdan.cc/sh/cmd/shfmt",
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

func run(cmd string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	command := exec.CommandContext(ctx, cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
