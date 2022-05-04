package main

import (
	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"

	_ "github.com/tliron/kutil/logging/simple"
)

func main() {
	util.ExitOnSIGTERM()
	logging.Configure(1, nil)
	server := delegate.NewDelegatePluginServer(new(Delegate))
	server.Start()
	util.Exit(0)
}
