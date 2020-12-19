package main

import (
	"log"
	"math"
	"parasound/cec"

	"parasound/amp"
)

var (
	myAmp     *amp.Amp
	status    amp.StatusMap
	cecClient *cec.Listener
)

func main() {
	var err error
	myAmp, err = amp.Open("/dev/tty.usbserial-1410")
	defer myAmp.Close()

	statusChannel := make(chan amp.StatusMap)
	cecChannel := make(chan cec.Message)
	cecClient = cec.Open(cecChannel)

	defer cecClient.Close()

	status = make(amp.StatusMap)

	if err != nil {
		log.Fatalf("Unable to open port %f", err)
	}

	go myAmp.Poll(statusChannel)

	myAmp.SendCommand(amp.CommandGetStatus)

	for {
		select {
		case message := <-cecChannel:
			handleCecMessage(message)
		case ampStatus := <-statusChannel:
			handleAmpMessage(ampStatus)
		}
	}

	// ui.Run(myAmp, statusChannel)
}

func handleAmpMessage(ampStatus amp.StatusMap) {
	// Update master status map
	for key, value := range ampStatus {
		status[key] = value
	}

	if volume, hasVolume := ampStatus[amp.StatusVolume]; hasVolume {
		var scaledVolume = int(math.Round(((float64(volume) - 10.0) / 96.0) * 100.0))
		values := []uint{uint(scaledVolume)}

		if muted, hasMuted := status[amp.StatusMute]; hasMuted {
			if muted == 1 {
				values[0] += 128
			}
		}

		cecClient.Send(cec.Message{
			Source:  cec.TypeAudio,
			Target:  cec.TypeTV,
			Message: cec.MessageReportAudio,
			Values:  values,
		})
	}
}

func handleCecMessage(message cec.Message) {
	if message.Target == cec.TypeAudio {
		switch message.Message {
		case cec.MessageControlPressed:
			switch message.Values[0] {
			case cec.ButtonVolumeUp:
				myAmp.SendCommand(amp.CommandVolumeUp)
			case cec.ButtonVolumeDown:
				myAmp.SendCommand(amp.CommandVolumeDown)
			case cec.ButtonMute:
				myAmp.SendCommand(amp.CommandMuteToggle)
			}
		}
	}
}
