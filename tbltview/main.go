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

// Data type.
type Data struct {
	cells      [][]*tview.TableCell
	lastColumn int

	Columns    []Column
	currentRow int
	currentCol int
}

func NewDataTable() *Data {
	return &Data{}
}
func (t *Data) Clear() {
	t.cells = nil
}
func (t *Data) InsertColumn(column int) {
	for row := range t.cells {
		if column >= len(t.cells[row]) {
			continue
		}
		t.cells[row] = append(t.cells[row], nil)             // Extend by one.
		copy(t.cells[row][column+1:], t.cells[row][column:]) // Shift to the right.
		t.cells[row][column] = &tview.TableCell{}            // New element is an uninitialized table cell.
	}
}
func (t *Data) InsertRow(row int) {
	if row >= len(t.cells) {
		return
	}
	t.cells = append(t.cells, nil)       // Extend by one.
	copy(t.cells[row+1:], t.cells[row:]) // Shift down.
	t.cells[row] = nil                   // New row is uninitialized.
}
func (d *Data) SetCurrentRow(row int) {
	d.currentRow = row
}
func (d *Data) SetCurrentCol(col int) {
	d.currentCol = col
}
func (d *Data) CurrentRow() int {
	return d.currentRow
}
func (d *Data) CurrentCol() int {
	return d.currentCol
}
func (d *Data) AddDataRow(dataRow []*tview.TableCell) {
	d.cells = append(d.cells, dataRow)
}
func (d *Data) GetCell(row, column int) *tview.TableCell {
	// Coordinates are outside our table.
	if row > d.GetRowCount()-1 || column > d.GetColumnCount()-1 {
		return nil
	}

	cell := d.cells[row][column]
	// Draw table coordinates.
	if row == 0 { // This is top row with col numbers.
		if column == 0 {
			return cell
		}
		cell.SetAttributes(tcell.AttrDim)
		cell.SetAlign(1) //AlignCenter

		// Highlight row header cell for current selection.
		if column == dataTbl.currentCol {
			cell.SetAttributes(tcell.AttrBold)
			cell.SetAttributes(tcell.AttrUnderline)
			return cell
		}
		return cell
	}

	if column == 0 { // This is leftmost row with row numbers.
		cell.SetAttributes(tcell.AttrDim)

		// Highlight col header cell for current selection.
		if row == dataTbl.currentRow {
			cell.SetAttributes(tcell.AttrBold)
			cell.SetAttributes(tcell.AttrUnderline)
			return cell
		}
		return cell
	}

	return cell
}

func (d *Data) SetCell(row, column int, cell *tview.TableCell) {
	cell.SetText(strconv.Itoa(row) + " : " + strconv.Itoa(column))
}

func (d *Data) GetRowCount() int {
	return len(d.cells)
}

func (d *Data) GetColumnCount() int {
	return len(d.cells[0])
}

func (d *Data) RemoveRow(row int) {
	if row < 0 || row >= len(d.cells) {
		return // Invalid row index
	}
	d.cells = append(d.cells[:row], d.cells[row+1:]...)
}

func (d *Data) RemoveColumn(col int) {
	if col < 0 || len(d.cells) == 0 || col >= len(d.cells[0]) {
		return // Invalid column index
	}
	for i := range d.cells {
		d.cells[i] = append(d.cells[i][:col], d.cells[i][col+1:]...)
	}
}

func (d *Data) AddEmptyRow() {
	var row []*tview.TableCell
	for i := 0; i < d.GetColumnCount(); i++ {
		row = append(row, NewCell())
	}

	d.cells = append(d.cells, row)

	// Add col header.
	row[0].SetText(fmt.Sprintf("%d", d.GetRowCount()-2))
}

func (d *Data) AddEmptyColumn() {
	counter := 0
	for i := range d.cells {
		d.cells[i] = append(d.cells[i], NewCell()) // add an empty string DataCell to the end of each row
		counter++
	}

	// Set column header.
	d.cells[0][d.GetColumnCount()-1].SetText(fmt.Sprintf("%d", d.GetColumnCount()-2))
}

func (d *Data) GetCurrentCell() *tview.TableCell {
	return d.cells[d.CurrentRow()][d.CurrentCol()]
}

// Sort column  string values.

