package async_script

import (
	"bufio"
	"io"
	"os"
)

type TransformerAsyncPipe struct {
	splitFunc     bufio.SplitFunc
	transformFunc func([]byte) []byte

	in    io.ReadCloser
	pipeR *os.File
	pipeW *os.File
}

func Transformer(splitFunc bufio.SplitFunc, transformFunc func([]byte) []byte) Op {
	return &TransformerAsyncPipe{
		splitFunc,
		transformFunc,
		nil,
		nil,
		nil,
	}
}

func (p *TransformerAsyncPipe) SetInput(in io.ReadCloser) {
	p.in = in
}

func (p *TransformerAsyncPipe) GetReader() io.ReadCloser {
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

func (p *TransformerAsyncPipe) Run() error {
	defer p.pipeW.Close()

	scanner := bufio.NewScanner(p.in)
	scanner.Split(p.splitFunc)

	for scanner.Scan() {
		part := scanner.Bytes()
		outPart := p.transformFunc(part)

		nw, ew := p.pipeW.Write(outPart)
		if ew != nil {
			return ew
		}
		if len(outPart) != nw {
			return io.ErrShortWrite
		}
	}

	return nil
}
