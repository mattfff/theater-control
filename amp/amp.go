package amp

import (
	"log"
	"sync"

	"github.com/jacobsa/go-serial/serial"

	"io"
)

// Command denotes byte codes that can be sent to change state in the 7100
type Command byte

const (
	CommandPowerOff                Command = 34
	CommandPowerOn                 Command = 35
	CommandPowerToggle             Command = 6
	CommandVolumeUp                Command = 53
	CommandVolumeDown              Command = 54
	CommandSetVolume               Command = 180
	CommandMuteToggle              Command = 7
	CommandMuteOn                  Command = 32
	CommandMuteOff                 Command = 33
	CommandSourcePlus              Command = 8
	CommandSourceMinus             Command = 9
	CommandSourceDvd               Command = 13
	CommandSourceSat               Command = 14
	CommandSourceVcr               Command = 15
	CommandSourceIn4               Command = 16
	CommandSourceIn5               Command = 17
	CommandSourceIn6               Command = 18
	CommandSourceIn7               Command = 19
	CommandSourceIn8               Command = 20
	CommandSourceIn9               Command = 72
	CommandSourceIn10              Command = 73
	CommandSource71                Command = 100
	CommandSource71Toggle          Command = 136
	CommandSourceSearch            Command = 144
	CommandSurroundMinus           Command = 10
	CommandSurroundPlus            Command = 11
	CommandSurroundMono            Command = 42
	CommandSurroundStereo          Command = 43
	CommandSurroundProLogic        Command = 44
	CommandSurroundNatural         Command = 45
	CommandSurroundParty           Command = 46
	CommandSurroundClub            Command = 47
	CommandSurroundConcert         Command = 48
	CommandSurroundProLogicII      Command = 160
	CommandSurroundProLogicIIMusic Command = 161
	CommandSurroundDtsNeo6Cinema   Command = 162
	CommandSurroundDtsMatrix       Command = 163
	CommandSurroundDirect          Command = 164
	CommandSurroundDtsNeo6Music    Command = 167
	CommandSurroundDolbyEx         Command = 168
	CommandSurroundStero96         Command = 169
	CommandLateNightOn             Command = 37
	CommandLateNightOff            Command = 38
	CommandLateNightToggle         Command = 12
	CommandCinemaEqToggle          Command = 99
	CommandThxToggle               Command = 27
	CommandBassPlus                Command = 68
	CommandBassMinus               Command = 69
	CommandTreblePlus              Command = 70
	CommandTrebleMinus             Command = 71
	CommandSubwooferPlus           Command = 97
	CommandSubwooferMinus          Command = 98
	CommandSurroundVolPlus         Command = 131
	CommandSurroundVolMinus        Command = 132
	CommandCenterPlus              Command = 129
	CommandCenterMinus             Command = 130
	CommandPreset1                 Command = 124
	CommandPreset2                 Command = 125
	CommandPreset3                 Command = 126
	CommandPreset4                 Command = 127
	CommandEnhancedBassOn          Command = 156
	CommandEnhancedBassOff         Command = 157
	CommandEnhancedBassToggle      Command = 133
	CommandSetupMenuToggle         Command = 103
	CommandCursorUp                Command = 104
	CommandCursorDown              Command = 105
	CommandCursorLeft              Command = 106
	CommandCursorRight             Command = 107
	CommandCursorSelect            Command = 108
	CommandSetupMenuExit           Command = 109
	CommandCursorStepDown          Command = 114
	CommandTestNoise               Command = 23
	CommandPanelLockToggle         Command = 145
	CommandShowStatus              Command = 122
	CommandPingHello               Command = 80
	CommandGetStatus               Command = 227
)

type StatusFlag byte

