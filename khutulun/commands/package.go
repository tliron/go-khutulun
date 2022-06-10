package commands

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/kutil/terminal"
	formatpkg "github.com/tliron/kutil/transcribe"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
)

func listPackages(namespace string, type_ string) {
	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	identifiers, err := client.ListPackages(namespace, type_)
	util.FailOnError(err)
	if len(identifiers) > 0 {
		err = formatpkg.Print(identifiers, format, terminal.Stdout, strict, pretty)
		util.FailOnError(err)
	}
}

func registerPackage(namespace string, type_ string, args []string) {
	name := args[0]

	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	context := urlpkg.NewContext()
	util.OnExitError(context.Release)

	var packageFiles []clientpkg.SetPackageFile

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
				packageFiles = append(packageFiles, clientpkg.SetPackageFile{
					Reader: reader,
					PackageFile: clientpkg.PackageFile{
						Path:       path[length:],
						Executable: util.IsFileExecutable(stat.Mode()),
					},
				})
			}
			return nil
		})
		util.FailOnError(err)
	} else {
		if isFile {
			var archive string
			if strings.HasSuffix(path, ".zip") || strings.HasSuffix(path, ".csar") {
				archive = "zip"
			} else if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
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
							packageFiles = append(packageFiles, clientpkg.SetPackageFile{
								Reader: reader,
								PackageFile: clientpkg.PackageFile{
									Path:       file.Name,
									Executable: util.IsFileExecutable(file.Mode()),
								},
							})
						}
					}

				case "tgz":
					reader, err := os.Open(path)
					util.FailOnError(err)
					gzipReader, err := gzip.NewReader(reader)
					util.FailOnError(err)
					tarReader := tar.NewReader(gzipReader)

					var packageFiles_ []clientpkg.PackageFile

					for {
						header, err := tarReader.Next()
						if err == io.EOF {
							break
						}
						util.FailOnError(err)
						if header.Typeflag == tar.TypeReg {
							packageFiles_ = append(packageFiles_, clientpkg.PackageFile{
								Path:       header.Name,
								Executable: util.IsFileExecutable(fs.FileMode(header.Mode)),
							})
						}
					}

					err = gzipReader.Close()
					util.FailOnError(err)
					err = reader.Close()
					util.FailOnError(err)

					reader, err = os.Open(path)
					util.FailOnError(err)
					util.OnExitError(reader.Close)
					gzipReader, err = gzip.NewReader(reader)
					util.FailOnError(err)
					util.OnExitError(gzipReader.Close)
					tarReader = tar.NewReader(gzipReader)

					done := func() {
						for {
							header, err := tarReader.Next()
							if err == io.EOF {
								break
							}
							util.FailOnError(err)
							if header.Typeflag == tar.TypeReg {
								break
							}
						}
					}

					for _, packageFile := range packageFiles_ {
						packageFiles = append(packageFiles, clientpkg.SetPackageFile{
							Reader:      tarReader,
							Done:        done,
							PackageFile: packageFile,
						})
					}
				}

			} else {
				// Single file
				reader, err := os.Open(path)
				util.FailOnError(err)
				util.OnExitError(reader.Close)
				packageFiles = append(packageFiles, clientpkg.SetPackageFile{
					Reader: reader,
					PackageFile: clientpkg.PackageFile{
						Path:       filepath.Base(path),
						Executable: util.IsFileExecutable(stat.Mode()),
					},
				})
			}
		} else {
			// Single URL
			reader, err := url.Open()
			util.FailOnError(err)
			util.OnExitError(reader.Close)
			packageFiles = append(packageFiles, clientpkg.SetPackageFile{
				Reader: reader,
				PackageFile: clientpkg.PackageFile{
					Path: filepath.Base(path),
				},
			})
		}
	}

	if len(packageFiles) > 0 {
		err = client.SetPackageFiles(namespace, type_, name, packageFiles)
		util.FailOnError(err)
	}
}

func fetchPackage(namespace string, type_ string, args []string) {
	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	name := args[0]
	if len(args) > 1 {
		path := args[1]
		if (type_ == "service") && terminal.Colorize {
			var buffer strings.Builder
			err = client.GetPackageFile(namespace, type_, name, path, coerce, &buffer)
			util.FailOnError(err)
			err = formatpkg.PrettifyYAML(buffer.String(), terminal.Stdout)
			util.FailOnError(err)
		} else {
			err = client.GetPackageFile(namespace, type_, name, path, coerce, terminal.Stdout)
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
