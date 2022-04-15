package commands

import (
	"os"

	clientpkg "github.com/tliron/khutulun/client"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

func listArtifacts(namespace string, type_ string) {
	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	artifacts, err := client.ListArtifacts(namespace, type_)
	util.FailOnError(err)
	if len(artifacts) > 0 {
		err = formatpkg.Print(artifacts, format, terminal.Stdout, strict, pretty)
		util.FailOnError(err)
	}
}

func registerArtifact(namespace string, type_ string, args []string) {
	name := args[0]

	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	if len(args) == 2 {
		// TODO: upload entire dir

		file, err := os.Open(args[1])
		util.FailOnError(err)
		util.OnExitError(file.Close)

		err = client.SetArtifact(namespace, type_, name, file)
		util.FailOnError(err)
	} else {
		// TODO stdin
	}
}

func fetchArtifact(namespace string, type_ string, args []string) {
	name := args[0]

	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	err = client.GetArtifact(namespace, type_, name, terminal.Stdout)
	util.FailOnError(err)
}

func delist(namespace string, type_ string, args []string) {
	name := args[0]

	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	err = client.RemoveArtifact(namespace, type_, name)
	util.FailOnError(err)
}
