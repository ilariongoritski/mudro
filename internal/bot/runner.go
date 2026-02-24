package bot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/goritskimihail/mudro/internal/config"
)

type Runner struct {
	RepoRoot string
	DSN      string
	Limit    int
	Timeout  time.Duration
}

func NewRunner() *Runner {
	return &Runner{
		RepoRoot: config.RepoRoot(),
		DSN:      config.DSN(),
		Limit:    config.TelegramMessageLimit(),
		Timeout:  30 * time.Second,
	}
}

func (r *Runner) Logs() ([]byte, error) {
	return r.runSteps([][]string{{"docker", "compose", "logs", "--no-color", "--tail=200", "db"}})
}

func (r *Runner) runStep(step []string) ([]byte, error) {
	if len(step) == 0 {
		return nil, fmt.Errorf("empty command step")
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, step[0], step[1:]...)
	cmd.Dir = r.RepoRoot
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DSN="+r.DSN)
	return cmd.CombinedOutput()
}

func (r *Runner) runSteps(steps [][]string) ([]byte, error) {
	var out strings.Builder
	for _, step := range steps {
		out.WriteString("$ " + strings.Join(step, " ") + "\n")
		b, err := r.runStep(step)
		out.Write(b)
		if err != nil {
			return []byte(out.String()), err
		}
	}
	return []byte(out.String()), nil
}
