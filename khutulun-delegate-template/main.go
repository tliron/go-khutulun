package main

import (
	"os"

	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	"github.com/tliron/kutil/util"
)

func main() {
	util.ExitOnSIGTERM()
	sdk.ConfigureDelegateLogging(1)
	host, _ := os.Hostname()
	server := delegate.NewDelegatePluginServer(&Delegate{host: host})
	server.Start()
	util.Exit(0)
}
