package agent

func (self *Agent) DeployService(templateNamespace string, templateName string, serviceNamespace string, serviceName string) error {
	if _, problems, err := self.CompileTOSCA(templateNamespace, templateName, serviceNamespace, serviceName); err == nil {
		self.ProcessService(serviceNamespace, serviceName, "schedule")
		return nil
	} else {
		if problems != nil {
			return problems.WithError(nil, false)
		} else {
			return err
		}
	}
}
