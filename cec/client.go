package cec

import (
	"fmt"
	"io"
	"os"

	"github.com/go-cmd/cmd"
)

type Listener struct {
	Stdout chan string
	Stderr chan string
	Stdin  io.Reader
}

func processOutput(str string) string {
	fmt.Printf("Outline: %v", str)
	return str
}

func Open() *Listener {
	listener := Listener{}
	listener.Launch()

	return &listener
}

func (l *Listener) Launch() {
	cecCommand := cmd.NewCmd("cec-client")
	stdout := make(chan string, 100)
	stderr := make(chan string, 100)

	_, stdin, _ := os.Pipe()

	cecCommand.StartWithStdin(stdin)

	l.Stdout = stdout
	l.Stderr = stderr
	l.Stdin = stdin

	go func() {
		for {
			select {
			case output := <-stdout:
				processOutput(output)
				processOutput(output)
			}
		}
	}()

	cecCommand.Stdout = stdout
	cecCommand.Stderr = stderr

	statusChannel := cecCommand.Start()

	select {
	case <-statusChannel:
		return
	default:
	}
}
