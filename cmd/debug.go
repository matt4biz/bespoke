package main

import "fmt"

type DebugCommand struct {
	*App
}

func (cmd *DebugCommand) Run() error {
	return fmt.Errorf("not implemented")
}
