// Package implements shell like DSL for executing commands and making
// transformations. Main motivation was to write CI/CD scripts in Go instead of
// Makefile/bash etc.
//
package async_script

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
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

// Run a chain of operations similar to shell:
// $> op1 | op2 | op3
// Returns errors if one of the operation fails
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

// The same as Run, but if one of operations fails will call
// log.Fatal(err) and stops execution
func MustRun(pipes ...Op) {
	err := Run(pipes...)
	if err != nil {
		log.Fatal(err)
	}
}

type stringOp struct {
	in io.ReadCloser
}

// TODO
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
	f     func(string) []string
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

// TODO: reuse transformer code
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
