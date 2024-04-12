package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/k0kubun/pp/v3"
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

func readCsvFile(fileName string, dataTbl *DataTable) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error opening file: %s", err.Error())
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error opening file: %s", err.Error())
			return
		}
		pp.Print(record)
	}
}

var dataTbl = NewDataTable()
var table = tview.NewTable()
var app = tview.NewApplication()
var cellInput = tview.NewInputField()

func main() {
	// Parse cli arguments.
	csvFile := flag.String("file", "", "path to csv file")
	flag.Parse()
	if *csvFile == "" {
		log.Fatal("-file not specified")
	}

	// Load csv file.
	readCsvFile(*csvFile, dataTbl)
	return

	dataTbl.Data = [][]DataCell{
		{DataCell("one"), DataCell("two"), DataCell("three")},
		{DataCell("one"), DataCell("two tee\n\nto two"), DataCell("three")},
		{DataCell("one"), DataCell("two"), DataCell("three")},
		{DataCell("one"), DataCell("two"), DataCell("three")},
		{DataCell("one"), DataCell("two"), DataCell("three")},
	}
	dataTbl.SelectedRow = 0
	dataTbl.SelectedCol = 0

	// Configure cell input widget.
	cellInput.
		SetLabel(fmt.Sprintf("%d:%d ", dataTbl.SelectedRow, dataTbl.SelectedCol)).
		SetText(string(dataTbl.Data[dataTbl.SelectedRow][dataTbl.SelectedCol])).
		SetDoneFunc(func(key tcell.Key) {
			dataTbl.Data[dataTbl.SelectedRow][dataTbl.SelectedCol] = DataCell(cellInput.GetText())
			app.SetFocus(table)
		})

	// Configure table widget.
	table.
		SetBorders(false).
		SetSelectable(true, true).
		SetContent(dataTbl).
		SetSelectedFunc(func(row, col int) {
			app.SetFocus(cellInput)
		}).
		SetSelectionChangedFunc(func(row, col int) {
			dataTbl.SelectedRow = row
			dataTbl.SelectedCol = col
			cellInput.SetLabel(fmt.Sprintf("%d:%d ", dataTbl.SelectedRow, dataTbl.SelectedCol))
			cellInput.SetText(string(dataTbl.Data[dataTbl.SelectedRow][dataTbl.SelectedCol]))
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
