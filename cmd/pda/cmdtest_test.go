package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	cmdtest "github.com/google/go-cmdtest"
)

var (
	buildOnce sync.Once
	binary    string
	buildErr  error
)

func buildCLI(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		tmp, err := os.MkdirTemp("", "pda-cli-")
		if err != nil {
			buildErr = err
			return
		}
		bin := filepath.Join(tmp, "pda")
		cmd := exec.Command("go", "build", "-o", bin, ".")
		cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = fmt.Errorf("go build failed: %w\n%s", err, out)
			return
		}
		binary = bin
	})
	if buildErr != nil {
		t.Fatalf("build failed: %v", buildErr)
	}
	return binary
}

func TestCmdtests(t *testing.T) {
	bin := buildCLI(t)

	suite, err := cmdtest.Read("cmdtest")
	if err != nil {
		t.Fatalf("read suite: %v", err)
	}

	suite.Commands["pda"] = cmdtest.Program(bin)

	suite.Run(t, false)
}
