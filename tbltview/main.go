// This is simple tui spreadsheet program with vim keybindings.
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

// Data type.
type Data struct {
	cells      [][]*tview.TableCell
	currentRow int
	currentCol int
}

func NewData() *Data {
	return &Data{}
}
func (t *Data) Clear() {
	t.cells = nil
}
func (d *Data) InsertColumn(column int) {
	for row := range d.cells {
		if column >= len(d.cells[row]) {
			continue
		}
		d.cells[row] = append(d.cells[row], nil)             // Extend by one.
		copy(d.cells[row][column+1:], d.cells[row][column:]) // Shift to the right.
		d.cells[row][column] = NewCell()
		d.drawXYCoordinates()
	}
}
func (d *Data) InsertRow(row int) {
	if row >= d.GetRowCount() {
		return
	}

	d.cells = append(d.cells, nil)       // Extend by one.
	copy(d.cells[row+1:], d.cells[row:]) // Shift down.
	d.cells[row] = d.createRow()         // New row is initialized.
	d.drawXYCoordinates()
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
		if column == data.currentCol {
			cell.SetAttributes(tcell.AttrBold)
			cell.SetAttributes(tcell.AttrUnderline)
			return cell
		}
		return cell
	}

	if column == 0 { // This is leftmost row with row numbers.
		cell.SetAttributes(tcell.AttrDim)

		// Highlight col header cell for current selection.
		if row == data.currentRow {
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
	if d.GetRowCount() == 2 {
		return
	}
	if row <= 0 || row >= d.GetRowCount() {
		return // Invalid row index
	}
	d.cells = append(d.cells[:row], d.cells[row+1:]...)
	d.drawXYCoordinates()
}

func (d *Data) RemoveColumn(col int) {
	if d.GetColumnCount() == 2 {
		return
	}
	if col <= 0 || col >= d.GetColumnCount() {
		return // Invalid column index
	}
	for i := range d.cells {
		d.cells[i] = append(d.cells[i][:col], d.cells[i][col+1:]...)
	}
	d.drawXYCoordinates()
}

func (d *Data) AddEmptyRow() {
	row := d.createRow()
	d.cells = append(d.cells, row)
	d.drawXYCoordinates()
}
func (d *Data) createRow() []*tview.TableCell {
	var row []*tview.TableCell
	for i := 0; i < d.GetColumnCount(); i++ {
		row = append(row, NewCell())
	}
	return row
}

func (d *Data) AddEmptyColumn() {
	counter := 0
	for i := range d.cells {
		d.cells[i] = append(d.cells[i], NewCell()) // add an empty string DataCell to the end of each row
		counter++
	}
	d.drawXYCoordinates()
}

func (d *Data) GetCurrentCell() *tview.TableCell {
	// Check of out of bounds values.
	if d.CurrentRow() >= d.GetRowCount() {
		return NewCell()
	}
	if d.CurrentCol() >= d.GetColumnCount() {
		return NewCell()
	}

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
func (d *Data) drawXYCoordinates() {
	for rowIdx := range d.cells {
		d.cells[rowIdx][0].SetText(fmt.Sprintf("%d", rowIdx-1))
	}
	for colIdx, col := range d.cells[0] {
		col.SetText(fmt.Sprintf("%d", colIdx-1))
	}

	d.cells[0][0].SetText("")
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
		SetContent(data).
		SetSelectable(true, true).
		SetFixed(2, 1).
		Select(1, 1).
		SetSelectedFunc(func(row, col int) {
			app.SetFocus(cellInput)
		}).
		SetSelectionChangedFunc(
			func(row, col int) {
				// Don't select x,y coordinates.
				if row == 0 {
					data.SetCurrentRow(1)
					table.Select(data.CurrentRow(), col)
					return
				}
				if col == 0 {
					data.SetCurrentCol(1)
					table.Select(row, data.CurrentCol())
					return
				}

				// Select individual cell.
				data.SetCurrentRow(row) // account for top coordinate row
				data.SetCurrentCol(col) // account for leftmost coordinates col
				// TODO: encapsulate, somehow
				cellInput.SetLabel(fmt.Sprintf("%d:%d ", row-1, col-1))
				cellInput.SetText(data.GetCurrentCell().Text)
			}).
		SetInputCapture(
			func(event *tcell.EventKey) *tcell.EventKey {
				bottomBar.SetText(fmt.Sprintf("rune: %v, key: %v, modifier: %v, name: %v", event.Rune(), event.Key(), event.Modifiers(), event.Name()))
				row, col := table.GetSelection()
				rowSelectable, colSelectable := table.GetSelectable()
				rowSelected := rowSelectable && !colSelectable
				colSelected := !rowSelectable && colSelectable

				rune := event.Rune()
				key := event.Key()

				switch rune {
				case 'V': // Select row.
					table.SetSelectable(true, false)
				case 22:
					if key == 22 { // Rune 22, key 22 = CTRL+V. Select column.
						table.SetSelectable(false, true)
					}
				case 0:
					if key == 27 { // Rune 0, key 27 = ESC.  Select individual cell.
						table.SetSelectable(true, true)
					}
				case 'd':
					if rowSelected {
						data.RemoveRow(row)
						if row == data.GetRowCount() { // Last row deleted, shift selection up.
							if data.GetRowCount() > 0 {
								table.Select(data.GetRowCount()-1, col)
							}
						}
					} else if colSelected {
						data.RemoveColumn(col)
						if col == data.GetColumnCount() { // Last column deleted, shift selection left.
							if data.GetColumnCount() > 0 {
								table.Select(row, data.GetColumnCount()-1)
							}
						}
					}
				case '>': // Increase column width.
					for rowIdx := range data.cells {
						cell := data.cells[rowIdx][data.CurrentCol()]
						cell.SetMaxWidth(cell.MaxWidth + 1)
					}
				case '<': // Decrease column width.
					for rowIdx := range data.cells {
						cell := data.cells[rowIdx][data.CurrentCol()]
						if cell.MaxWidth == 1 {
							break
						}
						cell.SetMaxWidth(cell.MaxWidth - 1)
					}
				case 'f': // Sort string values asc.
					data.SortColStrAsc(data.CurrentCol())
				case 'F': // Sort string values desc.
					data.SortColStrDesc(data.CurrentCol())
				case 'o': // Insert row below.
					history.Do(NewInsertRowBelowCommand(data, data.CurrentRow()))
				case 'O': // Insert row above.
					history.Do(NewInsertRowAboveCommand(table, data, data.CurrentRow()))
				case 'i':
					history.Do(NewInsertColRightCommand(data, data.CurrentCol()))
				case 'I':
					data.InsertColumn(data.CurrentCol())
					data.SetCurrentCol(data.CurrentCol() + 1)
					table.Select(data.CurrentRow(), data.CurrentCol())
				case 'u':
					history.Undo()
				case 18:
					if key == 18 { // CTRL+R, redo.
						history.Redo()
					}
				}
				return event
			},
		)
}

func buildCellInput() {
	cellInput.
		SetLabel(fmt.Sprintf("%d:%d ", data.CurrentRow()-1, data.CurrentCol()-1)).
		SetText(data.GetCurrentCell().Text).
		SetDoneFunc(func(key tcell.Key) {
			data.GetCurrentCell().SetText(cellInput.GetText())
			app.SetFocus(table)
		})
}

var csvFile *string
var data = NewData()
var table = tview.NewTable()
var app = tview.NewApplication()
var cellInput = tview.NewInputField()
var pages = tview.NewPages()
var modalContents = tview.NewBox()
var bottomBar = tview.NewTextView()
var history = NewHistory()

func main() {
	// Parse cli arguments.
	csvFile = flag.String("file", "", "path to csv file")
	flag.Parse()
	if *csvFile == "" {
		log.Fatal("-file not specified")
	}

	// Load csv file data.
	readCsvFile(*csvFile, data)

	data.SetCurrentRow(1)
	data.SetCurrentCol(1)

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
				data.AddEmptyRow()
			case 'C':
				data.AddEmptyColumn()
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

// Undo / redo functionality.
type Command interface {
	Execute()
	Unexecute()
}

type History struct {
	UndoStack []Command
	RedoStack []Command
}

func NewHistory() *History {
	return &History{}
}
func (h *History) Do(cmd Command) {
	cmd.Execute()
	h.UndoStack = append(h.UndoStack, cmd)
	// Clear RedoStack because a new action has been taken
	h.RedoStack = nil
}

func (h *History) Undo() {
	if len(h.UndoStack) == 0 {
		return
	}
	// Pop command from UndoStack and reverse the action
	last := len(h.UndoStack) - 1
	cmd := h.UndoStack[last]
	cmd.Unexecute()
	h.UndoStack = h.UndoStack[:last]
	// Push the command onto RedoStack
	h.RedoStack = append(h.RedoStack, cmd)
}

func (h *History) Redo() {
	if len(h.RedoStack) == 0 {
		return
	}
	// Pop command from RedoStack and re-apply the action
	last := len(h.RedoStack) - 1
	cmd := h.RedoStack[last]
	cmd.Execute()
	h.RedoStack = h.RedoStack[:last]
	// Push the command back onto UndoStack
	h.UndoStack = append(h.UndoStack, cmd)
}

// InsertRowBelowCommand.
type InsertRowBelowCommand struct {
	data *Data
	row  int
}

func NewInsertRowBelowCommand(data *Data, row int) *InsertRowBelowCommand {
	return &InsertRowBelowCommand{data: data, row: row + 1}
}
func (cmd *InsertRowBelowCommand) Execute() {
	cmd.data.InsertRow(cmd.row)
}

func (cmd *InsertRowBelowCommand) Unexecute() {
	cmd.data.RemoveRow(cmd.row)
}

// InsertRowAboveCommand.
type InsertRowAboveCommand struct {
	data  *Data
	Row   int
	table *tview.Table
}

func NewInsertRowAboveCommand(table *tview.Table, data *Data, row int) *InsertRowAboveCommand {
	return &InsertRowAboveCommand{table: table, data: data, Row: row}
}
func (cmd *InsertRowAboveCommand) Execute() {
	cmd.data.InsertRow(cmd.Row)
	cmd.data.SetCurrentRow(cmd.Row + 1)
	cmd.table.Select(data.CurrentRow(), data.CurrentCol())
}

func (cmd *InsertRowAboveCommand) Unexecute() {
	cmd.data.RemoveRow(cmd.Row)
}

// InsertRowAboveCommand.
type InsertColRightCommand struct {
	data  *Data
	col   int
}

func NewInsertColRightCommand(data *Data, col int) *InsertColRightCommand {
	return &InsertColRightCommand{data: data, col: col+1}
}

func (cmd *InsertColRightCommand) Execute() {
	cmd.data.InsertColumn(cmd.col )
}

func (cmd *InsertColRightCommand) Unexecute() {
	cmd.data.RemoveColumn(cmd.col)
}