func (d *Data) SortColStrAsc(col int) {
	d.sortColumn(col, func(a, b *tview.TableCell) bool {
		return a.Text < b.Text // Compare the text of the cells for ascending order.
	})
}

func (d *Data) SortColStrDesc(col int) {
	d.sortColumn(col, func(a, b *tview.TableCell) bool {
		return a.Text > b.Text // Compare the text of the cells for descending order.
	})
}

// Sorts column. Accept column index and a sorter function that
// takes slice of vertical column cells as an argument.
func (d *Data) sortColumn(col int, sorter func(a, b *tview.TableCell) bool) {
	// Perform a stable sort to maintain the relative order of other elements.
	// Account for cols row and header row (+2)
	sort.SliceStable(d.cells[2:], func(i, j int) bool {
		return sorter(d.cells[i+2][col], d.cells[j+2][col])
	})
}

// Factory functions.
func NewCell() *tview.TableCell {
	cell := tview.NewTableCell("")
	cell.SetMaxWidth(10)
	return cell
}

func readCsvFile(fileName string, dataTbl *Data) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error opening file: %s", err.Error())
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	recordCounter := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error opening file: %s", err.Error())
			return
		}

		// Add row header.
		if recordCounter == 0 {
			var header []string
			for colCount := range record {
				header = append(header, fmt.Sprintf("%d", colCount))
			}
			addRecordToDataTable(recordCounter, header, dataTbl)
		}

		// Add record to data table.
		addRecordToDataTable(recordCounter, record, dataTbl)

		recordCounter++
	}

	// Pretty-print top left cell (empty it).
	dataTbl.cells[0][0].SetText("")

	// Pretty-print table header.
	for _, headerCell := range dataTbl.cells[1] {
		headerCell.SetAttributes(tcell.AttrBold)
	}
}

func addRecordToDataTable(recordCount int, record []string, dataTbl *Data) {
	var dataRow []*tview.TableCell

	// Set col header.
	colHead := NewCell()
	colHead.SetText(fmt.Sprintf("%d", recordCount))
	dataRow = append(dataRow, colHead)

	// Add row (record) data.
	for _, val := range record {
		cell := NewCell()
		cell.SetText(val)
		cell.SetMaxWidth(10)
		dataRow = append(dataRow, cell)
	}

	dataTbl.AddDataRow(dataRow)
}

func convertDataToArr(dataTbl *Data) [][]string {
	var data [][]string
	for _, row := range dataTbl.cells[1:] { // account for top col numbers row
		row = row[1:] // account for row numbers col
		stringRow := make([]string, len(row))
		for j, cell := range row {
			stringRow[j] = cell.Text
		}
		data = append(data, stringRow)
	}
	return data
}

func saveDataToFile(path string, dataDataTable *Data) {
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

			// Select individual cell.
			dataTbl.SetCurrentRow(row) // account for top coordinate row
			dataTbl.SetCurrentCol(col) // account for leftmost coordinates col
			// TODO: encapsulate, somehow
			cellInput.SetLabel(fmt.Sprintf("%d:%d ", row, col))
			cellInput.SetText(dataTbl.GetCurrentCell().Text)
		}).
		SetInputCapture(
			func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Rune() {
				case 'r':
					table.SetSelectable(true, false)
				case 'c':
					table.SetSelectable(false, true)
				case 's':
					table.SetSelectable(true, true)
				case 'd':
					row, col := table.GetSelection()
					rowSelctbl, colSelectbl := table.GetSelectable()
					rowSelected := rowSelctbl && !colSelectbl
					colSelected := !rowSelctbl && colSelectbl
					if rowSelected {
						dataTbl.RemoveRow(row)
					} else if colSelected {
						dataTbl.RemoveColumn(col)
					}
				case '>': // inclrease column width
					for rowIdx := range dataTbl.cells {
						cell := dataTbl.cells[rowIdx][dataTbl.CurrentCol()]
						cell.SetMaxWidth(cell.MaxWidth + 1)
					}
				case '<': // decrease column width
					for rowIdx := range dataTbl.cells {
						cell := dataTbl.cells[rowIdx][dataTbl.CurrentCol()]
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

	dataTbl.SetCurrentRow(1)
	dataTbl.SetCurrentCol(1)

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
				dataTbl.AddEmptyRow()
			case 'C':
				dataTbl.AddEmptyColumn()
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
