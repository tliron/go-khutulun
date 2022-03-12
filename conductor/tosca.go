package conductor

import (
	"errors"
	"os"

	"github.com/tliron/kutil/format"
	problemspkg "github.com/tliron/kutil/problems"
	urlpkg "github.com/tliron/kutil/url"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
	"github.com/tliron/puccini/tosca/parser"
)

var parserContext = parser.NewContext()

func (self *Conductor) GetClout(namespace string, serviceName string, coerce bool) (*cloutpkg.Clout, error) {
	if lock, err := self.lockArtifact(namespace, "clout", serviceName, false); err == nil {
		defer lock.Unlock()

		cloutPath := self.getArtifactFile(namespace, "clout", serviceName)
		reconcilerLog.Infof("reading clout: %q", cloutPath)
		if clout, err := cloutpkg.Load(cloutPath, "yaml"); err == nil {
			if coerce {
				problems := problemspkg.NewProblems(nil)
				js.Coerce(clout, problems, self.urlContext, true, "yaml", true, false, false)
				if !problems.Empty() {
					return nil, errors.New(problems.String())
				}
			}

			return clout, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Conductor) CompileTosca(templateNamespace string, templateName string, serviceNamespace string, serviceName string) (*cloutpkg.Clout, *problemspkg.Problems, error) {
	profilePath := self.getArtifactTypeDir(templateNamespace, "profile")
	commonProfilePath := self.getArtifactTypeDir("common", "profile")

	origins := []urlpkg.URL{
		urlpkg.NewFileURL(profilePath, self.urlContext),
		urlpkg.NewFileURL(commonProfilePath, self.urlContext),
	}

	if lock, err := self.lockArtifact(templateNamespace, "template", templateName, false); err == nil {
		defer lock.Unlock()

		templatePath := self.getArtifactFile(templateNamespace, "template", templateName)
		if url_, err := urlpkg.NewValidURL(templatePath, nil, self.urlContext); err == nil {
			if _, serviceTemplate, problems, err := parserContext.Parse(url_, origins, nil, nil, nil); err == nil {
				if clout, err := serviceTemplate.Compile(false); err == nil {
					js.Resolve(clout, problems, self.urlContext, true, "yaml", true, false, false)
					if !problems.Empty() {
						return nil, nil, problems.WithError(nil, false)
					}

					if lock, err := self.lockArtifact(serviceNamespace, "clout", serviceName, true); err == nil {
						defer lock.Unlock()

						cloutPath := self.getArtifactFile(serviceNamespace, "clout", serviceName)
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
				return nil, nil, problems.WithError(err, false)
			}
		} else {
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}
