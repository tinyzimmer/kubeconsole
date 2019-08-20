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

	quit = "QUIT"
)

var logContext context.Context
var logCancel func()
var focus *widgets.List
var logsPaused bool

func (c *controller) checkCommon(focus *widgets.List, event string) (q string) {
	switch event {
	// quit
	case "q", ctrlC:
		return quit

	case down:
		c.focusScroll(focus, down)

	case up:
		c.focusScroll(focus, up)

	case home:
		c.focusScroll(focus, home)

	case end:
		c.focusScroll(focus, end)

	case pageUp:
		c.focusScroll(focus, pageUp)

	case pageDown:
		c.focusScroll(focus, pageDown)
	}
	return ""
}

func (c *controller) pollNamespaces(ch chan string) {
	c.renderNamespaceList()
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {

		// switch to pod view
		case "p":
			c.navWindow.FocusRight()
			ui.Clear()
			c.pollPods()
			return

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
			c.currentNamespace = c.getSelectedNamespace()
			ch <- c.currentNamespace
			c.navWindow.FocusRight()
			ui.Clear()
			c.renderDefaults()
			c.debug(fmt.Sprintf("Fetching pods for %s", c.currentNamespace))
			c.pollPods()
			return

		default:
			if q := c.checkCommon(c.namespaceList, e.ID); q == quit {
				cancelIfNotNil(logCancel)
				return
			}
		}
		c.renderNamespaceList()
	}
}

func (c *controller) pollPods() {
	uiEvents := ui.PollEvents()
	focus = c.podList
	for {
		ui.Clear()
		c.renderDefaults()
		e := <-uiEvents
		switch e.ID {

		// reload
		case "r":
			ch := make(chan string)
			c.podList = c.newPodList(ch)
			ch <- c.currentNamespace

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

			// tail pod logs
		case "t":
			cancelIfNotNil(logCancel)
			logContext, logCancel = context.WithCancel(context.Background())
			if q := c.tailPod(); q == quit {
				cancelIfNotNil(logCancel)
				return
			}

		// get pod details
		case enter:
			c.debug("Loading pod...")
			c.selectPod()

		case "e":
			cancelIfNotNil(logCancel)
			stdin, stopch, q := c.RunExecutor()
			if q == quit {
				return
			}
			c.pollExecutor(stdin, stopch)
			return

		default:
			if q := c.checkCommon(focus, e.ID); q == quit {
				cancelIfNotNil(logCancel)
				return
			}

		}

	}
}

func (c *controller) pollExecutor(stdin *io.PipeWriter, stch chan struct{}) {

	// redirect all stdin to the terminal,
	ctx, cancel := context.WithCancel(context.Background())
	go asyncCopy(ctx, stdin, os.Stdin)
	//	events := ui.PollEvents()

	// wait for a stop from the exec stream
	//
	// I'd like to take more control over the stdin copy and feed back
	// page events to scroll through the command history
	for {
		select {
		case <-stch:
			cancel()
			stdin.Close()
			c.pollPods()
			return
		// case e := <-events:
		// 	switch e.ID {
		// 	case pageUp:
		// 		c.execWindow.ScrollPageUp()
		// 		ui.Render(c.execWindow)
		// 	case pageDown:
		// 		c.execWindow.ScrollPageDown()
		// 		ui.Render(c.execWindow)
		// 	}
		default:
		}
	}
}

func (c *controller) pollConsole() {
	uiEvents := ui.PollEvents()
	c.consoleFocused = true

	for {
		e := <-uiEvents
		switch e.ID {

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

		default:
			if q := c.checkCommon(c.console, e.ID); q == quit {
				return
			}
		}
		c.renderConsole()

	}
}
