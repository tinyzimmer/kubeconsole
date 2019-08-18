package term

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"runtime/debug"
	"strings"

	ui "github.com/gizak/termui/v3"
)

func (c *controller) podScroll(f func()) {
	c.setupIfLogWindow()
	if len(focus.Rows) > 0 {
		f()
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

func newErrorWithStack(err error) (serr error) {
	serr = errors.New(fmt.Sprintf("ERROR: %v\n%s", err, string(debug.Stack())))
	return serr
}

func (c *controller) getPodDetails() {
	details, err := c.factory.GetPod(currentNamespace, currentPod)
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
	c.resetLogWindow()
	logsPaused = false
	if len(c.podList.Rows) > 0 {
		currentPod = c.podList.Rows[c.podList.SelectedRow]
		c.debug(fmt.Sprintf("Fetching details for %s", currentPod))
		c.detailsChan <- fmt.Sprintf("Loading details for %s...\n", currentPod)
		c.logChan <- clearEvent
		c.logChan <- fmt.Sprintf("Fetching logs for %s...\n", currentPod)
		if stream, err := c.factory.GetLogStream(currentNamespace, currentPod, logContext); err != nil {
			c.errorChan <- newErrorWithStack(err)
		} else {
			c.debug(fmt.Sprintf("Starting log stream for %s", currentPod))
			go c.streamLogsToWindow(logContext, stream)
		}
		go c.getPodDetails()
	}
}

func (c *controller) switchPane() {
	focus.Title = strings.Replace(focus.Title, " * ", "", 1)
	if focus == c.podList {
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
