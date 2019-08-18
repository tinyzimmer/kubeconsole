package term

import (
	"bytes"
	"context"
	"io"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/google/go-cmp/cmp"
)

func (c *controller) newLogWindow() (*widgets.List, chan string) {
	ch := make(chan string)
	logs := widgets.NewList()
	logs.Title = logTitle
	x, y := ui.TerminalDimensions()
	logs.SetRect(0, y/2, x, y-3)
	logs.TextStyle = ui.NewStyle(ui.ColorBlue)
	logs.WrapText = true
	logs.SelectedRowStyle.Fg = ui.ColorMagenta
	go func() {
		for {
			select {
			case ev := <-ch:
				if ev == clearEvent {
					c.debug("Clearing current log window")
					logs.Rows = make([]string, 0)
				} else {
					newRows := strings.Split(strings.Replace(ev, "\r\n", "\n", -1), "\n")
					if !cmp.Equal(logs.Rows, newRows) {
						logs.Rows = newRows
						logs.ScrollBottom()
						logs.ScrollPageDown()

						c.mux.Lock()
						ui.Render(logs)
						c.mux.Unlock()
					}
				}
			}
		}
	}()
	return logs, ch
}

func (c *controller) streamLogsToWindow(ctx context.Context, stream io.ReadCloser) {
	defer stream.Close()
	buf := new(bytes.Buffer)
	go asyncCopy(logContext, buf, stream)
	for {
		select {
		case <-ctx.Done():
			c.debug("Got cancel for log stream")
			return
		default:
			c.logChan <- buf.String()
		}
	}
}
