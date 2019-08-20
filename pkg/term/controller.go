package term

import (
	"fmt"
	"os"
	"sync"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/tinyzimmer/kubeconsole/pkg/k8sutils"
	"golang.org/x/crypto/ssh/terminal"
)

// Controller is the exported controller interface
type Controller interface {
	Run() error
}

type controller struct {
	Controller

	factory k8sutils.KubernetesFactory

	navWindow     *widgets.TabPane
	helpWindow    *widgets.Paragraph
	serverWindow  *widgets.Paragraph
	namespaceList *widgets.List
	podList       *widgets.List
	detailsWindow *widgets.List
	logWindow     *widgets.List
	console       *widgets.List
	execWindow    *widgets.List
	errorWindow   *widgets.Paragraph

	currentNamespace string

	consoleFocused bool
	debugToFile    bool

	logChan     chan string
	detailsChan chan string
	debugChan   chan string
	errorChan   chan *errorWithStack

	mux       sync.Mutex
	resizemux sync.Mutex
}

// New returns a new terminal ui controller
func New(factory k8sutils.KubernetesFactory, debug bool) Controller {
	c := &controller{factory: factory}
	c.debugToFile = debug
	c.errorChan = make(chan *errorWithStack)
	c.debugChan = make(chan string)
	c.navWindow = newNavWindow(c.debugToFile)
	c.serverWindow = c.newAPIServerWindow()
	c.helpWindow = newHelpWindow()
	c.detailsWindow, c.detailsChan = newDetailsWindow()
	c.logWindow, c.logChan = c.newLogWindow()
	c.console = newConsoleWindow()
	c.execWindow = newExecWindow()
	c.errorWindow = newErrorWindow()
	return c
}

func (c *controller) Run() error {
	c.debug(fmt.Sprintf("Connected to %s", c.factory.APIHost()))
	c.debug("Starting handlers")
	// listen on the error channel
	go c.listenForErrors()
	//  handle terminal resizes - work in progress
	go c.handleResize()

	c.debug("Fetching namespaces")
	// set up the namespace list
	selectionChan := make(chan string)
	c.namespaceList = c.newNamespaceList()

	// prepare the pod list
	c.podList = c.newPodList(selectionChan)

	// render defaults and bring up namespace prompt
	c.debug("Starting poll loops")
	c.renderDefaults()
	c.pollNamespaces(selectionChan)
	return nil
}

// renderDefaults renders the default panes
func (c *controller) renderDefaults() {
	c.mux.Lock()
	defer c.mux.Unlock()
	ui.Render(
		c.navWindow,
		c.serverWindow,
		c.helpWindow,
		c.podList,
		c.detailsWindow,
		c.logWindow,
	)
}

// resizeDefaults is my hacky resizer for now. It essentially starts
// a brand new session, saving current values where appropriate.
// This current causes some bugginess after the resize. You basically need
// to switch around the namespace and pod view to get things back to perfect.
// But at least you don't have to restart.
func (c *controller) resizeDefaults() {
	c.resizemux.Lock()
	defer c.resizemux.Unlock()
	c.navWindow = newNavWindow(c.debugToFile)
	c.serverWindow = c.newAPIServerWindow()
	c.helpWindow = newHelpWindow()

	ch := make(chan string)
	c.podList = c.newPodList(ch)
	if c.currentNamespace != "" {
		ch <- c.currentNamespace
	} else {
		c.namespaceList = c.newNamespaceList()
	}

	consoleBak := c.console.Rows
	c.console = newConsoleWindow()
	c.console.Rows = consoleBak

	detailsBak := c.detailsWindow.Rows
	c.detailsWindow, c.detailsChan = newDetailsWindow()
	c.detailsWindow.Rows = detailsBak

	execBak := c.execWindow.Rows
	c.execWindow = newExecWindow()
	c.execWindow.Rows = execBak

	logBak := c.logWindow.Rows
	c.logWindow, c.logChan = c.newLogWindow()
	c.logWindow.Rows = logBak

	ui.Clear()
	c.renderDefaults()
	if c.currentNamespace == "" {
		c.renderNamespaceList()
	}
}

// just render the namespace prompt
func (c *controller) renderNamespaceList() {
	c.debug("Rendering namespace list")
	c.mux.Lock()
	defer c.mux.Unlock()
	ui.Render(c.namespaceList)
}

// render the debug console
func (c *controller) renderConsole() {
	c.mux.Lock()
	defer c.mux.Unlock()
	ui.Render(c.console)
}

// reset the Log Window
func (c *controller) resetLogWindow() {
	c.debug("Resetting log window")
	c.logWindow, c.logChan = c.newLogWindow()
}

// listen on the error channel and bring up a prompt when
// any get raised
func (c *controller) listenForErrors() {
	c.debug("Starting error listener")
	for {
		select {
		case err := <-c.errorChan:
			c.mux.Lock()
			c.debug(err.Error())
			c.debug(err.Stack())
			c.errorWindow.Text = err.Error()
			ui.Render(c.errorWindow)
			time.Sleep(time.Duration(2) * time.Second)
			c.mux.Unlock()
		}
	}
}

// write debug message to console
func (c *controller) debug(msg string) {
	newMsg := fmt.Sprintf("> [%v] %s", time.Now().Local(), msg)
	c.console.Rows = append(c.console.Rows, newMsg)
	c.console.ScrollBottom()
	if c.consoleFocused {
		c.renderConsole()
	}
	if c.debugToFile {
		c.appendDebugFile(newMsg)
	}
}

func (c *controller) appendDebugFile(msg string) {
	f, err := os.OpenFile("debug.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		c.errorChan <- newErrorWithStack(err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("%s\n", msg)); err != nil {
		c.errorChan <- newErrorWithStack(err)
		return
	}
}

// check for a change in terminal size (run in a goroutine)
// when terminal size change +/- 1 pixel - resise the windows
func (c *controller) handleResize() {
	var x, y, nx, ny int
	var err error
	x, y, err = terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}
	for {
		nx, ny, err = terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			panic(err)
		}
		if !plusOrMinus(x, nx, 2) || !plusOrMinus(y, ny, 2) {
			c.debug("Detected terminal resize")
			c.resizeDefaults()
		}
		x = nx
		y = ny
	}
}

func plusOrMinus(val1 int, val2 int, comp int) bool {
	for x := 0; x <= comp; x++ {
		switch {
		case val1 == val2-x || val1 == val2+x:
			return true
		}
	}
	return false
}
