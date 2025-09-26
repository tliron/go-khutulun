package main

import (
	"github.com/tliron/go-khutulun/khutulun-agent/commands"
	"github.com/tliron/go-kutil/util"
)

func main() {
	util.ExitOnSignals()
	commands.Execute()
	util.Exit(0)
}
