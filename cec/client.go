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
	MessagePower          uint = 90
	MessageGiveAudio      uint = 71
	MessageControlPressed uint = 44
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
	ButtonVolumeUp     uint = 41
	ButtonVolumeDown   uint = 42
	ButtonMute         uint = 43
)

type Message struct {
	Target  uint
	Source  uint
	Message uint
	Values  []uint
}

const INCOMING = "<<"
const OUTGOING = ">>"

func handleOutput(raw string) Message {
	// example command
	// 2020/12/18 17:51:22 TRAFFIC: [         5310420]	<< 50:90:00

	log.Printf("CEC command received: %s\n", raw)

	var indexStart = strings.LastIndex(raw, INCOMING) + 1
	command := strings.TrimSpace(raw[indexStart:])

	parts := strings.Split(command, ":")

	source := uint(parts[0][0])
	target := uint(parts[0][1])
	message, _ := strconv.ParseUint(parts[0], 16, 8)

	values := make([]uint, len(parts)-2)

	for index, val := range parts[2:] {
		parsed, _ := strconv.ParseUint(val, 16, 8)
		values[index] = uint(parsed)
	}

	return Message{
		Target:  target,
		Source:  source,
		Message: uint(message),
		Values:  values,
	}
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
				if strings.LastIndex(out, INCOMING) >= 0 {
					output <- handleOutput(out)
				}
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

	command := strings.ToUpper(strings.Join(values, ":"))

	fmt.Printf("CEC command sent: %s\n", command)

	l.stdin.Write([]byte(command))
}
