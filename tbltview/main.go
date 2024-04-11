package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()
var textView = tview.NewTextView()
var inputField = tview.NewInputField()

func main() {
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
		SetDoneFunc(func(key tcell.Key) {
			fmt.Fprintf(textView, "\ninput: "+inputField.GetText())
			app.SetFocus(table)
		}).
		SetText("Input text")

		// table
	table.
		SetBorders(false).
		SetSelectable(true, true).
		SetContent(data)

	table.
		SetSelectedFunc(func(row, col int) {
			app.SetFocus(inputField)
		}).
		SetSelectionChangedFunc(func(row, col int) {
			data.SelectedRow = row
			data.SelectedCol = col
			inputField.SetLabel(fmt.Sprintf("%d:%d ", data.SelectedRow, data.SelectedCol))
			inputField.SetText(string(data.Data[data.SelectedRow][data.SelectedCol]))
		})

	table.SetSelectable(true, true)

	flex := tview.NewFlex().
		// left
		AddItem(textView, 0, 1, false).
		// center
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(inputField, 0, 1, false).
				AddItem(table, 0, 19, false),
			// AddItem(tview.NewBox().SetBorder(true).SetTitle("Bottom (5 rows)"), 5, 1, false),
			0, 2, false,
		).
		// right
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Right (20 cols)"), 20, 1, false)

	if err := app.SetRoot(flex, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}
