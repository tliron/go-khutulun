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
	server := delegate.NewDelegateGRPCServer(&Delegate{host: host})
	err := server.Start("tcp6", "::1", 8250)
	util.FailOnError(err)
}
