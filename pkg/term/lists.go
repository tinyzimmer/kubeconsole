package term

import (
	"fmt"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const (
	namespaceTitle   = " Namespaces "
	namespaceLoading = "Loading namespaces..."

	podsTitle   = " Pods "
	podsLoading = "Loading pods..."
)

func (c *controller) newNamespaceList() (l *widgets.List) {
	l = widgets.NewList()
	l.Title = namespaceTitle

	rowChan := make(chan []string)

	go func() {
		rows, err := c.factory.ListNamespaces()
		if err != nil {
			c.errorChan <- newErrorWithStack(err)
			return
		}
		rowChan <- rows
	}()

	go func() {
		select {
		case newRows := <-rowChan:
			l.Rows = newRows
			c.mux.Lock()
			ui.Render(l)
			c.mux.Unlock()
			return
		}
	}()

	l.Rows = []string{namespaceLoading}
	l.TextStyle = ui.NewStyle(ui.ColorCyan)
	l.WrapText = false

	x, y := ui.TerminalDimensions()
	l.SetRect(x/3, y/3, (x - x/3), (y - y/3))
	return
}

func (c *controller) newPodList(ch chan string) (l *widgets.List) {
	l = widgets.NewList()
	l.Title = podsTitle

	go func() {
		for {
			select {
			case selection := <-ch:
				l.Rows = []string{podsLoading}
				ui.Render(l)
				rows, err := c.factory.ListPods(selection)
				if err != nil {
					c.errorChan <- newErrorWithStack(err)
					return
				}
				l.Title = fmt.Sprintf(" %s   Namespace: %s ", podsTitle, selection)
				l.Rows = rows

				c.mux.Lock()
				ui.Render(l)
				c.mux.Unlock()
			}
		}
	}()

	l.Rows = []string{}
	l.TextStyle = ui.NewStyle(ui.ColorCyan)
	l.WrapText = true
	x, y := ui.TerminalDimensions()
	l.SetRect(0, 3, x/2, y/2)
	return
}
