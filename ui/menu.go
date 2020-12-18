package ui

import (
	"fmt"
	"parasound/amp"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type statusOutputs map[amp.StatusFlag]*tview.TableCell

var (
	app       *tview.Application
	table     *tview.Table
	container *tview.Flex
	form      *tview.Form
	outputs   statusOutputs
	myAmp     *amp.Amp
)

func Run(pAmp *amp.Amp, statusChannel chan amp.StatusMap) {
	myAmp = pAmp

	app = tview.NewApplication()
	container = tview.NewFlex()
	table = tview.NewTable()
	form = tview.NewForm()

	outputs = make(statusOutputs)

	container.SetBorder(true)

	container.AddItem(table, 0, 1, false)
	container.AddItem(form, 0, 1, true)

	var r = 0
	for _, flag := range amp.StatusFlags {
		table.SetCell(r, 0, tview.NewTableCell(amp.StatusLabel[flag]).SetTextColor(tcell.ColorRed).SetAlign(tview.AlignRight))

		cell := tview.NewTableCell("").SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft)
		table.SetCell(r, 1, cell)

		outputs[flag] = cell
		r++
	}

	form.AddButton("Power", func() { myAmp.SendCommand(amp.CommandPowerToggle) })
	form.AddButton("Vol +", func() { myAmp.SendCommand(amp.CommandVolumeUp) })
	form.AddButton("Vol -", func() { myAmp.SendCommand(amp.CommandVolumeDown) })
	// form.AddButton("Mute", func() { myAmp.SendCommand(amp.CommandMuteToggle) })
	form.AddButton("THX", func() { myAmp.SendCommand(amp.CommandThxToggle) })
	// form.AddButton("Surround +", func() { myAmp.SendCommand(amp.CommandSurroundPlus) })
	// form.AddButton("Surround -", func() { myAmp.SendCommand(amp.CommandSurroundMinus) })
	form.AddButton("Test", func() { myAmp.SendCommand(amp.CommandTestNoise) })
	// form.AddButton("Menu", func() { myAmp.SendCommand(amp.CommandSetupMenuToggle) })
	// form.AddButton("↑", func() { myAmp.SendCommand(amp.CommandCursorUp) })
	// form.AddButton("↓", func() { myAmp.SendCommand(amp.CommandCursorDown) })
	// form.AddButton("←", func() { myAmp.SendCommand(amp.CommandCursorLeft) })
	// form.AddButton("→", func() { myAmp.SendCommand(amp.CommandCursorRight) })

	go HandleInput(statusChannel)

	app.SetRoot(container, true).SetFocus(form).Run()
}

func HandleInput(statusChannel chan amp.StatusMap) {
	input := make(chan string)

	for {
		select {
		case status := <-statusChannel:
			app.QueueUpdateDraw(func() {
				var r = 0
				for _, flag := range amp.StatusFlags {
					table.GetCell(r, 1).SetText(fmt.Sprintf("%d", status[flag]))
					r++
				}
			})
		case command := <-input:
			converted, err := strconv.Atoi(strings.TrimRight(command, "\n"))
			if err == nil {
				myAmp.SendCommand(amp.Command(converted))
			} else {
				fmt.Printf("Err: %v", err)
			}
		}
	}
}
