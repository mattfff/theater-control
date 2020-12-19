package cec

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-cmd/cmd"
)

type Listener struct {
	stdin   io.Writer
	stdout  io.Reader
	command *cmd.Cmd
}

const (
	TypeTV         uint = 0
	TypeRecording1 uint = 1
	TypeRecording2 uint = 2
	TypeTuner1     uint = 3
	TypePlayback1  uint = 4
	TypeAudio      uint = 5
	TypeTuner2     uint = 6
	TypeTuner3     uint = 7
	TypePlayback2  uint = 8
	TypePlayback3  uint = 9
	TypeTuner4     uint = 10
	TypePlayback4  uint = 11
)

const (
	MessagePower          uint = 144
	MessageGiveAudio      uint = 113
	MessageControlPressed uint = 68
	MessageReportAudio    uint = 122
)

const (
	ButtonUp           uint = 1
	ButtonDown         uint = 2
	ButtonLeft         uint = 3
	ButtonRight        uint = 4
	ButtonRightUp      uint = 5
	ButtonRightDown    uint = 6
	ButtonLeftUp       uint = 7
	ButtonLeftDown     uint = 8
	ButtonRootMenu     uint = 9
	ButtonSetupMenu    uint = 10
	ButtonContentsMenu uint = 11
	ButtonExit         uint = 13
	ButtonVolumeUp     uint = 65
	ButtonVolumeDown   uint = 66
	ButtonMute         uint = 67
)

type Message struct {
	Target  uint
	Source  uint
	Message uint
	Values  []uint
}

const OUTGOING = "<<"
const INCOMING = ">>"

func handleOutput(raw string) (Message, bool) {
	// example command
	// 2020/12/18 17:51:22 TRAFFIC: [         5310420]	<< 50:90:00

	var indexStart = strings.LastIndex(raw, INCOMING) + len(INCOMING) + 1
	command := strings.TrimSpace(raw[indexStart:])

	parts := strings.Split(command, ":")

	if len(parts) < 3 {
		return Message{}, false
	}

	source, _ := strconv.ParseUint(string(parts[0][0]), 16, 8)
	target, _ := strconv.ParseUint(string(parts[0][1]), 16, 8)
	message, _ := strconv.ParseUint(parts[1], 16, 8)

	values := make([]uint, len(parts)-2)

	for index, val := range parts[2:] {
		parsed, _ := strconv.ParseUint(val, 16, 8)
		values[index] = uint(parsed)
	}

	return Message{
		Target:  uint(target),
		Source:  uint(source),
		Message: uint(message),
		Values:  values,
	}, true
}

func Open(output chan Message) *Listener {
	listener := Listener{}
	listener.launch(output)

	return &listener
}

func (l *Listener) launch(output chan Message) {
	opts := cmd.Options{
		Streaming: true,
		Buffered:  false,
	}

	l.command = cmd.NewCmdOptions(opts, "cec-client", "-t", "a", "-d", "8")
	reader, writer := io.Pipe()

	defer func() {
		writer.Close()
		reader.Close()
	}()

	l.stdin = writer
	l.stdout = reader

	go func() {
		for {
			select {
			case out := <-l.command.Stdout:
				log.Printf("CEC command: %s\n", out)
				if strings.LastIndex(out, INCOMING) >= 0 {
					message, hasMessage := handleOutput(out)
					if hasMessage {
						output <- message
					}
				}
			}
		}
	}()

	statusChannel := l.command.Start()

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

func (l *Listener) Send(msg Message) {
	source := strconv.FormatInt(int64(msg.Source), 16)
	target := strconv.FormatInt(int64(msg.Target), 16)
	message := strconv.FormatInt(int64(msg.Message), 16)
	values := make([]string, len(msg.Values)+2)

	values[0] = source + target
	values[1] = message

	for index, each := range msg.Values {
		values[index+2] = strconv.FormatInt(int64(each), 16)
	}

	command := "tx " + strings.ToLower(strings.Join(values, ":"))

	fmt.Printf("CEC command sent: %s\n", command)

	fmt.Fprintf(l.stdin, command)

	io.Copy(os.Stdout, l.stdout)
}
