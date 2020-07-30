package async_script

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

const ttyClearAfterCursor = "\u001b[0J"
const ttyCursorUp = "\u001b[%vA"

type WatchAsyncPipe struct {
	in     io.ReadCloser
	pipeR  *os.File
	pipeW  *os.File
	buf    []string
	bufIdx int
}

func Watch(lines int) Op {
	return &WatchAsyncPipe{
		os.Stdin,
		nil,
		nil,
		make([]string, lines, lines),
		0,
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

func (p *WatchAsyncPipe) addBufLine(line string) {
	p.buf[p.bufIdx] = line

	p.bufIdx += 1
	if p.bufIdx >= len(p.buf) {
		p.bufIdx = 0
	}
}

func (p *WatchAsyncPipe) printBufLine() {
	fmt.Print(ttyClearAfterCursor)
	//linesUp := 0
	//terminalWidth, _, _ := terminal.GetSize(int(os.Stdout.Fd()))

	for i := 0; i < len(p.buf); i++ {
		idx := i + p.bufIdx
		if idx >= len(p.buf) {
			idx -= len(p.buf)
		}

		//if len(p.buf[idx]) > terminalWidth {
		//	linesUp += len(p.buf[idx]) / terminalWidth
		//}
		fmt.Println(p.buf[idx])
	}

	fmt.Printf(ttyCursorUp, len(p.buf))
}

func (p *WatchAsyncPipe) Run() error {
	if !terminal.IsTerminal(int(os.Stdout.Fd())) && p.pipeW == nil {
		p.pipeW = os.Stdout
	}

	var err error
	c := make(chan string)
	go func() {
		if p.pipeW != nil && p.pipeW != os.Stdout {
			defer p.pipeW.Close()
		}

		scanner := bufio.NewScanner(p.in)
		scanner.Split(ScanLinesWithoutDrop)

		for scanner.Scan() {
			line := scanner.Bytes()
			c <- strings.TrimRight(string(line), " \t\r\n")

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

	for d := range c {
		if !terminal.IsTerminal(int(os.Stdout.Fd())) {
			continue
		} else {
			p.addBufLine(d)
			p.printBufLine()
		}
	}

	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Print(ttyClearAfterCursor)
	}

	return err
}
