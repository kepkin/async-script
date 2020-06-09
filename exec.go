package async_script

import (
	"io"
	"os"
	"os/exec"

	"bitbucket.org/creachadair/shell"
)

type execOp struct {
	cmd *exec.Cmd
	in  io.ReadCloser
	out io.ReadCloser
}

func ExecWithCmd(cmdString string, cmd exec.Cmd) Op {
	res := &execOp{}

	if cmdString != "" {
		args, ok := shell.Split(cmdString) // strings.Fields doesn't handle quotes
		if !ok {
			panic("TODO")
		}
		res.cmd = exec.Command(args[0], args[1:]...)
	}

	res.cmd.Dir = cmd.Dir
	res.cmd.Env = cmd.Env

	return res
}

func Exec(cmd string) Op {
	args, ok := shell.Split(cmd) // strings.Fields doesn't handle quotes
	if !ok {
		panic("TODO")
	}

	res := &execOp{
		exec.Command(args[0], args[1:]...),
		nil,
		nil,
	}

	return res
}

func (p *execOp) SetInput(in io.ReadCloser) {
	p.cmd.Stdin = in
}

func (p *execOp) GetReader() io.ReadCloser {
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

func (p *execOp) Run() error {
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
