package commands

import (
	"os"

	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/khutulun/sdk"
	"github.com/tliron/kutil/exec"
	terminalpkg "github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
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

	err = client.Interact(identifier, os.Stdin, terminalpkg.Stdout, terminalpkg.Stderr, terminal, environment, command...)

	if terminal != nil {
		terminal.Close()
	}

	if details := sdk.InteractionErrorDetails(err); details != nil {
		terminalpkg.Stderr.Write(details.Stderr)
		if forwardExitCode {
			util.Exit(int(details.ExitCode))
		}
	}

	util.FailOnError(err)
}
