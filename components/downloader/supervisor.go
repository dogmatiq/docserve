package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/dogmatiq/browser/components/internal/worker"
	"github.com/dogmatiq/browser/messages"
	"github.com/dogmatiq/minibus"
)

// Supervisor oversees a pool of works that download Go modules.
type Supervisor struct {
	Environment []string
	Workers     int
	Logger      *slog.Logger
}

// Run starts the supervisor.
func (s *Supervisor) Run(ctx context.Context) error {
	return worker.RunPool(
		ctx,
		max(s.Workers, 1),
		s.download,
	)
}

func (s *Supervisor) download(
	ctx context.Context,
	workerID int,
	m messages.ModuleDiscovered,
) (err error) {
	logger := s.Logger.With(
		slog.Group(
			"worker",
			slog.Int("id", workerID),
		),
		slog.Group(
			"module",
			slog.String("path", m.ModulePath),
			slog.String("version", m.ModuleVersion),
		),
	)

	start := time.Now()
	cached := false

	defer func() {
		if err == nil {
			if cached {
				logger.InfoContext(
					ctx,
					"module already downloaded",
					slog.Duration("elapsed", time.Since(start)),
				)
			} else {
				logger.InfoContext(
					ctx,
					"module downloaded",
					slog.Duration("elapsed", time.Since(start)),
				)
			}
		} else if ctx.Err() == nil {
			logger.ErrorContext(
				ctx,
				"module download failed",
				slog.Duration("elapsed", time.Since(start)),
				slog.Any("error", err),
			)
		}
	}()

	dir, version, err := s.exec(
		ctx,
		"go",
		"list",
		"-m",
		"-json",
		fmt.Sprintf("%s@%s", m.ModulePath, m.ModuleVersion),
	)
	if err != nil {
		return err
	}

	if dir == "" {
		dir, version, err = s.exec(
			ctx,
			"go",
			"mod",
			"download",
			"-json",
			fmt.Sprintf("%s@%s", m.ModulePath, m.ModuleVersion),
		)
		if err != nil {
			return err
		}
	} else {
		cached = true
	}

	return minibus.Send(
		ctx,
		messages.ModuleDownloaded{
			RepoSource:    m.RepoSource,
			RepoID:        m.RepoID,
			ModulePath:    m.ModulePath,
			ModuleVersion: version,
			ModuleDir:     dir,
		},
	)
}

func (s *Supervisor) exec(
	ctx context.Context,
	command string,
	args ...string,
) (dir, version string, err error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = s.Environment
	cmd.Dir = os.TempDir()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	var output struct {
		Dir     string
		Version string
		Error   string
	}

	parseErr := json.
		NewDecoder(&stdout).
		Decode(&output)

	if parseErr == nil && output.Error != "" {
		return "", "", fmt.Errorf(
			"unable to download module: %s",
			output.Error,
		)
	}

	if runErr != nil {
		return "", "", fmt.Errorf(
			"unable to download module: %w\n"+stderr.String(),
			runErr,
		)
	}

	if parseErr != nil {
		return "", "", fmt.Errorf("unable to parse 'go mod download' output: %w", parseErr)
	}

	return output.Dir, output.Version, nil
}
