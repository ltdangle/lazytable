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
			{"one", "two tee\n\nto two", "three"},
			{"one", "two", "three"},
			{"one", "two", "three"},
			{"one", "two", "three"},
		},
	}
}
func (d *TableData) GetCell(row, column int) *tview.TableCell {
	cell := tview.NewTableCell("")
	cell.MaxWidth =10 
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

func main() {
	data := NewTableData()
	table := tview.NewTable().
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
	if err := tview.NewApplication().SetRoot(table, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
