package cec

import (
	"io"
	"log"
	"os"

	"github.com/go-cmd/cmd"
)

type Listener struct {
	stdin   io.Writer
	command *cmd.Cmd
}

func processOutput(str string) string {
	log.Printf("Outline: %v\n", str)
	return str
}

func Open(output chan string) *Listener {
	listener := Listener{}
	listener.launch(output)

	return &listener
}

func (l *Listener) launch(output chan string) {
	opts := cmd.Options{
		Streaming: true,
		Buffered:  false,
	}

	l.command = cmd.NewCmdOptions(opts, "cec-client", "-t", "a", "-d", "8")
	reader, writer, err := os.Pipe()

	if err != nil {
		log.Fatalf("Failed to open pipe %v", err)
		return
	}

	defer func() {
		writer.Close()
		reader.Close()
	}()

	l.stdin = writer

	go func() {
		for {
			select {
			case out := <-l.command.Stdout:
				output <- processOutput(out)
			}
		}
	}()

	statusChannel := l.command.StartWithStdin(reader)

	select {
	case <-statusChannel:
		return
	case err := <-l.command.Stderr:
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
