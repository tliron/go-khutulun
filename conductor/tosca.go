package conductor

import (
	"os"

	"github.com/tliron/kutil/format"
	problemspkg "github.com/tliron/kutil/problems"
	urlpkg "github.com/tliron/kutil/url"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
	"github.com/tliron/puccini/tosca/normal"
	"github.com/tliron/puccini/tosca/parser"
)

var parserContext = parser.NewContext()

func (self *Conductor) ParseTosca(templateNamespace string, templateName string) (*normal.ServiceTemplate, *problemspkg.Problems, error) {
	profilePath := self.getBundleTypeDir(templateNamespace, "profile")
	commonProfilePath := self.getBundleTypeDir("common", "profile")

	origins := []urlpkg.URL{
		urlpkg.NewFileURL(profilePath, self.urlContext),
		urlpkg.NewFileURL(commonProfilePath, self.urlContext),
	}

	if lock, err := self.lockBundle(templateNamespace, "template", templateName, false); err == nil {
		defer func() {
			if err := lock.Unlock(); err != nil {
				log.Errorf("unlock: %s", err.Error())
			}
		}()

		templatePath := self.getBundleMainFile(templateNamespace, "template", templateName)
		if url, err := urlpkg.NewValidURL(templatePath, nil, self.urlContext); err == nil {
			if _, serviceTemplate, problems, err := parserContext.Parse(url, origins, nil, nil, nil); err == nil {
				return serviceTemplate, problems, nil
			} else {
				if problems != nil {
					return nil, nil, problems.WithError(err, false)
				} else {
					return nil, nil, err
				}
			}
		} else {
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}

func (self *Conductor) CompileTosca(templateNamespace string, templateName string, serviceNamespace string, serviceName string) (*cloutpkg.Clout, *problemspkg.Problems, error) {
	if serviceTemplate, problems, err := self.ParseTosca(templateNamespace, templateName); err == nil {
		if clout, err := serviceTemplate.Compile(false); err == nil {
			js.Resolve(clout, problems, self.urlContext, true, "yaml", true, false, false)
			if !problems.Empty() {
				return nil, nil, problems.WithError(nil, false)
			}

			if lock, err := self.lockBundle(serviceNamespace, "clout", serviceName, true); err == nil {
				defer func() {
					if err := lock.Unlock(); err != nil {
						log.Errorf("unlock: %s", err.Error())
					}
				}()

				cloutPath := self.getBundleMainFile(serviceNamespace, "clout", serviceName)
				log.Infof("writing to %q", cloutPath)
				if file, err := os.OpenFile(cloutPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666); err == nil {
					defer file.Close()

					if err := format.WriteYAML(clout, file, "  ", false); err != nil {
						return nil, nil, err
					}
				} else {
					return nil, nil, err
				}
			} else {
				return nil, nil, err
			}

			return clout, problems, nil
		} else {
			return nil, nil, problems.WithError(err, false)
		}
	} else {
		return nil, nil, err
	}
}
