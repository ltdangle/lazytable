package main

import (
	"math"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TableData struct {
	tview.TableContentReadOnly
	Data [][]string
}

func NewTableData() *TableData {
	return &TableData{
		Data: [][]string{
			{"one", "two", "three"},
			{"one", "two tee\n\nto two", "three"},
			{"one", "two", "three"},
			{"one", "two", "three"},
			{"one", "two", "three"},
		},
	}
}
func (d *TableData) GetCell(row, column int) *tview.TableCell {
	cell := tview.NewTableCell("")
	cell.MaxWidth = 10
	if row >= len(d.Data) {
		cell.SetText("unchartered")
	} else if column >= len(d.Data[0]) {
		cell.SetText("unchartered")
	} else {
		cell.SetText(d.Data[row][column])
	}
	return cell
}

func (d *TableData) SetCell(row, column int, cell *tview.TableCell) {
	cell.SetText(strconv.Itoa(row) + " : " + strconv.Itoa(column))
}
func (d *TableData) GetRowCount() int {
	return math.MaxInt64
}

func (d *TableData) GetColumnCount() int {
	return math.MaxInt64
}

var data = NewTableData()
var table = tview.NewTable()
var app = tview.NewApplication()

func main() {
	table.
		SetBorders(false).
		SetSelectable(true, true).
		SetContent(data)

	table.
		SetSelectedFunc(func(row, column int) {
			data.Data[row][column] = "changed"
		}).
		SetSelectionChangedFunc(func(row, column int) {
			data.Data[row][column] = "selected"
		})

	table.SetSelectable(true, true)
	table.SetInputCapture(tableInputCapture)

	if err := app.SetRoot(table, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}

func tableInputCapture(event *tcell.EventKey) *tcell.EventKey {
	selectedRowIndex, selectedColumnIndex := table.GetSelection()
	rune := event.Rune()
	if rune == '1' {
		data.Data[0][0] = strconv.Itoa(selectedRowIndex) + ":" + strconv.Itoa(selectedColumnIndex)
	}
	if rune == '2' {
		StartEditingCell(selectedRowIndex, selectedColumnIndex)
	}
	return event
}

func StartEditingCell(row int, col int) {
	table.SetInputCapture(nil)

	cell := table.GetCell(row, col)
	inputField := tview.NewInputField()
	inputField.SetText(cell.Text)
	inputField.SetFieldBackgroundColor(tview.Styles.PrimaryTextColor)
	inputField.SetFieldTextColor(tcell.ColorBlack)

	inputField.SetDoneFunc(func(key tcell.Key) {
		currentValue := cell.Text
		newValue := inputField.GetText()
		if key == tcell.KeyEnter {
			if currentValue != newValue {
				cell.SetText(inputField.GetText())
			}
		} else if key == tcell.KeyTab {
			nextEditableColumnIndex := col + 1

			if nextEditableColumnIndex <= table.GetColumnCount()-1 {
				cell.SetText(inputField.GetText())
				table.Select(row, nextEditableColumnIndex)
			}
		} else if key == tcell.KeyBacktab {
			nextEditableColumnIndex := col - 1

			if nextEditableColumnIndex >= 0 {
				cell.SetText(inputField.GetText())
				table.Select(row, nextEditableColumnIndex)
			}
		}

		if key == tcell.KeyEnter || key == tcell.KeyEscape {
			table.SetInputCapture(tableInputCapture)
			app.SetFocus(table)
		}
	})

	x, y, width := cell.GetLastPosition()
	inputField.SetRect(x, y, width+1, 1)
	// table.Page.AddPage("edit", inputField, false, true)
	app.SetFocus(inputField)
}
