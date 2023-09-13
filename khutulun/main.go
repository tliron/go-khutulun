package main

import (
	"github.com/tliron/khutulun/khutulun/commands"
	"github.com/tliron/kutil/util"

	_ "github.com/tliron/commonlog/simple"
)

func main() {
	util.ExitOnSignals()
	commands.Execute()
	util.Exit(0)
}
