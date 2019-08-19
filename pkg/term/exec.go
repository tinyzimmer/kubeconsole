package term

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"k8s.io/client-go/tools/remotecommand"
)

func (c *controller) RunExecutor() (stdinWriter *io.PipeWriter, stopch chan struct{}) {
	currentPod = c.podList.Rows[c.podList.SelectedRow]
	exec, err := c.factory.GetExecutor(currentNamespace, currentPod)
	// var (
	// 	stdin  bytes.Buffer
	// 	stdout bytes.Buffer
	// )

	if err != nil {
		c.errorChan <- newErrorWithStack(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	stopch = make(chan struct{})

	// Create pipes for stdin/stdout and a buffer for transferring stdout
	// to the terminal window
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	var buf bytes.Buffer
	go asyncCopy(ctx, &buf, stdoutReader)

	// stream stdout to screen
	go func() {
		for {
			select {
			case <-ctx.Done():
				// stop streaming and signal the poller to stop blocking
				c.debug("Stopping exec window stream")
				stopch <- struct{}{}
				return
			default:

				// der be demons that need cleaning here - especiallfor special events
				out, _ := decode(buf.Bytes())
				stripped := stripAnsi(string(out))
				split := strings.Split(strings.Replace(stripped, "\r\n", "\n", -1), "\n")

				if !cmp.Equal(c.execWindow.Rows, split) {
					c.execWindow.Rows = split
					c.execWindow.ScrollBottom()
					c.mux.Lock()
					ui.Render(c.execWindow)
					c.mux.Unlock()
				}

			}
		}
	}()

	// Start the actual command
	opts := remotecommand.StreamOptions{
		Stdin:  stdinReader,
		Stdout: stdoutWriter,
		Stderr: stdoutWriter,
		Tty:    true,
	}

	go func() {
		defer cancel()
		c.debug(fmt.Sprintf("Starting exec stream for %s", currentPod))
		err = exec.Stream(opts)
		if err != nil {
			c.errorChan <- newErrorWithStack(err)
			time.Sleep(time.Duration(1) * time.Second)
		}
		c.debug(fmt.Sprintf("Finished exec stream for %s", currentPod))
	}()

	return
}

const ansire = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansire)

func stripAnsi(str string) string {
	return re.ReplaceAllString(str, "")
}

func decode(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, unicode.UTF8.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}
