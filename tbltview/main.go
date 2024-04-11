package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()
var inputField = tview.NewInputField()

func main() {
	// textview

	// input field
	inputField.SetLabel("Enter a number: ").
		SetDoneFunc(func(key tcell.Key) {
			data.Data[data.SelectedRow][data.SelectedCol] = Cell(inputField.GetText())
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
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(inputField, 0, 1, false).
				AddItem(table, 0, 19, false),
			0, 2, false,
		)

	if err := app.SetRoot(flex, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}
