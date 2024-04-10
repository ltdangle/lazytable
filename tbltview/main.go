package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()
var modal = tview.NewModal()
var textView = tview.NewTextView()
var inputField = tview.NewInputField()


func main() {
	// modal
	modal.SetText(fmt.Sprintf("This is page %d. Choose where to go next.", 0)).
		AddButtons([]string{"Next", "Quit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				fmt.Fprintf(textView, "ok button on modal\n")
			} else {
				app.Stop()
			}
		})
		// textview
	textView.SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetDoneFunc(func(key tcell.Key) {
			fmt.Fprintf(textView, "DoneFunc on textView\n")
		})
	textView.SetText("this is textview")

	// input field
	inputField.SetLabel("Enter a number: ").
		SetFieldWidth(10).
		// SetAcceptanceFunc(tview.InputFieldInteger).
		SetDoneFunc(func(key tcell.Key) {
			fmt.Fprintf(textView, "DoneFunc on inputField\n")
		})
		// table
	table.
		SetBorders(false).
		SetSelectable(true, true).
		SetContent(data)

	table.
		SetSelectedFunc(func(row, column int) {
			StartEditingCell(row, column)
			// data.Data[row][column] = fmt.Sprintf("x: %d, y: %d", x, y)
		}).
		SetSelectionChangedFunc(func(row, column int) {
			// data.Data[row][column] = "selected"
		})

	table.SetSelectable(true, true)
	table.SetInputCapture(tableInputCapture)

	flex := tview.NewFlex().
		AddItem(textView, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(inputField, 0, 1, false).
			AddItem(table, 0, 3, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Bottom (5 rows)"), 5, 1, false), 0, 2, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Right (20 cols)"), 20, 1, false)

	if err := app.SetRoot(flex, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}
