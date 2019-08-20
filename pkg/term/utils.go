package term

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"runtime/debug"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type errorWithStack struct {
	error

	errMsg   string
	errStack string
}

func (e *errorWithStack) Error() string {
	return e.errMsg
}

func (e *errorWithStack) Stack() string {
	return e.errStack
}

func (c *controller) focusScroll(focus *widgets.List, direction string) {
	switch direction {

	case up:
		if focus == c.detailsWindow {
			focus.ScrollPageUp()
		} else {
			focus.ScrollUp()
		}

	case down:
		if focus == c.detailsWindow {
			focus.ScrollPageDown()
		} else {
			focus.ScrollDown()
		}

	case pageUp:
		focus.ScrollPageUp()

	case pageDown:
		focus.ScrollPageDown()

	case home:
		focus.ScrollTop()

	case end:
		focus.ScrollBottom()

	}
}

func (c *controller) setupIfLogWindow() {
	if focus == c.logWindow {
		cancelIfNotNil(logCancel)
		if !logsPaused {
			c.logWindow.Title = fmt.Sprintf("%s   PAUSED: Press <enter> to resume ", c.logWindow.Title)
			c.mux.Lock()
			ui.Render(c.logWindow)
			c.mux.Unlock()
			logsPaused = true
		}
	}
}

func cancelIfNotNil(f func()) {
	if f != nil {
		f()
	}
}

func newErrorWithStack(err error) (serr *errorWithStack) {
	serr = &errorWithStack{}
	serr.errMsg = fmt.Sprintf("ERROR: %v", err.Error())
	serr.errStack = string(debug.Stack())
	return serr
}

func (c *controller) getPodDetails(pod string) {
	details, err := c.factory.GetPod(c.currentNamespace, pod)
	if err != nil {
		c.errorChan <- newErrorWithStack(err)
	} else {
		var buf bytes.Buffer
		t := template.Must(template.New("pod-details").Parse(podDetailsTemplate))
		err = t.Execute(&buf, details)
		if err != nil {
			c.errorChan <- newErrorWithStack(err)
		} else {
			c.detailsChan <- buf.String()
		}
	}
}

func (c *controller) selectPod() {
	if empty(c.podList) {
		return
	}
	pod := c.getSelectedPod()
	c.debug(fmt.Sprintf("Fetching details for %s", pod))
	c.detailsChan <- fmt.Sprintf("Loading details for %s...\n", pod)
	go c.getPodDetails(pod)
}

func (c *controller) switchPane() {
	focus.Title = strings.Replace(focus.Title, " * ", "", 1)
	if focus == c.podList {
		focus = c.detailsWindow
	} else if focus == c.detailsWindow {
		focus = c.logWindow
	} else {
		focus = c.podList
	}
	focus.Title = fmt.Sprintf(" * %s ", focus.Title)
	c.mux.Lock()
	ui.Render(focus)
	c.mux.Unlock()
}

func (c *controller) displayNamespaceList() {
	cancelIfNotNil(logCancel)
	c.navWindow.FocusLeft()
	ui.Clear()
	c.renderDefaults()
	ch := make(chan string)
	go func() {
		newCh := make(chan string)
		select {
		case sel := <-ch:
			c.podList = c.newPodList(newCh)
			newCh <- sel
			return
		}
	}()
	c.pollNamespaces(ch)
	return
}

func (c *controller) getSelectedPod() string {
	if empty(c.podList) {
		return ""
	}
	return c.podList.Rows[c.podList.SelectedRow]
}

func empty(s *widgets.List) bool {
	return len(s.Rows) == 0
}

func notEmpty(s *widgets.List) bool {
	return len(s.Rows) != 0
}

type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) { return rf(p) }

func asyncCopy(ctx context.Context, dst io.Writer, src io.Reader) (err error) {
	_, err = io.Copy(dst, readerFunc(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return src.Read(p)
		}
	}))
	return
}
