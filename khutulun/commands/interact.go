package commands

import (
	"os"

	clientpkg "github.com/tliron/go-khutulun/client"
	"github.com/tliron/go-khutulun/sdk"
	"github.com/tliron/go-kutil/exec"
	"github.com/tliron/go-kutil/util"
)

var forwardExitCode bool

func interact(identifier []string, command []string) {
	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	var terminal *exec.Terminal
	if pseudoTerminal {
		terminal, err = exec.NewTerminal()
		util.FailOnError(err)
	}

	environment := map[string]string{"TERM": os.Getenv("TERM")}

	err = client.Interact(identifier, os.Stdin, os.Stdout, os.Stderr, terminal, environment, command...)

	if terminal != nil {
		terminal.Close()
	}

	if details := sdk.InteractionErrorDetails(err); details != nil {
		os.Stderr.Write(details.Stderr)
		if forwardExitCode {
			util.Exit(int(details.ExitCode))
		}
	}

	util.FailOnError(err)
}
