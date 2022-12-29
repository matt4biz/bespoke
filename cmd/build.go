package main

import "github.com/matt4biz/bespoke"

type BuildCommand struct {
	*App
}

func (cmd *BuildCommand) Run() error {
	runner := bespoke.Runner{Args: cmd.args, Writer: cmd.stdout}

	return runner.Run()
}
