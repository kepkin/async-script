package async_script

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

type ReplaceAsyncPipe struct {
	What  []byte
	To    []byte
	in    io.ReadCloser
	pipeR *os.File
	pipeW *os.File
}

func Replace(what, to string) Op {
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
