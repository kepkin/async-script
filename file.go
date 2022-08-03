package async_script

import (
	"io"
	"os"
	"path/filepath"
)

type fromFile struct {
	Path string
	in   io.ReadCloser
}

func FromFile(path string) Op {
	return &fromFile{
		Path: path,
		in:   nil,
	}
}

func (p *fromFile) SetInput(in io.ReadCloser) {
	panic("Can't set Input into source")
}

func (p *fromFile) GetReader() io.ReadCloser {
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

func (p *fromFile) Run() error {
	return nil
}

type toFile struct {
	Path string
	in   io.ReadCloser
}

func ToFile(path string) Op {
	return &toFile{
		Path: path,
		in:   nil,
	}
}

func (p *toFile) SetInput(in io.ReadCloser) {
	p.in = in
}

func (p *toFile) GetReader() io.ReadCloser {
	panic("not supported")
}

func (p *toFile) Run() error {
	out, err := os.Create(p.Path)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(out, p.in)
	return err
}

type glob struct {
	Path  string
	pipeR *os.File
	pipeW *os.File
}

func Glob(path string) Op {
	return &glob{
		Path: path,
	}
}

func (p *glob) SetInput(in io.ReadCloser) {
	panic("Can't set Input into source")
}

func (p *glob) GetReader() io.ReadCloser {
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

func (p *glob) Run() error {
	dir, err := os.Open(p.Path)
	if err != nil {
		return err
	}

	files, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	for _, f := range files {
		p.pipeW.WriteString(filepath.Join(p.Path, f.Name()))
		p.pipeW.Write([]byte{'\n'})
	}

	return nil
}
