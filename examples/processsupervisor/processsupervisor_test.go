package processsupervisor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestSupervisorCapturesCommandOutput(t *testing.T) {
	supervisor := Supervisor{
		Factory: func(ctx context.Context) *exec.Cmd {
			return helperCommand(ctx, "emit")
		},
	}

	result, err := supervisor.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if strings.TrimSpace(result.Stdout) != "ready" {
		t.Fatalf("stdout = %q, want ready", result.Stdout)
	}
	if strings.TrimSpace(result.Stderr) != "warn: warm path" {
		t.Fatalf("stderr = %q, want warning output", result.Stderr)
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", result.ExitCode)
	}
}

func TestSupervisorReturnsErrorWhenContextCancelsProcess(t *testing.T) {
	supervisor := Supervisor{
		Factory: func(ctx context.Context) *exec.Cmd {
			return helperCommand(ctx, "sleep")
		},
		WaitDelay: 25 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()

	result, err := supervisor.Run(ctx)
	if err == nil {
		t.Fatalf("Run error = nil, want cancellation-related error")
	}
	if result.Stdout == "" {
		t.Fatalf("expected partial stdout before cancellation")
	}
	if result.ExitCode == 0 {
		t.Fatalf("exit code = %d, want non-zero after forced stop", result.ExitCode)
	}
}

func TestSupervisorRejectsNilFactory(t *testing.T) {
	_, err := (Supervisor{}).Run(context.Background())
	if !errors.Is(err, ErrNilFactory) {
		t.Fatalf("Run error = %v, want ErrNilFactory", err)
	}
}

func TestSupervisorRejectsNilCommand(t *testing.T) {
	supervisor := Supervisor{
		Factory: func(context.Context) *exec.Cmd {
			return nil
		},
	}

	_, err := supervisor.Run(context.Background())
	if !errors.Is(err, ErrNilCommand) {
		t.Fatalf("Run error = %v, want ErrNilCommand", err)
	}
}

func helperCommand(ctx context.Context, mode string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestProcessHelper", "--", mode)
	cmd.Env = append(os.Environ(), "GO_WANT_PROCESS_HELPER=1")
	return cmd
}

func TestProcessHelper(t *testing.T) {
	if os.Getenv("GO_WANT_PROCESS_HELPER") != "1" {
		return
	}

	args := os.Args
	idx := 0
	for i, arg := range args {
		if arg == "--" {
			idx = i + 1
			break
		}
	}
	if idx <= 0 || idx >= len(args) {
		fmt.Fprintln(os.Stderr, "missing helper mode")
		os.Exit(2)
	}

	switch args[idx] {
	case "emit":
		fmt.Fprintln(os.Stdout, "ready")
		fmt.Fprintln(os.Stderr, "warn: warm path")
		os.Exit(0)
	case "sleep":
		fmt.Fprintln(os.Stdout, "starting long task")
		time.Sleep(5 * time.Second)
		fmt.Fprintln(os.Stdout, "finished")
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown helper mode %q\n", args[idx])
		os.Exit(2)
	}
}
