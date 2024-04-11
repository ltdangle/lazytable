package main

import (
	"strconv"

	"github.com/rivo/tview"
)

type Cell string

func (c Cell) String() string {
	return string(c)
}

type TableData struct {
	tview.TableContentReadOnly
	Data      [][]Cell
	SelectedRow int
	SelectedCol int
}

func NewTableData() *TableData {
	return &TableData{
		Data: [][]Cell{
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
