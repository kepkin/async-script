package async_script

import (
	"bitbucket.org/creachadair/shell"
	"golang.org/x/crypto/ssh/terminal"
	"bufio"
	"bytes"
	"fmt"
	"github.com/alexflint/go-arg"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

func ScanLinesWithoutDrop(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

type AsyncPipe interface {
	Run() error
	SetInput(io.ReadCloser)
	GetReader() io.ReadCloser
}

type ExecAsyncPipe struct {
	cmd *exec.Cmd
	in  io.ReadCloser
	out io.ReadCloser
}

func ExecutePipes(pipes ...AsyncPipe) {
	wg := sync.WaitGroup{}

	var err error
	for i := len(pipes) - 2; i >= 0; i-- {
		sourcePipe := pipes[i]
		targetPipe := pipes[i+1]

		targetPipe.SetInput(sourcePipe.GetReader())
		wg.Add(1)
		go func() {
			localErr := targetPipe.Run()
			if localErr != nil {
				err = localErr
			}
			wg.Done()
		}()
	}
	localErr := pipes[0].Run()
	if localErr != nil {
		err = localErr
	}
	wg.Wait()

	if err != nil {
		log.Fatal(err)
	}
}

func ExecAsync(cmd string) AsyncPipe {
	args, ok := shell.Split(cmd) // strings.Fields doesn't handle quotes
	if !ok {
		panic("TODO")
	}

	res := &ExecAsyncPipe{
		exec.Command(args[0], args[1:]...),
		nil,
		nil,
	}

	return res
}

func (p *ExecAsyncPipe) SetInput(in io.ReadCloser) {
	p.cmd.Stdin = in
}

func (p *ExecAsyncPipe) GetReader() io.ReadCloser {
	if p.out != nil {
		return p.out
	}

	var err error
	p.out, err = p.cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	return p.out
}

func (p *ExecAsyncPipe) Run() error {
	if p.cmd.Stdin == nil {
		p.cmd.Stdin = os.Stdin
	}

	if p.cmd.Stdout == nil {
		p.cmd.Stdout = os.Stderr
	}

	//@TODO: save stderr somewhere
	if p.cmd.Stderr == nil {
		p.cmd.Stderr = os.Stderr
	}

	return p.cmd.Run()
}

type WatchAsyncPipe struct {
	in    io.ReadCloser
	pipeR *os.File
	pipeW *os.File
}

func Watch() AsyncPipe {
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

type StringAsyncPipe struct {
	in io.ReadCloser
}

func StringPipe() AsyncPipe {
	return &StringAsyncPipe{
		os.Stdin,
	}
}

func (p *StringAsyncPipe) SetInput(in io.ReadCloser) {
	p.in = in
}

func (p *StringAsyncPipe) GetReader() io.ReadCloser {
	panic("Can't get reader from String")
}

func (p *StringAsyncPipe) Run() error {
	_, err := ioutil.ReadAll(p.in)
	return err
}

type FileAsyncPipe struct {
	Path string
	in   io.ReadCloser
}

func (p *FileAsyncPipe) SetInput(in io.ReadCloser) {
	panic("Can't set Input into source")
}

func (p *FileAsyncPipe) GetReader() io.ReadCloser {
	if p.in != nil {
		return p.in
	}

	var err error
	p.in, err = os.Open(p.Path)
	if err != nil {
		panic(err)
	}
	return p.in
}

func (p *FileAsyncPipe) Run() error {
	return nil
}

type FileToAsyncPipe struct {
	Path string
	in   io.ReadCloser
}

func (p *FileToAsyncPipe) SetInput(in io.ReadCloser) {
	p.in = in
}

func (p *FileToAsyncPipe) GetReader() io.ReadCloser {
	panic("not supported")
}

func (p *FileToAsyncPipe) Run() error {
	out, err := os.Create(p.Path)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(out, p.in)
	return err
}

type ReplaceAsyncPipe struct {
	What  []byte
	To    []byte
	in    io.ReadCloser
	pipeR *os.File
	pipeW *os.File
}

func Replace(what, to string) AsyncPipe {
	return &ReplaceAsyncPipe{
		[]byte(what),
		[]byte(to),
		nil,
		nil,
		nil,
	}
}

func (p *ReplaceAsyncPipe) SetInput(in io.ReadCloser) {
	p.in = in
}

func (p *ReplaceAsyncPipe) GetReader() io.ReadCloser {
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

func (p *ReplaceAsyncPipe) Run() error {
	defer p.pipeW.Close()

	scanner := bufio.NewScanner(p.in)
	scanner.Split(ScanLinesWithoutDrop)

	for scanner.Scan() {
		line := scanner.Bytes()
		replaced := bytes.Replace(line, p.What, p.To, -1)
		nw, ew := p.pipeW.Write(replaced)
		if ew != nil {
			return ew
		}
		if len(replaced) != nw {
			return io.ErrShortWrite
		}
	}

	return nil
}

