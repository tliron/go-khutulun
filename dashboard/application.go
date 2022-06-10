package dashboard

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	clientpkg "github.com/tliron/khutulun/client"
)

type UpdateTableFunc func(*tview.Table, *tview.DropDown)

//
// Application
//

type Application struct {
	client    *clientpkg.Client
	frequency time.Duration

	application *tview.Application
	menu        *tview.List
	pages       *tview.Pages
	views       map[string]tview.Primitive
	ticker      *Ticker
}

func NewApplication(client *clientpkg.Client, frequency time.Duration) *Application {
	self := Application{
		client:      client,
		frequency:   frequency,
		application: tview.NewApplication(),
		menu:        tview.NewList(),
		pages:       tview.NewPages(),
		views:       make(map[string]tview.Primitive),
	}

	self.application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				self.application.Stop()
				return nil
			}
		}
		return event
	})

	self.menu.
		ShowSecondaryText(false).
		SetShortcutColor(tcell.ColorBlue).
		SetDoneFunc(self.application.Stop).
		SetBorder(true).
		SetTitle("Khutulun")

	self.AddTableView("home", "Home", 'm', false, nil)
	self.AddTableView("services", "Services", 's', true, self.updateServices)
	self.AddTableView("activities", "Activities", 'a', true, self.updateActivities)
	self.AddTableView("connections", "Connections", 'c', true, self.updateConnections)
	self.AddTableView("storage", "Storage", 's', true, nil)
	self.AddTableView("delegates", "Delegates", 'd', true, self.updateDelegates)
	self.AddTableView("templates", "Templates", 't', true, self.updateTemplates)
	self.AddTableView("profiles", "Profiles", 'p', true, self.updateProfiles)
	self.AddTableView("hosts", "Hosts", 'h', false, self.updateHosts)
	self.AddTableView("users", "Users", 'u', false, nil)
	self.menu.AddItem("Quit", "", 'q', self.application.Stop)
	self.pages.SwitchToPage("home")

	menuWidth, _ := GetListMinSize(self.menu)
	layout := tview.NewFlex().
		AddItem(self.menu, menuWidth+2, 0, true).
		AddItem(self.pages, 0, 1, false)

	self.application.
		SetRoot(layout, true).
		EnableMouse(true).
		SetFocus(layout)

	return &self
}

func (self *Application) AddTableView(name string, title string, key rune, withNamespace bool, updateTable UpdateTableFunc) {
	table := tview.NewTable().
		SetBorders(true)
	table.
		Select(1, 0).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEscape:
				table.SetSelectable(false, false)
				self.application.SetFocus(self.menu)
			case tcell.KeyEnter:
				// TODO
			}
		})

	view := tview.NewFlex().SetDirection(tview.FlexRow)
	view.
		SetBlurFunc(func() {
			if self.ticker != nil {
				self.ticker.Stop()
				self.ticker = nil
			}
		})

	var namespace *tview.DropDown
	if withNamespace {
		namespace = tview.NewDropDown().
			SetLabel("(n) Namespace: ").
			SetLabelColor(tcell.ColorBlue).
			SetDoneFunc(func(key tcell.Key) {
				self.application.SetFocus(table)
			})
		view.AddItem(namespace, 1, 0, false)
		table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyRune:
				switch event.Rune() {
				case 'n':
					self.application.SetFocus(namespace)
					return nil
				}
			}
			return event
		})
	}

	view.AddItem(table, 0, 1, true)
	view.
		SetBorder(true).
		SetTitle(title)

	self.views[name] = view
	self.pages.AddPage(name, view, true, false)
	self.menu.AddItem(title, "", key, func() {
		self.pages.SwitchToPage(name)
		self.application.SetFocus(table)
		table.SetSelectable(true, false)
		if updateTable != nil {
			self.ticker = NewTicker(self.application, self.frequency, func() {
				updateTable(table, namespace)
			})
			self.ticker.Start()
		}
	})
}
