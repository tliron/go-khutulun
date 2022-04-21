package commands

import (
	"archive/zip"
	"io/fs"
	"os"
	"path/filepath"

	clientpkg "github.com/tliron/khutulun/client"
	khutulunutil "github.com/tliron/khutulun/util"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/terminal"
	urlpkg "github.com/tliron/kutil/url"
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

	context := urlpkg.NewContext()
	util.OnExitError(context.Release)

	var bundleFiles []clientpkg.SetBundleFile

	var url urlpkg.URL
	var path string
	var isFile bool
	var isDir bool
	var stat os.FileInfo

	if len(args) == 2 {
		path = args[1]
		url, err = urlpkg.NewValidURL(path, nil, context)
		util.FailOnError(err)
	} else {
		switch type_ {
		case "profile", "template":
			path = type_ + ".yaml"
			url, err = urlpkg.ReadToInternalURLFromStdin("yaml")
			util.FailOnError(err)

		default:
			path = type_
			url, err = urlpkg.ReadToInternalURLFromStdin("")
			util.FailOnError(err)
		}
	}

	if _, isFile = url.(*urlpkg.FileURL); isFile {
		stat, err = os.Stat(path)
		util.FailOnError(err)
		isDir = stat.IsDir()
	}

	if isDir {
		// All files in directory
		length := len(path)
		err = filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
			if !entry.IsDir() {
				stat, err = os.Stat(path)
				util.FailOnError(err)
				reader, err := os.Open(path)
				util.FailOnError(err)
				util.OnExitError(reader.Close)
				bundleFiles = append(bundleFiles, clientpkg.SetBundleFile{
					Reader: reader,
					BundleFile: clientpkg.BundleFile{
						Path:       path[length:],
						Executable: khutulunutil.IsExecutable(stat.Mode()),
					},
				})
			}
			return nil
		})
		util.FailOnError(err)
	} else {
		if isFile {
			var archive string
			switch filepath.Ext(path) {
			case ".zip", ".csar":
				archive = "zip"
			case ".tar.gz", ".tgz":
				archive = "tgz"
			}

			var unpack_ bool
			switch unpack {
			case "auto":
				unpack_ = (archive != "")
			case "false":
			default:
				util.Failf("\"--unpack\" must be \"auto\" or \"false\": %s", unpack)
			}

			if unpack_ {
				// All files in archive
				switch archive {
				case "zip":
					zipReader, err := zip.OpenReader(path)
					util.FailOnError(err)
					util.OnExitError(zipReader.Close)
					for _, file := range zipReader.File {
						if !file.FileInfo().IsDir() {
							reader, err := file.Open()
							util.FailOnError(err)
							util.OnExitError(reader.Close)
							bundleFiles = append(bundleFiles, clientpkg.SetBundleFile{
								Reader: reader,
								BundleFile: clientpkg.BundleFile{
									Path:       file.Name,
									Executable: khutulunutil.IsExecutable(file.Mode()),
								},
							})
						}
					}
				}
			} else {
				// Single file
				reader, err := os.Open(path)
				util.FailOnError(err)
				util.OnExitError(reader.Close)
				bundleFiles = append(bundleFiles, clientpkg.SetBundleFile{
					Reader: reader,
					BundleFile: clientpkg.BundleFile{
						Path:       filepath.Base(path),
						Executable: khutulunutil.IsExecutable(stat.Mode()),
					},
				})
			}
		} else {
			// Single URL
			reader, err := url.Open()
			util.FailOnError(err)
			util.OnExitError(reader.Close)
			bundleFiles = append(bundleFiles, clientpkg.SetBundleFile{
				Reader: reader,
				BundleFile: clientpkg.BundleFile{
					Path: filepath.Base(path),
				},
			})
		}
	}

	if len(bundleFiles) > 0 {
		err = client.SetBundleFiles(namespace, type_, name, bundleFiles)
		util.FailOnError(err)
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
