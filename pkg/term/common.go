package term

import (
	"fmt"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const (
	mainTitle    = " Kubeconsole "
	serverTitle  = " Server "
	serverFormat = "Connected to: %s"
	detailsTitle = " Details "
	logTitle     = " Logs "
	helpTitle    = " Help "
	helpText     = "'q' quit  'r' refresh  'e' exec/pod  <tab> switch panes"
	consoleTitle = " Console "
	execTitle    = " Exec  Ctrl-D to exit "
	errorTitle   = "ERROR"
	clearEvent   = "CLEAR"
)

var tabPanes = []string{"[N]amespaces", "[P]ods", "[C]onsole"}

func newErrorWindow() *widgets.Paragraph {
	pane := widgets.NewParagraph()
	pane.Title = errorTitle
	pane.WrapText = true
	pane.TextStyle = ui.NewStyle(ui.ColorRed)
	x, y := ui.TerminalDimensions()
	pane.SetRect(x/4, y/3, (x - x/4), (y - y/3))
	return pane
}

func newNavWindow() *widgets.TabPane {
	pane := widgets.NewTabPane(tabPanes...)
	pane.Title = mainTitle
	x, _ := ui.TerminalDimensions()
	pane.SetRect(0, 0, x/2, 3)
	return pane
}

func newAPIServerWindow(host string) *widgets.Paragraph {
	pane := widgets.NewParagraph()
	pane.Title = serverTitle
	pane.Text = fmt.Sprintf(serverFormat, host)
	pane.TextStyle = ui.NewStyle(ui.ColorGreen)
	x, _ := ui.TerminalDimensions()
	pane.SetRect(x/2, 0, x, 3)
	return pane
}

func newHelpWindow() *widgets.Paragraph {
	par := widgets.NewParagraph()
	par.Text = helpText
	par.Title = helpTitle
	x, y := ui.TerminalDimensions()
	par.SetRect(0, y-3, x, y)
	return par
}

func newDetailsWindow() (*widgets.Paragraph, chan string) {
	ch := make(chan string)
	par := widgets.NewParagraph()
	par.Title = detailsTitle
	par.WrapText = false
	x, y := ui.TerminalDimensions()
	par.SetRect(x/2, 3, x, y/2)
	go func() {
		for {
			select {
			case ev := <-ch:
				par.Text = ev
				ui.Render(par)
			}
		}
	}()
	return par, ch
}

func newConsoleWindow() *widgets.List {
	par := widgets.NewList()
	par.Title = consoleTitle
	par.TextStyle = ui.NewStyle(ui.ColorWhite)
	par.SelectedRowStyle = ui.NewStyle(ui.ColorWhite)
	par.WrapText = true
	x, y := ui.TerminalDimensions()
	par.SetRect(0, 3, x, y-3)
	return par
}

func newExecWindow() *widgets.Paragraph {
	//ex := widgets.NewList()
	ex := widgets.NewParagraph()
	ex.Title = execTitle
	x, y := ui.TerminalDimensions()
	ex.SetRect(0, y/2, x, y-3)
	ex.TextStyle = ui.NewStyle(ui.ColorWhite)
	ex.WrapText = true
	//ex.SelectedRowStyle = ui.NewStyle(ui.ColorWhite)
	return ex
}
