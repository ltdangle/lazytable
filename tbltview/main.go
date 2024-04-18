package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"

	// "strings"

	"github.com/gdamore/tcell/v2"
	// "github.com/k0kubun/pp/v3"
	"github.com/rivo/tview"
)

// Column type. (Column settings).
type Column struct {
	width int
}

// DataTable type.
type DataTable struct {
	tview.TableContentReadOnly
	Data       [][]*tview.TableCell
	Columns    []Column
	Selection  *Selection
	currentRow int
	currentCol int
}

func NewDataTable() *DataTable {
	return &DataTable{Selection: NewSelection()}
}
func (d *DataTable) SetCurrentRow(row int) {
	d.currentRow = row
}
func (d *DataTable) SetCurrentCol(col int) {
	d.currentCol = col
}
func (d *DataTable) CurrentRow() int {
	return d.currentRow
}
func (d *DataTable) CurrentCol() int {
	return d.currentCol
}
func (d *DataTable) AddDataRow(dataRow []*tview.TableCell) {
	d.Data = append(d.Data, dataRow)
}
func (d *DataTable) GetCell(row, column int) *tview.TableCell {
	// Draw table coordinates.
	if row == 0 { // This is top row with col numbers.
		if column == 0 {
			return NewCell()
		}
		cell := NewCell()
		cell.SetAttributes(tcell.AttrDim)
		cell.SetAlign(1) //AlignCenter
		cell.SetText(strconv.Itoa(column))

		// Highlight row header cell for current selection.
		if column == dataTbl.currentCol+1 {
			cell.SetAttributes(tcell.AttrBold)
			cell.SetAttributes(tcell.AttrUnderline)
			return cell
		}
		return cell
	}

	if column == 0 { // This is leftmost row with row numbers.
		cell := NewCell()
		cell.SetAttributes(tcell.AttrDim)
		cell.SetText(strconv.Itoa(row))

		// Highlight col header cell for current selection.
		if row == dataTbl.currentRow+1 {
			cell.SetAttributes(tcell.AttrBold)
			cell.SetAttributes(tcell.AttrUnderline)
			return cell
		}
		return cell
	}

	// There no data in these coordinates.
	if row >= len(d.Data) {
		cell := NewCell()
		cell.SetText("unchartered")
		return cell
	}
	if column >= len(d.Data[0]) {
		cell := NewCell()
		cell.SetText("unchartered")
		return cell
	}

	return d.Data[row-1][column-1]
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

func (d *DataTable) DeleteRow(row int) {
	if row < 0 || row >= len(d.Data) {
		return // Invalid row index
	}
	d.Data = append(d.Data[:row], d.Data[row+1:]...)
}

func (d *DataTable) DeleteColumn(col int) {
	if col < 0 || len(d.Data) == 0 || col >= len(d.Data[0]) {
		return // Invalid column index
	}
	for i := range d.Data {
		d.Data[i] = append(d.Data[i][:col], d.Data[i][col+1:]...)
	}
}

func (d *DataTable) AddRow() {
	rowSize := len(d.Data[0])
	newRow := make([]*tview.TableCell, rowSize)
	for i := range newRow {
		newRow[i] = NewCell() // initialize all cells in the new row with empty strings
	}
	d.Data = append(d.Data, newRow)
}

func (d *DataTable) AddColumn() {
	for i := range d.Data {
		d.Data[i] = append(d.Data[i], NewCell()) // add an empty string DataCell to the end of each row
	}
}

func (d *DataTable) DeleteSelection() {
	if d.Selection.kind == ROW_SELECTED {
		d.DeleteRow(d.Selection.value)
	} else if d.Selection.kind == COL_SELECTED {
		d.DeleteColumn(d.Selection.value)
	}
}
func (d *DataTable) GetCurrentCell() *tview.TableCell {
	return d.Data[d.CurrentRow()][d.CurrentCol()]
}

// Sort column  string values.

func (d *DataTable) SortColStrAsc(col int) {
	d.sortColumn(col, func(a, b *tview.TableCell) bool {
		return a.Text < b.Text // Compare the text of the cells for ascending order.
	})
}

func (d *DataTable) SortColStrDesc(col int) {
	d.sortColumn(col, func(a, b *tview.TableCell) bool {
		return a.Text > b.Text // Compare the text of the cells for descending order.
	})
}

// Sorts column. Accept column index and a sorter function that
// takes slice of vertical column cells as an argument.
func (d *DataTable) sortColumn(col int, sorter func(a, b *tview.TableCell) bool) {
	// Perform a stable sort to maintain the relative order of other elements.
	sort.SliceStable(d.Data[1:], func(i, j int) bool {
		return sorter(d.Data[i+1][col], d.Data[j+1][col])
	})
}

func NewCell() *tview.TableCell {
	return tview.NewTableCell("")
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
	// Convert string values to cells.
	var dataRow []*tview.TableCell
	for _, val := range record {
		cell := NewCell()
		cell.SetText(val)
		cell.SetMaxWidth(10)
		dataRow = append(dataRow, cell)
	}

	dataTbl.AddDataRow(dataRow)
}

func convertDataToArr(dataTbl *DataTable) [][]string {
	var data [][]string
	for _, row := range dataTbl.Data {
		stringRow := make([]string, len(row))
		for j, cell := range row {
			stringRow[j] = cell.Text
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

// Configure ui elements.
func buildTableWidget() {
	table.
		SetBorders(false).
		SetContent(dataTbl).
		SetSelectable(true, true).
		SetFixed(2, 1).
		Select(1, 1).
		SetSelectedFunc(func(row, col int) {
			app.SetFocus(cellInput)
		}).
		SetSelectionChangedFunc(func(row, col int) {
			// Don't select x,y coordinates.
			if row == 0 {
				dataTbl.SetCurrentRow(1)
				table.Select(dataTbl.CurrentRow(), col)
				return
			}
			if col == 0 {
				dataTbl.SetCurrentCol(1)
				table.Select(row, dataTbl.CurrentCol())
				return
			}

			// Check if the whole row or column is selected.
			selRow, selCol := table.GetSelectable()
			if selRow && !selCol {
				dataTbl.selectRow(row)
				return
			}
			if !selRow && selCol {
				dataTbl.selectCol(col)
				return
			}

			// Select individual cell.
			dataTbl.SetCurrentRow(row - 1) // account for top coordinate row
			dataTbl.SetCurrentCol(col - 1) // account for leftmost coordinates col
			// TODO: encapsulate, somehow
			cellInput.SetLabel(fmt.Sprintf("%d:%d ", row, col))
			cellInput.SetText(dataTbl.GetCurrentCell().Text)
		}).
		SetInputCapture(
			func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Rune() {
				case 'r':
					table.SetSelectable(true, false)
					dataTbl.selectRow(dataTbl.CurrentRow())
				case 'c':
					table.SetSelectable(false, true)
					dataTbl.selectCol(dataTbl.CurrentCol())
				case 's':
					table.SetSelectable(true, true)
				case 'd':
					dataTbl.DeleteSelection()
				case '>': // inclrease column width
					for rowIdx := range dataTbl.Data {
						cell := dataTbl.Data[rowIdx][dataTbl.CurrentCol()]
						cell.SetMaxWidth(cell.MaxWidth + 1)
					}
				case '<': // decrease column width
					for rowIdx := range dataTbl.Data {
						cell := dataTbl.Data[rowIdx][dataTbl.CurrentCol()]
						if cell.MaxWidth == 1 {
							break
						}
						cell.SetMaxWidth(cell.MaxWidth - 1)
					}
				case 'f': // sort string values asc
					dataTbl.SortColStrAsc(dataTbl.CurrentCol())
				case 'F': // sort string values desc
					dataTbl.SortColStrDesc(dataTbl.CurrentCol())
				}
				return event
			})
}

func buildCellInput() {
	cellInput.
		SetLabel(fmt.Sprintf("%d:%d ", dataTbl.CurrentRow(), dataTbl.CurrentCol())).
		SetText(dataTbl.GetCurrentCell().Text).
		SetDoneFunc(func(key tcell.Key) {
			dataTbl.GetCurrentCell().SetText(cellInput.GetText())
			saveDataToFile(*csvFile, dataTbl)
			app.SetFocus(table)
		})

	// TODO: encapsulate, somehow
	cellInput.SetLabel(fmt.Sprintf("%d:%d ", dataTbl.CurrentRow()+1, dataTbl.CurrentCol()+1))
}

var csvFile *string
var dataTbl = NewDataTable()
var table = tview.NewTable()
var app = tview.NewApplication()
var cellInput = tview.NewInputField()
var pages = tview.NewPages()
var modalContents = tview.NewBox()
var bottomBar = tview.NewTextView()

func main() {
	// Parse cli arguments.
	csvFile = flag.String("file", "", "path to csv file")
	flag.Parse()
	if *csvFile == "" {
		log.Fatal("-file not specified")
	}

	// Load csv file data.
	readCsvFile(*csvFile, dataTbl)

	dataTbl.SetCurrentRow(0)
	dataTbl.SetCurrentCol(0)

	buildCellInput()
	buildTableWidget()

	// Configure layout.
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cellInput, 0, 1, false).
		AddItem(table, 0, 25, false).
		AddItem(bottomBar, 0, 1, false)

	flex.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'm':
				pages.ShowPage("modal")
				modalContents.SetTitle("You pressed the m button!")
			case 'R':
				dataTbl.AddRow()
			case 'C':
				dataTbl.AddColumn()
			}
			return event
		})

	// MODAL
	// Returns a new primitive which puts the provided primitive in the center and
	// sets its size to the given width and height.
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false), width, 1, true).
			AddItem(nil, 0, 1, false)
	}

	modalContents.
		SetBorder(true).
		SetTitle("Modal window").
		SetInputCapture(
			func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Rune() {
				case 'q':
					pages.HidePage("modal")
					app.SetFocus(table)
					modalContents.SetTitle("")
				}
				return event
			})

	pages.
		AddPage("background", flex, true, true).
		AddPage("modal", modal(modalContents, 40, 10), true, false)

	bottomBar.SetText("> ")
	if err := app.SetRoot(pages, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}
