package async_script

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

type Op interface {
	Run() error
	SetInput(io.ReadCloser)
	GetReader() io.ReadCloser
}

type Ops []Op

func (r *Ops) Add(ops ...Op) Ops {
	return append(*r, ops...)
}

func Run(pipes ...Op) error {
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

	return err
}

type stringOp struct {
	in io.ReadCloser
}

func String() Op {
	return &stringOp{
		os.Stdin,
	}
}

func (p *stringOp) SetInput(in io.ReadCloser) {
	p.in = in
}

func (p *stringOp) GetReader() io.ReadCloser {
	panic("Can't get reader from String")
}

func (p *stringOp) Run() error {
	_, err := ioutil.ReadAll(p.in)
	return err
}

type mapOp struct {
	f func(string) []string
	in    io.ReadCloser
	pipeR *os.File
	pipeW *os.File
}

func Map(f func(string) []string) Op {
	return &mapOp{
		f: f,
	}
}

func (p *mapOp) SetInput(in io.ReadCloser) {
	p.in = in
}

func (p *mapOp) GetReader() io.ReadCloser {
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

func (p *mapOp) Run() error {
	defer p.pipeW.Close()

	scanner := bufio.NewScanner(p.in)
	for scanner.Scan() {
		line := scanner.Text()
		outLines := p.f(line)
		for _, outLine := range outLines {
			nw, ew := p.pipeW.WriteString(outLine)
			if ew != nil {
				return ew
			}
			if len(outLine) != nw {
				return io.ErrShortWrite
			}
		}
	}

	return nil
}