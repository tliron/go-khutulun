package main

import (
	"github.com/tliron/go-khutulun/khutulun/commands"
	"github.com/tliron/go-kutil/util"

	_ "github.com/tliron/commonlog/simple"
)

func main() {
	util.ExitOnSignals()
	commands.Execute()
	util.Exit(0)
}
