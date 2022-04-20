package commands

import (
	"io/fs"
	"os"
	"path/filepath"

	clientpkg "github.com/tliron/khutulun/client"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

func listBundles(namespace string, type_ string) {
	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	identifiers, err := client.ListBundles(namespace, type_)
	util.FailOnError(err)
	if len(identifiers) > 0 {
		err = formatpkg.Print(identifiers, format, terminal.Stdout, strict, pretty)
		util.FailOnError(err)
	}
}

func registerBundle(namespace string, type_ string, args []string) {
	name := args[0]

	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	if len(args) == 2 {
		path := args[1]
		stat, err := os.Stat(path)
		util.FailOnError(err)

		var bundleFiles []clientpkg.SetBundleFile
		if stat.IsDir() {
			// Gather all files in directory
			length := len(path)
			err = filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
				if !entry.IsDir() {
					stat, err = os.Stat(path)
					util.FailOnError(err)
					bundleFiles = append(bundleFiles, clientpkg.SetBundleFile{
						SourcePath: path,
						BundleFile: clientpkg.BundleFile{
							Path:       path[length:],
							Executable: stat.Mode()&0100 != 0,
						},
					})
				}
				return nil
			})
			util.FailOnError(err)
		} else {
			// Single file
			// TODO: support URLs?
			// TODO: handle archives
			bundleFiles = append(bundleFiles, clientpkg.SetBundleFile{
				SourcePath: path,
				BundleFile: clientpkg.BundleFile{
					Path:       filepath.Base(path),
					Executable: stat.Mode()&0100 != 0,
				},
			})
		}

		err = client.SetBundleFiles(namespace, type_, name, bundleFiles)
		util.FailOnError(err)
	} else {
		// TODO stdin
	}
}

func fetchBundle(namespace string, type_ string, args []string) {
	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	name := args[0]
	if len(args) > 1 {
		path := args[1]

		err = client.GetBundleFile(namespace, type_, name, path, terminal.Stdout)
		util.FailOnError(err)
	} else {
		files, err := client.ListBundleFiles(namespace, type_, name)
		util.FailOnError(err)
		for _, file := range files {
			terminal.Println(file.Path)
		}
	}
}

func delist(namespace string, type_ string, args []string) {
	name := args[0]

	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	err = client.RemoveBundle(namespace, type_, name)
	util.FailOnError(err)
}
