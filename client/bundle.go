package client

import (
	"io"
	"os"

	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/util"
)

const BUFFER_SIZE = 65536

type BundleIdentifier struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Type      string `json:"type" yaml:"type"`
	Name      string `json:"name" yaml:"name"`
}

type BundleFile struct {
	Path       string
	Executable bool
}

type SetBundleFile struct {
	BundleFile
	SourcePath string
}

func (self *Client) ListBundles(namespace string, type_ string) ([]BundleIdentifier, error) {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	listBundles := api.ListBundles{
		Namespace: namespace,
		Type:      &api.BundleType{Name: type_},
	}

	if client, err := self.client.ListBundles(context, &listBundles); err == nil {
		var bundles []BundleIdentifier

		for {
			identifier, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return nil, util.UnpackGrpcError(err)
				}
			}

			bundles = append(bundles, BundleIdentifier{
				Namespace: identifier.Namespace,
				Type:      identifier.Type.Name,
				Name:      identifier.Name,
			})
		}

		return bundles, nil
	} else {
		return nil, util.UnpackGrpcError(err)
	}
}

func (self *Client) ListBundleFiles(namespace string, type_ string, name string) ([]BundleFile, error) {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	identifier := api.BundleIdentifier{
		Namespace: namespace,
		Type:      &api.BundleType{Name: type_},
		Name:      name,
	}

	if client, err := self.client.ListBundleFiles(context, &identifier); err == nil {
		var bundleFiles []BundleFile

		for {
			if bundleFile_, err := client.Recv(); err == nil {
				bundleFiles = append(bundleFiles, BundleFile{
					Path:       bundleFile_.Path,
					Executable: bundleFile_.Executable,
				})
			} else {
				if err == io.EOF {
					break
				} else {
					return nil, util.UnpackGrpcError(err)
				}
			}
		}

		return bundleFiles, nil
	} else {
		return nil, util.UnpackGrpcError(err)
	}
}

func (self *Client) GetBundleFile(namespace string, type_ string, name string, path string, writer io.Writer) error {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	identifier := api.BundleIdentifier{
		Namespace: namespace,
		Type:      &api.BundleType{Name: type_},
		Name:      name,
	}

	if client, err := self.client.GetBundleFiles(context, &api.GetBundleFiles{Identifier: &identifier, Paths: []string{path}}); err == nil {
		for {
			if content, err := client.Recv(); err == nil {
				if _, err := writer.Write(content.Bytes); err != nil {
					return err
				}
			} else {
				if err == io.EOF {
					break
				} else {
					return util.UnpackGrpcError(err)
				}
			}
		}

		return nil
	} else {
		return util.UnpackGrpcError(err)
	}
}

func (self *Client) SetBundleFiles(namespace string, type_ string, name string, bundleFiles []SetBundleFile) error {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	if client, err := self.client.SetBundleFiles(context); err == nil {
		identifier := api.BundleIdentifier{
			Namespace: namespace,
			Type:      &api.BundleType{Name: type_},
			Name:      name,
		}

		if err := client.Send(&api.BundleContent{Start: &api.BundleContent_Start{Identifier: &identifier}}); err != nil {
			return util.UnpackGrpcError(err)
		}

		buffer := make([]byte, BUFFER_SIZE)

		for _, bundleFile := range bundleFiles {
			if file, err := os.Open(bundleFile.SourcePath); err == nil {
				content := api.BundleContent{
					File: &api.BundleFile{
						Path:       bundleFile.Path,
						Executable: bundleFile.Executable,
					},
				}

				if err := client.Send(&content); err != nil {
					if err := file.Close(); err != nil {
						log.Errorf("file close: %s", err)
					}
					return util.UnpackGrpcError(err)
				}

				for {
					if count, err := file.Read(buffer); err == nil {
						content = api.BundleContent{Bytes: buffer[:count]}
						if err := client.Send(&content); err != nil {
							if err := file.Close(); err != nil {
								log.Errorf("file close: %s", err)
							}
							return util.UnpackGrpcError(err)
						}
					} else {
						if err == io.EOF {
							break
						} else {
							if err := file.Close(); err != nil {
								log.Errorf("file close: %s", err)
							}
							return util.UnpackGrpcError(err)
						}
					}
				}

				if err := file.Close(); err != nil {
					log.Errorf("file close: %s", err)
				}
			} else {
				return util.UnpackGrpcError(err)
			}
		}

		if _, err := client.CloseAndRecv(); err == io.EOF {
			return nil
		} else {
			return util.UnpackGrpcError(err)
		}
	} else {
		return util.UnpackGrpcError(err)
	}
}

func (self *Client) RemoveBundle(namespace string, type_ string, name string) error {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	identifier := api.BundleIdentifier{
		Namespace: namespace,
		Type:      &api.BundleType{Name: type_},
		Name:      name,
	}

	_, err := self.client.RemoveBundle(context, &identifier)
	return util.UnpackGrpcError(err)
}
