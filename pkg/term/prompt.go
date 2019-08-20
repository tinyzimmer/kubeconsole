package term

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func (c *controller) choicePrompt(title string, choices []string) (sel string) {
	ui.Clear()
	c.renderDefaults()
	prompt := widgets.NewList()
	prompt.Title = title
	prompt.Rows = choices
	prompt.TextStyle = ui.NewStyle(ui.ColorCyan)
	x, y := ui.TerminalDimensions()
	prompt.SetRect(x/3, y/3, (x - x/3), (y - y/3))

	for {
		ui.Render(prompt)

		events := ui.PollEvents()
		e := <-events
		switch e.ID {
		case enter:
			return prompt.Rows[prompt.SelectedRow]
		case "<Escape>":
			return _cancel
		default:
			if q := c.checkCommon(prompt, e.ID); q == _quit {
				return _quit
			}
		}
	}
}
