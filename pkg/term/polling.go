package term

import (
	"context"
	"fmt"
	"io"
	"os"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const ctrlC = "<C-c>"
const enter = "<Enter>"

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

		case "<Down>":
			c.namespaceList.ScrollDown()

		case "<Up>":
			c.namespaceList.ScrollUp()

		case "<Home>":
			c.namespaceList.ScrollTop()

		case "<End>":
			c.namespaceList.ScrollBottom()

		case "<PageUp>":
			c.namespaceList.ScrollPageUp()

		case "<PageDown>":
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

		case "<Down>":
			c.podScroll(focus.ScrollDown)

		case "<Up>":
			c.podScroll(focus.ScrollUp)

		case "<Home>":
			c.podScroll(focus.ScrollTop)

		case "<End>":
			c.podScroll(focus.ScrollBottom)

		case "<PageUp>":
			c.podScroll(focus.ScrollPageUp)

		case "<PageDown>":
			c.podScroll(focus.ScrollPageDown)

		// bring up namespace menu
		case "n":
			c.displayNamespaceList()
			return

		// switch between panes - will add detail window too
		case "<Tab>":
			c.switchPane()

		// render console
		case "c":
			cancelIfNotNil(logCancel)
			c.navWindow.FocusRight()
			ui.Render(c.navWindow, c.console)
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
	ctx, cancel := context.WithCancel(context.Background())
	go asyncCopy(ctx, stdin, os.Stdin)
	for {
		select {
		case <-stch:
			cancel()
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

		case "<Down>":
			c.console.ScrollDown()

		case "<Up>":
			c.console.ScrollUp()

		case "<Home>":
			c.console.ScrollTop()

		case "<End>":
			c.console.ScrollBottom()

		case "<PageUp>":
			c.console.ScrollPageUp()

		case "<PageDown>":
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
		ui.Render(c.console)
	}
}
