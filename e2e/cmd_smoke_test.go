package e2e

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func goBin(t *testing.T) string {
	t.Helper()
	if _, err := os.Stat("/usr/local/go/bin/go"); err == nil {
		return "/usr/local/go/bin/go"
	}
	return "go"
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("repoRoot: %v", err)
	}
	return root
}

func TestCmdAPISmokeHealthz(t *testing.T) {
	if testing.Short() {
		t.Skip("skip smoke test in short mode")
	}

	root := repoRoot(t)
	bin := filepath.Join(t.TempDir(), "mudro-api-smoke")
	build := exec.Command(goBin(t), "build", "-o", bin, "./services/feed-api/cmd")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build api: %v\n%s", err, string(out))
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin)
	cmd.Dir = root
	cmd.Env = append(os.Environ(),
		"API_ADDR="+addr,
		"DSN=postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		t.Fatalf("start api: %v", err)
	}
	defer func() {
		if cmd.Process != nil && (cmd.ProcessState == nil || !cmd.ProcessState.Exited()) {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()

	url := "http://" + addr + "/healthz"
	client := &http.Client{Timeout: 500 * time.Millisecond}
	var ok bool
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ok = true
				break
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
	if !ok {
		t.Fatalf("api did not become healthy at %s\noutput:\n%s", url, out.String())
	}

	if err := cmd.Process.Kill(); err != nil {
		t.Fatalf("kill api: %v", err)
	}
	if err := cmd.Wait(); err != nil && !strings.Contains(err.Error(), "signal: killed") {
		t.Fatalf("wait api: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), fmt.Sprintf("api listening on %s", addr)) {
		t.Fatalf("missing listen log, output:\n%s", out.String())
	}
}

func TestCmdBotSmokeMissingToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skip smoke test in short mode")
	}

	root := repoRoot(t)
	bin := filepath.Join(t.TempDir(), "mudro-bot-smoke")
	build := exec.Command(goBin(t), "build", "-o", bin, "./services/bot/cmd")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build bot: %v\n%s", err, string(out))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin)
	cmd.Dir = root
	env := make([]string, 0, len(os.Environ()))
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "TELEGRAM_BOT_TOKEN=") {
			continue
		}
		env = append(env, e)
	}
	env = append(env, "DSN=postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable")
	cmd.Env = env

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected bot startup failure without token")
	}
	if !strings.Contains(out.String(), "missing required env: TELEGRAM_BOT_TOKEN") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
}
