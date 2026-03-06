package processsupervisor

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"
)

var (
	ErrNilFactory = errors.New("command factory must not be nil")
	ErrNilCommand = errors.New("command factory returned nil")
)

type CommandFactory func(context.Context) *exec.Cmd

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type Supervisor struct {
	Factory   CommandFactory
	WaitDelay time.Duration
}

func (s Supervisor) Run(ctx context.Context) (Result, error) {
	if s.Factory == nil {
		return Result{}, ErrNilFactory
	}

	cmd := s.Factory(ctx)
	if cmd == nil {
		return Result{}, ErrNilCommand
	}

	if s.WaitDelay > 0 {
		cmd.WaitDelay = s.WaitDelay
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode(cmd.ProcessState),
	}, err
}

type exitCodeProvider interface {
	ExitCode() int
}

func exitCode(state any) int {
	if state == nil {
		return -1
	}
	if provider, ok := state.(exitCodeProvider); ok {
		return provider.ExitCode()
	}
	return -1
}
