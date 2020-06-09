package async_script

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

type WatchAsyncPipe struct {
	in    io.ReadCloser
	pipeR *os.File
	pipeW *os.File
}

func Watch() Op {
	return &WatchAsyncPipe{
		os.Stdin,
		nil,
		nil,
	}
}

func (p *WatchAsyncPipe) SetInput(in io.ReadCloser) {
	p.in = in
}

func (p *WatchAsyncPipe) GetReader() io.ReadCloser {
	if p.pipeR != nil {
		return p.pipeR
	}

	var err error
	p.pipeR, p.pipeW, err = os.Pipe()
	if err != nil {
		panic(err)
	}

	return p.pipeR
}

func (p *WatchAsyncPipe) Run() error {
	//TODO: check err
	lastLine := ""

	var err error
	c := make(chan struct{})
	go func() {
		if p.pipeW != nil {
			defer p.pipeW.Close()
		}

		scanner := bufio.NewScanner(p.in)
		for scanner.Scan() {
			line := scanner.Bytes()

			if len(line) > 1 {
				lastLine = string(line)
			}

			if p.pipeW == nil {
				continue
			}

			nw, ew := p.pipeW.Write(line)
			if ew != nil {
				err = ew
				break
			}
			if len(line) != nw {
				err = io.ErrShortWrite
				break
			}
		}
		close(c)
	}()

	var terminalWidth int
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		terminalWidth, _, _ = terminal.GetSize(int(os.Stdout.Fd()))
	}
	clearString := ""
	for ; terminalWidth > 0; terminalWidth-- {
		clearString += " "
	}

	for {
		select {
		case <-c:
			fmt.Println()
			return err

		case <-time.After(time.Second):
			fmt.Print("\r", clearString)
			fmt.Print("\r", lastLine)

		}
	}
}
