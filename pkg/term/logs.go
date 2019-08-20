package term

import (
	"bytes"
	"context"
	"fmt"
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

func (c *controller) tailPod() {
	if empty(c.podList) {
		return
	}
	podName := c.getSelectedPod()
	pod, err := c.factory.GetPod(c.currentNamespace, podName)
	if err != nil {
		c.errorChan <- newErrorWithStack(err)
		return
	}
	if len(pod.Spec.Containers) == 1 {
		c.startLogStream(podName, "")
	}
}

func (c *controller) startLogStream(pod, container string) {
	c.resetLogWindow()
	logsPaused = false
	c.logChan <- clearEvent
	c.logChan <- fmt.Sprintf("Fetching logs for %s...\n", pod)
	if stream, err := c.factory.GetLogStream(c.currentNamespace, pod, container, logContext); err != nil {
		c.errorChan <- newErrorWithStack(err)
		return
	} else {
		c.debug(fmt.Sprintf("Starting log stream for %s", pod))
		go c.streamLogsToWindow(logContext, stream)
	}
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
