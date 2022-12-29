package main

import "github.com/matt4biz/bespoke"

type BuildCommand struct {
	*App
}

func (cmd *BuildCommand) Run() int {
	return bespoke.Run(cmd.args)
}
