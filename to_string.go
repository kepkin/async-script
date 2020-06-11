package async_script

import (
	"io"
	"io/ioutil"
)

type toString struct {
	dst  *string
	in    io.ReadCloser
}

func ToString(dst *string) Op {
	return &toString{
		dst,
		nil,
	}
}

func (p *toString) SetInput(in io.ReadCloser) {
	p.in = in
}

func (p *toString) GetReader() io.ReadCloser {
	return nil
}

func (p *toString) Run() error {
	data, err := ioutil.ReadAll(p.in)
	if err != nil {
		return err
	}

	*p.dst = string(data)

	return nil
}
