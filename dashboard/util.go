package dashboard

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func GetListMinSize(list *tview.List) (int, int) {
	w := 0
	l := list.GetItemCount()
	for i := 0; i < l; i++ {
		text1, text2 := list.GetItemText(i)
		w_ := len(text1)
		if w_ > w {
			w = w_
		}
		w_ = len(text2)
		if w_ > w {
			w = w_
		}
	}
	return w + 4, l
}

func SetTableHeader(table *tview.Table, headers ...string) {
	for column, header := range headers {
		table.SetCell(0, column, NewHeaderTableCell(header))
	}
}

func NewHeaderTableCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetSelectable(false).
		SetStyle(tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true))
}
