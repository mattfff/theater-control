package cec

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/go-cmd/cmd"
)

type Listener struct {
	stdout  chan string
	stderr  chan string
	stdin   io.Writer
	command *cmd.Cmd
}

func processOutput(str string) string {
	fmt.Printf("Outline: %v", str)
	return str
}

func Open(output chan string) *Listener {
	listener := Listener{}
	listener.launch(output)

	return &listener
}

func (l *Listener) launch(output chan string) {
	l.command = cmd.NewCmd("cec-client -t a -d 8")
	reader, writer, err := os.Pipe()

	if err != nil {
		log.Fatalf("Failed to open pipe %v", err)
		return
	}

	defer func() {
		writer.Close()
		reader.Close()
	}()

	stderr := make(chan string, 100)

	l.command.StartWithStdin(reader)

	l.stdout = output
	l.stderr = stderr
	l.stdin = writer

	go func() {
		for {
			select {
			case out := <-output:
				processOutput(out)
			}
		}
	}()

	l.command.Stdout = output
	l.command.Stderr = stderr

	statusChannel := l.command.Start()

	select {
	case <-statusChannel:
		return
	case err := <-stderr:
		log.Printf("Error: %v\n", err)
	default:
	}
}

func (l *Listener) Close() {
	l.command.Stop()
}

func (l *Listener) Send(command []byte) {
	l.stdin.Write(command)
}
