package async_script

import (
	"bufio"
	"io"
	"os"
)

type tee struct {
	path  string
	in    io.ReadCloser
	pipeR *os.File
	pipeW *os.File
}

func Tee(path string) Op {
	return &tee{
		path,
		nil,
		nil,
		nil,
	}
}

func (p *tee) SetInput(in io.ReadCloser) {
	p.in = in
}

func (p *tee) GetReader() io.ReadCloser {
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

func (p *tee) Run() error {
	if p.pipeW != nil {
		defer p.pipeW.Close()
	} else {
		p.pipeW = os.Stdout
	}

	teeFile, err := os.OpenFile(p.path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer teeFile.Close()

	scanner := bufio.NewScanner(p.in)
	scanner.Split(ScanLinesWithoutDrop)

	for scanner.Scan() {
		line := scanner.Bytes()
		teeFile.Write(line)
		nw, ew := p.pipeW.Write(line)
		if ew != nil {
			return ew
		}
		if len(line) != nw {
			return io.ErrShortWrite
		}
	}

	return nil
}
