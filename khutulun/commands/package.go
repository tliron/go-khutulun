package commands

import (
	contextpkg "context"
	"os"
	"strings"

	"github.com/tliron/exturl"
	"github.com/tliron/go-transcribe"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/kutil/streampackage"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

func listPackages(namespace string, type_ string) {
	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	identifiers, err := client.ListPackages(namespace, type_)
	util.FailOnError(err)
	if len(identifiers) > 0 {
		err = Transcriber().Write(identifiers)
		util.FailOnError(err)
	}
}

func registerPackage(context contextpkg.Context, namespace string, type_ string, args []string) {
	name := args[0]

	switch unpack {
	case "tgz", "zip":
	case "auto":
		if len(args) == 2 {
			path := args[1]
			if strings.HasSuffix(path, ".tar") {
				unpack = "tar"
			} else if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
				unpack = "tgz"
			} else if strings.HasSuffix(path, ".zip") || strings.HasSuffix(path, ".csar") {
				unpack = "zip"
			} else {
				unpack = ""
			}
		}
	case "false":
		unpack = ""
	default:
		util.Failf("\"--unpack\" must be \"tar\", \"tgz\", \"zip\", \"auto\" or \"false\": %s", unpack)
	}

	urlContext := exturl.NewContext()
	util.OnExitError(urlContext.Release)

	var url exturl.URL
	var err error

	if len(args) == 2 {
		url, err = urlContext.NewValidURL(context, args[1], nil)
	} else {
		path := type_
		switch type_ {
		case "profile", "template":
			path += ".yaml"
		}
		url, err = urlContext.ReadToInternalURL(path, os.Stdin)
	}
	util.FailOnError(err)

	streamPackage, err := streampackage.NewStreamPackage(context, url, unpack)
	util.FailOnError(err)

	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	err = client.SetPackageFiles(context, namespace, type_, name, streamPackage)
	util.FailOnError(err)
}

func fetchPackage(namespace string, type_ string, args []string) {
	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	name := args[0]
	if len(args) > 1 {
		path := args[1]
		if (type_ == "service") && terminal.ColorizeStdout {
			var buffer strings.Builder
			err = client.GetPackageFile(namespace, type_, name, path, coerce, &buffer)
			util.FailOnError(err)
			err = transcribe.ColorizeYAML(buffer.String(), os.Stdout)
			util.FailOnError(err)
		} else {
			err = client.GetPackageFile(namespace, type_, name, path, coerce, os.Stdout)
			util.FailOnError(err)
		}
	} else {
		files, err := client.ListPackageFiles(namespace, type_, name)
		util.FailOnError(err)
		for _, file := range files {
			terminal.Println(file.Path)
		}
	}
}

func delistPackage(namespace string, type_ string, args []string) {
	name := args[0]

	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	err = client.RemovePackage(namespace, type_, name)
	util.FailOnError(err)
}
