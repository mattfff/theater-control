package main

import (
	"log"
	"parasound/cec"
	"parasound/ui"

	"parasound/amp"
)

var (
	myAmp *amp.Amp
)

func handleCecOutput(cecClient *cec.Listener, myAmp *amp.Amp) {
	for {
		select {
		case output := <-cecClient.Stdout:
			myAmp.SendCommand(amp.Command(output[0]))
		}
	}
}

func main() {
	var err error
	myAmp, err = amp.Open("/dev/tty.usbserial-1410")
	defer myAmp.Close()

	statusChannel := make(chan amp.StatusMap)
	// cecClient := cec.Open()

	// go handleCecOutput(cecClient, myAmp)

	if err != nil {
		log.Fatalf("Unable to open port %f", err)
	}

	go myAmp.Poll(statusChannel)

	myAmp.SendCommand(amp.CommandGetStatus)

	ui.Run(myAmp, statusChannel)
}
