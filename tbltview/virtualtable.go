package main

import (
	"math"
	"strconv"

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
			{"one", "two", "three"},
			{"one", "two", "three"},
			{"one", "two", "three"},
			{"one", "two", "three"},
		},
	}
}
func (d *TableData) GetCell(row, column int) *tview.TableCell {
	if row >= len(d.Data) {
		return tview.NewTableCell("unchartered")
	}
	if column >= len(d.Data[0]) {
		return tview.NewTableCell("unchartered")
	}
	return tview.NewTableCell(d.Data[row][column])
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

func main() {
	data := NewTableData()
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, true).
		SetContent(data)

	table.Select(0, 0).SetFixed(1, 1).SetSelectedFunc(func(row int, column int) {
		data.Data[row][column] = "selected"
	})

	table.SetSelectable(true, true)
	if err := tview.NewApplication().SetRoot(table, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
