package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

var (
	ErrUsage          = errors.New("usage")
	ErrUnknownCommand = errors.New("unknown command")
)

type App struct {
	args    []string
	version string
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
}

type Command interface {
	Run() error
}

func runApp(args []string, version string, stdin io.ReadCloser, stdout, stderr io.WriteCloser) int {
	a := App{version: version, stdin: stdin, stdout: stdout, stderr: stderr}

	if err := a.fromArgs(args); err != nil {
		if errors.Is(err, ErrUsage) {
			return 0
		}

		fmt.Fprintln(stderr, err)

		return 1
	}

	cmd, err := a.getCommand()

	if err != nil {
		fmt.Fprintln(stderr, err)
		return -1
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(stderr, "%s\n", err)
		return 1
	}

	return 0
}

func (a *App) usage() {
	msg := strings.TrimSpace(`
bespoke: a tool to run kustomize with env variables already substituted.

This allows resource names, patch targets, etc. to have variables in their
names in the kustomization file.
	`)

	fmt.Fprintln(a.stderr, msg)
}

func (a *App) fromArgs(args []string) error {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	help := fs.Bool("h", false, "")

	fs.Usage = a.usage

	if err := fs.Parse(args); err != nil {
		return ErrUsage
	} else if *help {
		a.usage()
		return ErrUsage
	}

	a.args = fs.Args()

	return nil
}

func (a *App) getCommand() (Command, error) {
	if len(a.args) == 0 {
		return nil, ErrUnknownCommand
	}

	s := a.args[0]
	a.args = a.args[1:]

	switch s {
	case "build":
		return &BuildCommand{a}, nil

	case "debug":
		return &DebugCommand{a}, nil
	}

	return nil, fmt.Errorf("%s: %w", s, ErrUnknownCommand)
}
