package main

import (
	"github.com/tliron/khutulun/khutulun-agent/commands"
	"github.com/tliron/kutil/util"
)

func main() {
	util.ExitOnSignals()
	commands.Execute()
	util.Exit(0)
}
