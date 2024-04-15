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
	// "github.com/k0kubun/pp/v3"
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
	Data       [][]DataCell
	Selection  *Selection
	CurrentRow int
	CurrentCol int
}

func NewDataTable() *DataTable {
	return &DataTable{Selection: NewSelection()}
}
func (d *DataTable) AddDataRow(dataRow []DataCell) {
	d.Data = append(d.Data, dataRow)
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

func (d *DataTable) selectRow(row int) {
	d.Selection.kind = ROW_SELECTED
	d.Selection.value = row
	cellInput.SetLabel("Selected row")
	cellInput.SetText(strconv.Itoa(row))
}

func (d *DataTable) selectCol(col int) {
	d.Selection.kind = COL_SELECTED
	d.Selection.value = col
	cellInput.SetLabel("Selected col")
	cellInput.SetText(strconv.Itoa(col))
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

		// Add record to data table.
		addRecordToDataTable(record, dataTbl)

	}
}
func addRecordToDataTable(record []string, dataTbl *DataTable) {
	// Convert []string to []DataCell
	var dataRow []DataCell
	for _, strCell := range record {
		dataRow = append(dataRow, DataCell(strCell))
	}

	dataTbl.AddDataRow(dataRow)
}
func convertDataToArr(dataTbl *DataTable) [][]string {
	var data [][]string
	for _, row := range dataTbl.Data {
		stringRow := make([]string, len(row))
		for j, cell := range row {
			stringRow[j] = cell.String()
		}
		data = append(data, stringRow)
	}
	return data
}

func saveDataToFile(path string, dataDataTable *DataTable) {
	// Truncates file.
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	arr := convertDataToArr(dataDataTable)
	if err := writer.WriteAll(arr); err != nil {
		panic(err)
	}

}

// Selection.
const (
	ROW_SELECTED = "row selected"
	COL_SELECTED = "col selected"
)

type Selection struct {
	kind  string
	value int
}

func NewSelection() *Selection {
	return &Selection{}
}
func (s *Selection) DeleteSelection() {
	// TODO: delete selected row / col from data.Data
}

var csvFile *string
var dataTbl = NewDataTable()
var table = tview.NewTable()
var app = tview.NewApplication()
var cellInput = tview.NewInputField()

func main() {
	// Parse cli arguments.
	csvFile = flag.String("file", "", "path to csv file")
	flag.Parse()
	if *csvFile == "" {
		log.Fatal("-file not specified")
	}

	// Load csv file data.
	readCsvFile(*csvFile, dataTbl)

	dataTbl.CurrentRow = 0
	dataTbl.CurrentCol = 0

	// Configure cell input widget.
	cellInput.
		SetLabel(fmt.Sprintf("%d:%d ", dataTbl.CurrentRow, dataTbl.CurrentCol)).
		SetText(string(dataTbl.Data[dataTbl.CurrentRow][dataTbl.CurrentCol])).
		SetDoneFunc(func(key tcell.Key) {
			dataTbl.Data[dataTbl.CurrentRow][dataTbl.CurrentCol] = DataCell(cellInput.GetText())
			saveDataToFile(*csvFile, dataTbl)
			app.SetFocus(table)
		})

	// Configure table widget.
	table.
		SetBorders(false).
		SetContent(dataTbl).
		SetSelectable(true, true).
		SetSelectedFunc(func(row, col int) {
			app.SetFocus(cellInput)
		}).
		SetSelectionChangedFunc(func(row, col int) {
			selRow, selCol := table.GetSelectable()
			if selRow && !selCol {
				dataTbl.selectRow(row)
				return
			}
			if !selRow && selCol {
				dataTbl.selectCol(col)
				return
			}
			dataTbl.CurrentRow = row
			dataTbl.CurrentCol = col
			cellInput.SetLabel(fmt.Sprintf("%d:%d ", dataTbl.CurrentRow, dataTbl.CurrentCol))
			cellInput.SetText(string(dataTbl.Data[dataTbl.CurrentRow][dataTbl.CurrentCol]))
		}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'r':
				table.SetSelectable(true, false)
				dataTbl.selectRow(dataTbl.CurrentRow)
			case 'c':
				table.SetSelectable(false, true)
				dataTbl.selectCol(dataTbl.CurrentCol)
			case 's':
				table.SetSelectable(true, true)
			}
			return event
		})

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
