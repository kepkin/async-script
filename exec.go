package async_script

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"bitbucket.org/creachadair/shell"
)

type execOp struct {
	cmd           *exec.Cmd
	stderrToStdin bool
	in            io.ReadCloser
	out           io.ReadCloser
}

func ExecWithCmd(cmdString string, cmd exec.Cmd, stderrToStdin bool) Op {
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
	res.stderrToStdin = stderrToStdin

	return res
}

func Execf(cmd string, a ...interface{}) Op {
	return Exec(fmt.Sprintf(cmd, a...))
}

// Runs shell command
func Exec(cmd string) Op {
	args, ok := shell.Split(cmd) // strings.Fields doesn't handle quotes
	if !ok {
		panic("TODO")
	}

	res := &execOp{
		exec.Command(args[0], args[1:]...),
		false,
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

	if p.stderrToStdin {
		p.cmd.Stderr = p.cmd.Stdout
	}

	return p.cmd.Run()
}
