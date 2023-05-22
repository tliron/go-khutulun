package agent

import (
	"github.com/tliron/commonlog"
	"github.com/tliron/exturl"
	problemspkg "github.com/tliron/kutil/problems"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
	"github.com/tliron/puccini/tosca/normal"
	"github.com/tliron/puccini/tosca/parser"
)

func (self *Agent) ParseTOSCA(templateNamespace string, templateName string) (*normal.ServiceTemplate, *problemspkg.Problems, error) {
	parserContext := parser.NewContext()

	profilePath := self.state.GetPackageTypeDir(templateNamespace, "profile")
	commonProfilePath := self.state.GetPackageTypeDir("common", "profile")

	// TODO: lock *all* profiles

	origins := []exturl.URL{
		exturl.NewFileURL(profilePath, self.urlContext),
		exturl.NewFileURL(commonProfilePath, self.urlContext),
	}

	if lock, err := self.state.LockPackage(templateNamespace, "template", templateName, false); err == nil {
		defer commonlog.CallAndLogError(lock.Unlock, "unlock", log)

		templatePath := self.state.GetPackageMainFile(templateNamespace, "template", templateName)
		if url, err := exturl.NewValidURL(templatePath, nil, self.urlContext); err == nil {
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

func (self *Agent) CompileTOSCA(templateNamespace string, templateName string, serviceNamespace string, serviceName string) (*cloutpkg.Clout, *problemspkg.Problems, error) {
	if serviceTemplate, problems, err := self.ParseTOSCA(templateNamespace, templateName); err == nil {
		if clout, err := serviceTemplate.Compile(); err == nil {
			js.Resolve(clout, problems, self.urlContext, true, "yaml", true, false)
			if !problems.Empty() {
				return nil, nil, problems.WithError(nil, false)
			}

			if lock, err := self.state.LockPackage(serviceNamespace, "service", serviceName, true); err == nil {
				defer commonlog.CallAndLogError(lock.Unlock, "unlock", log)

				if err := self.state.SaveServiceClout(serviceNamespace, serviceName, clout); err == nil {
					return clout, problems, nil
				} else {
					return nil, nil, err
				}
			} else {
				return nil, nil, err
			}
		} else {
			return nil, nil, problems.WithError(err, false)
		}
	} else {
		return nil, nil, err
	}
}
