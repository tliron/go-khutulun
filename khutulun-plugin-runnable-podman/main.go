package main

import (
	"os"

	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"

	_ "github.com/tliron/kutil/logging/simple"
)

func main() {
	util.ExitOnSIGTERM()
	logging.Configure(1, nil)
	host, _ := os.Hostname()
	server := delegate.NewDelegatePluginServer(&Delegate{host: host})
	server.Start()
	util.Exit(0)
}
