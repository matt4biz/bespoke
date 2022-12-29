package main

type DebugCommand struct {
	*App
}

func (cmd *DebugCommand) Run() int {
	return 1
}