const (
	StatusInputType     StatusFlag = 215
	StatusProLogic      StatusFlag = 216
	StatusFeedbackStart StatusFlag = 223
	StatusHeadphones    StatusFlag = 224
	StatusVolume        StatusFlag = 225
	StatusMute          StatusFlag = 226
	StatusSource        StatusFlag = 227
	StatusVideo         StatusFlag = 228
	StatusPower         StatusFlag = 229
	StatusZoneSource    StatusFlag = 230
	StatusZoneVideo     StatusFlag = 231
	StatusZoneVolume    StatusFlag = 232
	StatusZoneMute      StatusFlag = 233
	StatusDimmer        StatusFlag = 234
	StatusSurroundMode  StatusFlag = 236
	StatusAudioType     StatusFlag = 237
	StatusInputMethod   StatusFlag = 238
	StatusLateNight     StatusFlag = 239
	StatusCinemaEQ      StatusFlag = 240
	StatusTrebleTrim    StatusFlag = 242
	StatusBassTrim      StatusFlag = 243
	StatusCenterTrim    StatusFlag = 244
	StatusSurroundTrim  StatusFlag = 245
	StatusSubTrim       StatusFlag = 246
	StatusTriggerOne    StatusFlag = 247
	StatusTriggerTwo    StatusFlag = 248
	StatusVideoFormat   StatusFlag = 249
	StatusThx           StatusFlag = 250
	StatusSeparator     StatusFlag = 255
)

var StatusLabel = map[StatusFlag]string{
	StatusInputType:    "InputType",
	StatusProLogic:     "ProLogic",
	StatusHeadphones:   "Headphones",
	StatusVolume:       "Volume",
	StatusMute:         "Mute",
	StatusSource:       "Source",
	StatusVideo:        "Video",
	StatusPower:        "Power",
	StatusZoneSource:   "ZoneSource",
	StatusZoneVideo:    "ZoneVideo",
	StatusZoneVolume:   "ZoneVolume",
	StatusZoneMute:     "ZoneMute",
	StatusDimmer:       "Dimmer",
	StatusSurroundMode: "SurroundMode",
	StatusAudioType:    "AudioType",
	StatusInputMethod:  "InputMethod",
	StatusLateNight:    "LateNight",
	StatusCinemaEQ:     "CinemaEQ",
	StatusTrebleTrim:   "TrebleTrim",
	StatusBassTrim:     "BassTrim",
	StatusCenterTrim:   "CenterTrim",
	StatusSurroundTrim: "SurroundTrim",
	StatusSubTrim:      "SubTrim",
	StatusTriggerOne:   "TriggerOne",
	StatusTriggerTwo:   "TriggerTwo",
	StatusVideoFormat:  "VideoFormat",
	StatusThx:          "THX",
}

var StatusFlags []StatusFlag

func init() {
	StatusFlags = make([]StatusFlag, 0, len(StatusLabel))

	for k := range StatusLabel {
		StatusFlags = append(StatusFlags, k)
	}
}

type StatusMap map[StatusFlag]byte

// Amp represents operations on an amplifier.
type Amp struct {
	port io.ReadWriteCloser
	open bool
	mu   sync.Mutex
}

// Open opens a connection to the amp.
func Open(portName string) (*Amp, error) {
	options := serial.OpenOptions{
		PortName:              portName,
		BaudRate:              9600,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       16,
		InterCharacterTimeout: 500,
	}

	serialPort, err := serial.Open(options)

	if err != nil {
		log.Fatalf("serial.Open %f", err)
		return nil, err
	}

	return &Amp{port: serialPort, open: true}, nil
}

// Close port.
func (a *Amp) Close() {
	a.open = false
	a.port.Close()
}

type pollerMode uint

// Poll 7100 for data.
func (a *Amp) Poll(c chan<- StatusMap) {
	var currentFlag StatusFlag = 0

	for a.open {
		result := make([]byte, 16)

		_, err := a.port.Read(result)

		if err != nil && err != io.EOF {
			log.Fatalf("read command result %f", err)
		}

		if err == nil || err != io.EOF {
			status := make(StatusMap)
			for _, val := range result {
				switch res := StatusFlag(val); res {
				case StatusFeedbackStart:
					continue
				case StatusSeparator:
					currentFlag = 0
				default:
					if currentFlag == 0 {
						currentFlag = res
					} else {
						status[currentFlag] = byte(res)
					}
				}
			}

			statusReturn := make(StatusMap)
			for key, val := range status {
				statusReturn[key] = val
			}

			c <- statusReturn
		}
	}
}

// SendCommand sends a command to the amp and reads back the resulting value.
func (a *Amp) SendCommand(code ...Command) (byte, error) {
	command := []byte{224, 82, 83, 33}

	for _, each := range code {
		command = append(command, byte(each))
	}

	_, err := a.port.Write(command)

	if err != nil {
		log.Fatalf("sendCommand %f", err)
		return 0, err
	}

	return 0, nil
}
