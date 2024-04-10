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
	Page *tview.Pages
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
// var app = tview.NewApplication()
var pages = tview.NewPages()

func mainn() {
	pages.SetRect(0, 0, 20, 20)
	pages.AddPage("table", table, false, false)
	pages.SwitchToPage("table")

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

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
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
				data.Data[row][col] = inputField.GetText()
			}
			table.SetInputCapture(tableInputCapture)
			app.SetFocus(table)
		}
	})

	inputField.SetRect(20, 20, 10, 1)
	pages.AddPage("edit", inputField, false, true)
	app.SetFocus(inputField)
}
