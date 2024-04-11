package main

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Cell string

func (c Cell) String() string {
	return string(c)
}

type TableData struct {
	tview.TableContentReadOnly
	Data        [][]Cell
	SelectedRow int
	SelectedCol int
}

func NewTableData() *TableData {
	return &TableData{}
}
func (d *TableData) GetCell(row, column int) *tview.TableCell {
	cell := tview.NewTableCell("")
	cell.MaxWidth = 10
	if row >= len(d.Data) {
		cell.SetText("unchartered")
	} else if column >= len(d.Data[0]) {
		cell.SetText("unchartered")
	} else {
		cell.SetText(string(d.Data[row][column]))
	}
	return cell
}

func (d *TableData) SetCell(row, column int, cell *tview.TableCell) {
	cell.SetText(strconv.Itoa(row) + " : " + strconv.Itoa(column))
}
func (d *TableData) GetRowCount() int {
	return len(d.Data)
}

func (d *TableData) GetColumnCount() int {
	return len(d.Data[0])
}

var data = NewTableData()
var table = tview.NewTable()
var app = tview.NewApplication()
var inputField = tview.NewInputField()

func main() {
	data.Data = [][]Cell{
		{Cell("one"), Cell("two"), Cell("three")},
		{Cell("one"), Cell("two tee\n\nto two"), Cell("three")},
		{Cell("one"), Cell("two"), Cell("three")},
		{Cell("one"), Cell("two"), Cell("three")},
		{Cell("one"), Cell("two"), Cell("three")},
	}

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
		SetContent(data).
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
