package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// DataCell type.
type DataCell string

func (c DataCell) String() string {
	return string(c)
}

// DataTable type.
type DataTable struct {
	tview.TableContentReadOnly
	Data        [][]DataCell
	SelectedRow int
	SelectedCol int
}

func NewDataTable() *DataTable {
	return &DataTable{}
}

func (d *DataTable) GetCell(row, column int) *tview.TableCell {
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

func (d *DataTable) SetCell(row, column int, cell *tview.TableCell) {
	cell.SetText(strconv.Itoa(row) + " : " + strconv.Itoa(column))
}

func (d *DataTable) GetRowCount() int {
	return len(d.Data)
}

func (d *DataTable) GetColumnCount() int {
	return len(d.Data[0])
}

var data = NewDataTable()
var table = tview.NewTable()
var app = tview.NewApplication()
var cellInput = tview.NewInputField()

func main() {
	csvFile := flag.String("file", "", "path to csv file")
	flag.Parse()
	if *csvFile == "" {
		log.Fatal("-file not specified")
	}

	data.Data = [][]DataCell{
		{DataCell("one"), DataCell("two"), DataCell("three")},
		{DataCell("one"), DataCell("two tee\n\nto two"), DataCell("three")},
		{DataCell("one"), DataCell("two"), DataCell("three")},
		{DataCell("one"), DataCell("two"), DataCell("three")},
		{DataCell("one"), DataCell("two"), DataCell("three")},
	}
	data.SelectedRow = 0
	data.SelectedCol = 0

	// Configure cell input widget.
	cellInput.
		SetLabel(fmt.Sprintf("%d:%d ", data.SelectedRow, data.SelectedCol)).
		SetText(string(data.Data[data.SelectedRow][data.SelectedCol])).
		SetDoneFunc(func(key tcell.Key) {
			data.Data[data.SelectedRow][data.SelectedCol] = DataCell(cellInput.GetText())
			app.SetFocus(table)
		})

	// Configure table widget.
	table.
		SetBorders(false).
		SetSelectable(true, true).
		SetContent(data).
		SetSelectedFunc(func(row, col int) {
			app.SetFocus(cellInput)
		}).
		SetSelectionChangedFunc(func(row, col int) {
			data.SelectedRow = row
			data.SelectedCol = col
			cellInput.SetLabel(fmt.Sprintf("%d:%d ", data.SelectedRow, data.SelectedCol))
			cellInput.SetText(string(data.Data[data.SelectedRow][data.SelectedCol]))
		})

	table.SetSelectable(true, true)

	// Configure layout.
	flex := tview.NewFlex().
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(cellInput, 0, 1, false).
				AddItem(table, 0, 19, false),
			0, 2, false,
		)

	if err := app.SetRoot(flex, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}
