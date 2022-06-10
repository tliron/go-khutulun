package dashboard

import (
	"github.com/rivo/tview"
)

// UpdateTableFunc signature
func (self *Application) updateServices(table *tview.Table, namespace *tview.DropDown) {
	update := func() {
		namespace_ := getNamespace(namespace)
		if services_, err := self.client.ListPackages(namespace_, "service"); err == nil {
			table.Clear()

			headers := []string{"Name"}
			if namespace_ == "" {
				headers = append([]string{"Namespace"}, headers...)
			}
			SetTableHeader(table, headers...)

			column := 0
			row := 1
			for _, service := range services_ {
				if namespace_ == "" {
					table.SetCell(row, column, tview.NewTableCell(namespaceLabel(service.Namespace)))
					column++
				}
				table.SetCell(row, column, tview.NewTableCell(service.Name))
				row++
				column = 0
			}
		}
	}
	self.updateNamespaces(table, namespace, update)
}

// UpdateTableFunc signature
func (self *Application) updateActivities(table *tview.Table, namespace *tview.DropDown) {
	update := func() {
		namespace_ := getNamespace(namespace)
		if resources, err := self.client.ListResources(namespace_, "", "activity"); err == nil {
			table.Clear()

			headers := []string{"Name", "Service", "Host"}
			if namespace_ == "" {
				headers = append([]string{"Namespace"}, headers...)
			}
			SetTableHeader(table, headers...)

			column := 0
			row := 1
			for _, resource := range resources {
				if namespace_ == "" {
					table.SetCell(row, column, tview.NewTableCell(namespaceLabel(resource.Namespace)))
					column++
				}
				table.SetCell(row, column, tview.NewTableCell(resource.Name))
				column++
				table.SetCell(row, column, tview.NewTableCell(resource.Service))
				column++
				table.SetCell(row, column, tview.NewTableCell(resource.Host))
				row++
				column = 0
			}
		}
	}
	self.updateNamespaces(table, namespace, update)
}

// UpdateTableFunc signature
func (self *Application) updateConnections(table *tview.Table, namespace *tview.DropDown) {
	update := func() {
		namespace_ := getNamespace(namespace)
		if resources, err := self.client.ListResources(namespace_, "", "connection"); err == nil {
			table.Clear()

			headers := []string{"Name"}
			if namespace_ == "" {
				headers = append([]string{"Namespace"}, headers...)
			}
			SetTableHeader(table, headers...)

			column := 0
			row := 1
			for _, resource := range resources {
				if namespace_ == "" {
					table.SetCell(row, column, tview.NewTableCell(namespaceLabel(resource.Namespace)))
					column++
				}
				table.SetCell(row, column, tview.NewTableCell(resource.Name))
				row++
				column = 0
			}
		}
	}
	self.updateNamespaces(table, namespace, update)
}

// UpdateTableFunc signature
func (self *Application) updateDelegates(table *tview.Table, namespace *tview.DropDown) {
	update := func() {
		namespace_ := getNamespace(namespace)
		if delegates, err := self.client.ListPackages(namespace_, "delegate"); err == nil {
			table.Clear()

			headers := []string{"Name"}
			if namespace_ == "" {
				headers = append([]string{"Namespace"}, headers...)
			}
			SetTableHeader(table, headers...)

			column := 0
			row := 1
			for _, delegate := range delegates {
				if namespace_ == "" {
					table.SetCell(row, column, tview.NewTableCell(namespaceLabel(delegate.Namespace)))
					column++
				}
				table.SetCell(row, column, tview.NewTableCell(delegate.Name))
				row++
				column = 0
			}
		}
	}
	self.updateNamespaces(table, namespace, update)
}

// UpdateTableFunc signature
func (self *Application) updateTemplates(table *tview.Table, namespace *tview.DropDown) {
	update := func() {
		namespace_ := getNamespace(namespace)
		if templates, err := self.client.ListPackages(namespace_, "template"); err == nil {
			table.Clear()

			headers := []string{"Name"}
			if namespace_ == "" {
				headers = append([]string{"Namespace"}, headers...)
			}
			SetTableHeader(table, headers...)

			column := 0
			row := 1
			for _, template := range templates {
				if namespace_ == "" {
					table.SetCell(row, column, tview.NewTableCell(namespaceLabel(template.Namespace)))
					column++
				}
				table.SetCell(row, column, tview.NewTableCell(template.Name))
				row++
				column = 0
			}
		}
	}
	self.updateNamespaces(table, namespace, update)
}

// UpdateTableFunc signature
func (self *Application) updateProfiles(table *tview.Table, namespace *tview.DropDown) {
	update := func() {
		namespace_ := getNamespace(namespace)
		if profiles, err := self.client.ListPackages(namespace_, "profile"); err == nil {
			table.Clear()

			headers := []string{"Name"}
			if namespace_ == "" {
				headers = append([]string{"Namespace"}, headers...)
			}
			SetTableHeader(table, headers...)

			column := 0
			row := 1
			for _, profile := range profiles {
				if namespace_ == "" {
					table.SetCell(row, column, tview.NewTableCell(namespaceLabel(profile.Namespace)))
					column++
				}
				table.SetCell(row, column, tview.NewTableCell(profile.Name))
				row++
				column = 0
			}
		}
	}
	self.updateNamespaces(table, namespace, update)
}

// UpdateTableFunc signature
func (self *Application) updateHosts(table *tview.Table, namespace *tview.DropDown) {
	if services_, err := self.client.ListHosts(); err == nil {
		table.Clear()

		SetTableHeader(table, "Name", "gRPC Address")

		row := 1
		for _, service := range services_ {
			table.
				SetCell(row, 0, tview.NewTableCell(service.Name)).
				SetCell(row, 1, tview.NewTableCell(service.GRPCAddress))
			row++
		}
	}
}

func (self *Application) updateNamespaces(table *tview.Table, namespace *tview.DropDown, update func()) {
	if !namespace.HasFocus() {
		options := []string{namespaceLabel("")}
		if namespaces, err := self.client.ListNamespaces(); err == nil {
			for _, option := range namespaces {
				options = append(options, namespaceLabel(option))
			}
		}
		namespace.SetOptions(options, func(text string, index int) {
			self.application.SetFocus(table)
			update()
		})
		current, _ := namespace.GetCurrentOption()
		if current == -1 {
			namespace.SetCurrentOption(0)
		}
	}
	update()
}

func namespaceLabel(namespace string) string {
	switch namespace {
	case "":
		return "(all)"
	case "_":
		return "(default)"
	}
	return namespace
}

func getNamespace(namespace *tview.DropDown) string {
	_, namespace_ := namespace.GetCurrentOption()
	switch namespace_ {
	case "(all)":
		namespace_ = ""
	case "(default)":
		namespace_ = "_"
	}
	return namespace_
}
