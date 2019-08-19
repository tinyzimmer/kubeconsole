package term

import (
	"context"
	"fmt"
	"io"
	"os"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const (
	ctrlC    = "<C-c>"
	enter    = "<Enter>"
	up       = "<Up>"
	down     = "<Down>"
	left     = "<Left>"
	right    = "<Right>"
	pageUp   = "<PageUp>"
	pageDown = "<PageDown>"
	home     = "<Home>"
	end      = "<End>"
)

var currentNamespace string
var currentPod string
var logContext context.Context
var logCancel func()
var focus *widgets.List
var logsPaused bool

func (c *controller) pollNamespaces(ch chan string) {
	c.debug("Polling namespaces...")
	c.renderNamespaceList()
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {

		// quit
		case "q", ctrlC:
			return

		// switch to pod view
		case "p":
			c.navWindow.FocusRight()
			ui.Clear()
			c.pollPods()
			return

		case down:
			c.namespaceList.ScrollDown()

		case up:
			c.namespaceList.ScrollUp()

		case home:
			c.namespaceList.ScrollTop()

		case end:
			c.namespaceList.ScrollBottom()

		case pageUp:
			c.namespaceList.ScrollPageUp()

		case pageDown:
			c.namespaceList.ScrollPageDown()

		// reload
		case "r":
			c.namespaceList = c.newNamespaceList()

		// render console
		case "c":
			c.navWindow.FocusRight()
			c.navWindow.FocusRight()
			ui.Render(c.navWindow, c.console)
			c.pollConsole()
			return

		// load pods for selcted namespace
		case enter:
			currentNamespace = c.namespaceList.Rows[c.namespaceList.SelectedRow]
			ch <- currentNamespace
			c.navWindow.FocusRight()
			ui.Clear()
			c.renderDefaults()
			c.debug(fmt.Sprintf("Fetching pods for %s", currentNamespace))
			c.pollPods()
			return
		}
		c.renderNamespaceList()
	}
}

func (c *controller) pollPods() {
	c.debug("Polling pods...")
	c.renderDefaults()
	uiEvents := ui.PollEvents()
	focus = c.podList
	for {
		e := <-uiEvents
		switch e.ID {

		// reload
		case "r":
			ch := make(chan string)
			c.podList = c.newPodList(ch)
			ch <- currentNamespace

			// quit
		case "q", ctrlC:
			cancelIfNotNil(logCancel)
			return

		case down:
			c.podScroll(focus.ScrollDown)

		case up:
			c.podScroll(focus.ScrollUp)

		case home:
			c.podScroll(focus.ScrollTop)

		case end:
			c.podScroll(focus.ScrollBottom)

		case pageUp:
			c.podScroll(focus.ScrollPageUp)

		case pageDown:
			c.podScroll(focus.ScrollPageDown)

		// bring up namespace menu
		case "n":
			cancelIfNotNil(logCancel)
			c.displayNamespaceList()
			return

		// switch between panes - will add detail window too
		case "<Tab>":
			c.switchPane()

		// render console
		case "c":
			cancelIfNotNil(logCancel)
			c.navWindow.FocusRight()

			c.mux.Lock()
			ui.Render(c.navWindow, c.console)
			c.mux.Unlock()

			c.pollConsole()
			return

		// get pod details and stream logs
		case enter:
			cancelIfNotNil(logCancel)
			logContext, logCancel = context.WithCancel(context.Background())
			c.debug("Loading pod...")
			c.selectPod()

		case "e":
			cancelIfNotNil(logCancel)
			stdin, stopch := c.RunExecutor()
			c.pollExecutor(stdin, stopch)
			return

		}
		c.renderDefaults()
	}
}

func (c *controller) pollExecutor(stdin *io.PipeWriter, stch chan struct{}) {
	c.debug("Polling executor...")

	// redirect all stdin to the terminal,
	ctx, cancel := context.WithCancel(context.Background())
	go asyncCopy(ctx, stdin, os.Stdin)

	// wait for a stop from the exec stream
	for {
		select {
		case <-stch:
			cancel()
			stdin.Close()
			c.pollPods()
			return
		default:
		}
	}
}

func toByte(s string) []byte { return []byte(s) }

func (c *controller) pollConsole() {
	c.debug("Polling console...")
	uiEvents := ui.PollEvents()
	c.consoleFocused = true

	for {
		e := <-uiEvents
		switch e.ID {

		// quit
		case "q", "<C-c>":
			return

		case down:
			c.console.ScrollDown()

		case up:
			c.console.ScrollUp()

		case home:
			c.console.ScrollTop()

		case end:
			c.console.ScrollBottom()

		case pageUp:
			c.console.ScrollPageUp()

		case pageDown:
			c.console.ScrollPageDown()

		case "n":
			c.consoleFocused = false
			c.navWindow.FocusLeft()
			c.navWindow.FocusLeft()
			ui.Render(c.navWindow)
			c.displayNamespaceList()
			return

		case "p":
			c.consoleFocused = false
			c.navWindow.FocusLeft()
			ui.Clear()
			c.pollPods()
			return

		}
		c.renderConsole()

	}
}
