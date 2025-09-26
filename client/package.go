package client

import (
	contextpkg "context"
	"io"

	"github.com/tliron/commonlog"
	"github.com/tliron/go-khutulun/api"
	"github.com/tliron/go-khutulun/sdk"
	"github.com/tliron/go-kutil/streampackage"
	"github.com/tliron/go-kutil/util"
)

const BUFFER_SIZE = 65536

type PackageIdentifier struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Type      string `json:"type" yaml:"type"`
	Name      string `json:"name" yaml:"name"`
}

func (self *Client) ListPackages(namespace string, type_ string) ([]PackageIdentifier, error) {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	listPackages := api.ListPackages{
		Namespace: namespace,
		Type:      &api.PackageType{Name: type_},
	}

	if client, err := self.client.ListPackages(context, &listPackages); err == nil {
		var identifiers []PackageIdentifier

		for {
			identifier, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return nil, sdk.UnpackGRPCError(err)
				}
			}

			identifiers = append(identifiers, PackageIdentifier{
				Namespace: identifier.Namespace,
				Type:      identifier.Type.Name,
				Name:      identifier.Name,
			})
		}

		return identifiers, nil
	} else {
		return nil, sdk.UnpackGRPCError(err)
	}
}

type PackageFile struct {
	Path       string
	Executable bool
}

func (self *Client) ListPackageFiles(namespace string, type_ string, name string) ([]PackageFile, error) {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	identifier := api.PackageIdentifier{
		Namespace: namespace,
		Type:      &api.PackageType{Name: type_},
		Name:      name,
	}

	if client, err := self.client.ListPackageFiles(context, &identifier); err == nil {
		var packageFiles []PackageFile

		for {
			if packageFile_, err := client.Recv(); err == nil {
				packageFiles = append(packageFiles, PackageFile{
					Path:       packageFile_.Path,
					Executable: packageFile_.Executable,
				})
			} else {
				if err == io.EOF {
					break
				} else {
					return nil, sdk.UnpackGRPCError(err)
				}
			}
		}

		return packageFiles, nil
	} else {
		return nil, sdk.UnpackGRPCError(err)
	}
}

func (self *Client) GetPackageFile(namespace string, type_ string, name string, path string, coerce bool, writer io.Writer) error {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	identifier := api.PackageIdentifier{
		Namespace: namespace,
		Type:      &api.PackageType{Name: type_},
		Name:      name,
	}

	if client, err := self.client.GetPackageFiles(context, &api.GetPackageFiles{Identifier: &identifier, Paths: []string{path}, Coerce: coerce}); err == nil {
		for {
			if content, err := client.Recv(); err == nil {
				if _, err := writer.Write(content.Bytes); err != nil {
					return err
				}
			} else {
				if err == io.EOF {
					break
				} else {
					return sdk.UnpackGRPCError(err)
				}
			}
		}

		return nil
	} else {
		return sdk.UnpackGRPCError(err)
	}
}

func (self *Client) SetPackageFiles(context contextpkg.Context, namespace string, type_ string, name string, streamPackage streampackage.StreamPackage) error {
	defer commonlog.CallAndLogError(streamPackage.Close, "close package file sources", log)

	context, cancel := self.newContextWithTimeout()
	defer cancel()

	if client, err := self.client.SetPackageFiles(context); err == nil {
		identifier := api.PackageIdentifier{
			Namespace: namespace,
			Type:      &api.PackageType{Name: type_},
			Name:      name,
		}

		if err := client.Send(&api.PackageContent{Start: &api.PackageContent_Start{Identifier: &identifier}}); err != nil {
			return sdk.UnpackGRPCError(err)
		}

		buffer := make([]byte, BUFFER_SIZE)

		for {
			stream, err := streamPackage.Next()
			if err != nil {
				return err
			}
			if stream == nil {
				break
			}

			reader, path, executable, err := stream.Open(context)
			if err != nil {
				return err
			}
			reader = util.NewContextualReadCloser(context, reader)

			content := api.PackageContent{
				File: &api.PackageFile{
					Path:       path,
					Executable: executable,
				},
			}

			if err := client.Send(&content); err != nil {
				return sdk.UnpackGRPCError(err)
			}

			for {
				count, err := reader.Read(buffer)
				if count > 0 {
					content = api.PackageContent{Bytes: buffer[:count]}
					if err := client.Send(&content); err != nil {
						return sdk.UnpackGRPCError(err)
					}
				}
				if err != nil {
					if err == io.EOF {
						if err := reader.Close(); err != nil {
							return err
						}
						break
					} else {
						return err
					}
				}
			}
		}

		if _, err := client.CloseAndRecv(); err == io.EOF {
			return nil
		} else {
			return sdk.UnpackGRPCError(err)
		}
	} else {
		return sdk.UnpackGRPCError(err)
	}
}

func (self *Client) RemovePackage(namespace string, type_ string, name string) error {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	identifier := api.PackageIdentifier{
		Namespace: namespace,
		Type:      &api.PackageType{Name: type_},
		Name:      name,
	}

	_, err := self.client.RemovePackage(context, &identifier)
	return sdk.UnpackGRPCError(err)
}
