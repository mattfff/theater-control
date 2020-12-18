package main

import (
	"log"
	"parasound/cec"

	"parasound/amp"
)

var (
	myAmp *amp.Amp
)

func main() {
	var err error
	myAmp, err = amp.Open("/dev/ttyUSB0")
	defer myAmp.Close()

	statusChannel := make(chan amp.StatusMap)
	cecChannel := make(chan string)
	cecClient := cec.Open(cecChannel)

	defer cecClient.Close()

	// go handleCecOutput(cecClient, myAmp)

	if err != nil {
		log.Fatalf("Unable to open port %f", err)
	}

	go myAmp.Poll(statusChannel)

	myAmp.SendCommand(amp.CommandGetStatus)

	for {
		select {
		case cecOutput := <-cecChannel:
			log.Println(cecOutput)
		}
	}

	// ui.Run(myAmp, statusChannel)
}
