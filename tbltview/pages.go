package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	pages := tview.NewPages()

	// modal
	modal := tview.NewModal()
	modal.SetText(fmt.Sprintf("This is page %d. Choose where to go next.", 0)).
		AddButtons([]string{"Next", "Quit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				pages.SwitchToPage("page-1")
			} else {
				app.Stop()
			}
		})
	pages.AddPage("page-0", modal, false, true)

	// textview
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetDoneFunc(func(key tcell.Key) {
			pages.SwitchToPage("page-2")
		})
	textView.SetText("this is textview")
	pages.AddPage("page-1", textView, false, false)

	// input field
	inputField := tview.NewInputField().
		SetLabel("Enter a number: ").
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldInteger).
		SetDoneFunc(func(key tcell.Key) {
			pages.SwitchToPage("page-0")
		})
	pages.AddPage("page-2", inputField, false, false)

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
