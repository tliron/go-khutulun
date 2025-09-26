package main

import (
	"os"

	"github.com/tliron/go-khutulun/delegate"
	"github.com/tliron/go-khutulun/sdk"
	"github.com/tliron/go-kutil/util"
)

func main() {
	util.ExitOnSignals()
	sdk.ConfigureDelegateLogging(1)
	host, _ := os.Hostname()
	server := delegate.NewDelegatePluginServer(&Delegate{host: host})
	server.Start()
	//server := delegate.NewDelegateGRPCServer(&Delegate{host: host})
	//err := server.Start("tcp6", "::1", 8250)
	//util.FailOnError(err)
	util.Exit(0)
}
