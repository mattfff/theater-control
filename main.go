package main

import (
	"log"
	"math"
	"parasound/cec"
	"time"

	"parasound/amp"
)

var (
	myAmp     *amp.Amp
	status    amp.StatusMap
	cecClient *cec.Listener
)

func main() {
	var err error
	myAmp, err = amp.Open("/dev/ttyUSB0")
	defer myAmp.Close()

	statusChannel := make(chan amp.StatusMap)
	cecChannel := make(chan cec.Message)
	cecClient = cec.Open()

	defer cecClient.Close()

	status = make(amp.StatusMap)

	if err != nil {
		log.Fatalf("Unable to open port %f", err)
	}

	go myAmp.Poll(statusChannel)
	go cecClient.Start(cecChannel)

	myAmp.SendCommand(amp.CommandGetStatus)

	ticker := pollTVPower()
	defer ticker.Stop()

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
		case cec.MessagePower:
			switch message.Values[0] {
			case cec.PowerStatusOn:
				if status[amp.StatusPower] == 0 {
					myAmp.SendCommand(amp.CommandPowerOn)
				}
			case cec.PowerStatusStandby:
				if status[amp.StatusPower] == 1 {
					myAmp.SendCommand(amp.CommandPowerOff)
				}
			}
		}
	}
}

func pollTVPower() *time.Ticker {
	timer := time.NewTicker(10000 * time.Millisecond)

	go func() {
		for {
			select {
			case <-timer.C:
				cecClient.Send(cec.Message{
					Source:  cec.TypeAudio,
					Target:  cec.TypeTV,
					Message: cec.MessageGivePower,
				})
			}
		}
	}()

	return timer
}
