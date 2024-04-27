// This is simple tui spreadsheet program with vim keybindings.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	ascIndicator  = "↑"
	descIndicator = "↓"
)

type Cell struct {
	*tview.TableCell
	text string
}

func NewCell() *Cell {
	cell := &Cell{TableCell: tview.NewTableCell("")}
	cell.SetMaxWidth(10)
	return cell
}
func (cell *Cell) SetText(text string) {
	cell.text = strings.ReplaceAll(text, " ", "") // remove spaces
	cell.Calculate()
}
func (cell *Cell) ShowError(text string) {
	cell.TableCell.SetText("#ERR:" + text)
	cell.TableCell.SetTextColor(tcell.ColorRed)
}
func (cell *Cell) Calculate() *highlight {
	if cell.IsFormula() {
		fText := cell.text[1:] // strip leading =
		for _, formula := range formulas {
			isMatch, _ := formula.Match(fText)
			if isMatch {
				calculated, highlight, err := formula.Calculate(fText)
				if err != nil {
					cell.ShowError(err.Error())
					return nil
				}
				cell.TableCell.SetText(calculated)
				cell.SetTextColor(tcell.ColorGreen)
				return highlight
			}
		}
		cell.ShowError("no formula")
		return nil
	}
	cell.TableCell.SetText(cell.text)
	cell.TableCell.SetTextColor(tcell.ColorWhite)
	return nil
}

func (cell *Cell) GetText() string {
	return cell.text
}

func (cell *Cell) IsFormula() bool {
	return strings.HasPrefix(cell.text, "=")
}

type Formula interface {
	// Checks if provided text matches the formula.
	Match(text string) (ok bool, matches []string)
	// Calculates the formula.
	Calculate(text string) (string, *highlight, error)
}

type SumFormula struct{}

func NewSumFormula() *SumFormula {
	return &SumFormula{}
}
func (f *SumFormula) Match(text string) (ok bool, matches []string) {
	pattern := `^SUM\((\d+),(\d+);(\d+),(\d+)\)$`
	re := regexp.MustCompile(pattern)
	matches = re.FindStringSubmatch(text)
	return matches != nil, matches
}

func (f *SumFormula) Calculate(text string) (string, *highlight, error) {
	ok, matches := f.Match(text)
	if !ok {
		return "", nil, fmt.Errorf("string does not match formula")
	}

	// Assuming matches[1:] are {startRow, startY, endX, endY}
	startRow, _ := strconv.Atoi(matches[1])
	startCol, _ := strconv.Atoi(matches[2])
	endRow, _ := strconv.Atoi(matches[3])
	endCol, _ := strconv.Atoi(matches[4])

	// Call the sum method (assuming data is accessible)
	total, err := f.sum(startRow+1, startCol+1, endRow+1, endCol+1)
	if err != nil {
		return "", nil, err
	}

	highlight := NewHighlight()
	highlight.startRow = startRow
	highlight.startCol = startCol
	highlight.endRow = endRow
	highlight.endCol = endCol

	return fmt.Sprintf(floatFormat, total), highlight, nil
}

func (f *SumFormula) sum(startRow, startCol, endRow, endCol int) (float64, error) {
	sum := 0.0

	// Validate the coordinates
	if startCol > endCol || startRow > endRow {
		return 0, fmt.Errorf("start coordinates must be less than or equal to end coordinates")
	}
	if startCol < 0 || startRow < 0 || endRow >= len(data.cells) || endCol >=
		len(data.cells[0]) {
		return 0, fmt.Errorf("coordinates out of bounds")
	}

	// Sum cells in the range [startX:endX, startY:endY]
	for y := startRow; y <= endRow; y++ {
		for x := startCol; x <= endCol; x++ {
			val, err := strconv.ParseFloat(data.cells[y][x].TableCell.Text, 64)
			if err != nil {
				return 0, fmt.Errorf("%d,%d is not an integer", y-1, x-1)
			}
			sum += val
		}
	}
	return sum, nil
}

// Highlighted cells region.
type highlight struct {
	startRow int
	startCol int
	endRow   int
	endCol   int
}

func NewHighlight() *highlight {
	return &highlight{}
}

func (h *highlight) IsHighlighted() bool {
	if h.startCol == 0 && h.startRow == 0 {
		return false
	}
	return true
}

func (h *highlight) Clear() {
	h.startRow = 0
	h.startCol = 0
	h.endRow = 0
	h.endCol = 0
}

// Data type.
type Data struct {
	cells      [][]*Cell
	currentRow int
	currentCol int
	sortedCol  int
	sortOrder  string
	highlight  *highlight
}

func NewData() *Data {
	return &Data{sortedCol: -1, sortOrder: "", highlight: NewHighlight()}
}
func (t *Data) Clear() {
	t.cells = nil
}
func (d *Data) InsertColumn(column int) {
	for row := range d.cells {
		if column > len(d.cells[row]) {
			continue
		}
		d.cells[row] = append(d.cells[row], nil)             // Extend by one.
		copy(d.cells[row][column+1:], d.cells[row][column:]) // Shift to the right.
		d.cells[row][column] = NewCell()
		d.drawXYCoordinates()
	}
}
func (d *Data) InsertRow(row int) {
	if row > d.GetRowCount() {
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
func (d *Data) AddDataRow(dataRow []*Cell) {
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
			return cell.TableCell
		}
		cell.SetAttributes(tcell.AttrDim)
		cell.SetAlign(1) //AlignCenter

		// Highlight row header cell for current selection.
		if column == d.currentCol {
			cell.SetAttributes(tcell.AttrBold)
			cell.SetAttributes(tcell.AttrUnderline)
			return cell.TableCell
		}
		return cell.TableCell
	}

	if column == 0 { // This is leftmost row with row numbers.
		cell.SetAttributes(tcell.AttrDim)

		// Highlight col header cell for current selection.
		if row == d.currentRow {
			cell.SetAttributes(tcell.AttrBold)
			cell.SetAttributes(tcell.AttrUnderline)
			return cell.TableCell
		}
		return cell.TableCell
	}

	cell.Calculate()

	// Highlight cell if needed.
	// TODO: refactor
	if d.highlight != nil && d.highlight.IsHighlighted() {
		if row >= d.highlight.startRow+1 && column >= d.highlight.startCol+1 && row <= d.highlight.endRow+1 && column <= d.highlight.endCol+1 {
			cell.SetTextColor(tcell.ColorGreen)
		}
	} else {
		cell.SetAttributes(tcell.AttrNone)
	}
	return cell.TableCell
}

func (d *Data) SetCell(row, column int, cell *tview.TableCell) {
	// Part of the tview.TableContent interface.
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

func (d *Data) createRow() []*Cell {
	var row []*Cell
	for i := 0; i < d.GetColumnCount(); i++ {
		row = append(row, NewCell())
	}
	return row
}

func (d *Data) GetCurrentCell() *Cell {
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
	d.sortColumn(col, func(a, b *Cell) bool {
		return a.TableCell.Text < b.TableCell.Text // Compare the text of the cells for ascending order.
	})
	d.sortedCol = col
	d.sortOrder = ascIndicator
	d.drawXYCoordinates()
}

func (d *Data) SortColStrDesc(col int) {
	d.sortColumn(col, func(a, b *Cell) bool {
		return a.TableCell.Text > b.TableCell.Text // Compare the text of the cells for descending order.
	})
	d.sortedCol = col
	d.sortOrder = descIndicator
	d.drawXYCoordinates()
}

// Sorts column. Accept column index and a sorter function that
// takes slice of vertical column cells as an argument.
func (d *Data) sortColumn(col int, sorter func(a, b *Cell) bool) {
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
		colText := fmt.Sprintf("%d", colIdx-1)
		if d.sortedCol != -1 {
			if colIdx == d.sortedCol {
				colText = colText + d.sortOrder
			}
		}
		col.SetText(colText)
	}

	d.cells[0][0].SetText("")
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
	var dataRow []*Cell

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
			stringRow[j] = cell.GetText()
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

				cellInput.SetLabel(fmt.Sprintf("%d:%d ", row-1, col-1))
				cellInput.SetText(data.GetCurrentCell().GetText())

				data.highlight = data.GetCurrentCell().Calculate()
			}).
		SetInputCapture(
			func(event *tcell.EventKey) *tcell.EventKey {
				//bottomBar.SetText(fmt.Sprintf("rune: %v, key: %v, modifier: %v, name: %v", event.Rune(), event.Key(), event.Modifiers(), event.Name()))
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
						history.Do(NewDeleteRowCommand(row, col))
					} else if colSelected {
						history.Do(NewDeleteColumnCommand(row, col))
					}
				case '>': // Increase column width.
					history.Do(NewIncreaseColWidthCommand(data.CurrentCol()))
				case '<': // Decrease column width.
					history.Do(NewDecreaseColWidthCommand(data.CurrentCol()))
				case 'f': // Sort string values asc.
					history.Do(NewSortColStrAscCommand(data.CurrentCol()))
				case 'F': // Sort string values desc.
					history.Do(NewSortColStrDescCommand(data.CurrentCol()))
				case 'o': // Insert row below.
					history.Do(NewInsertRowBelowCommand(data.CurrentRow()))
				case 'O': // Insert row above.
					history.Do(NewInsertRowAboveCommand(data.CurrentRow(), data.CurrentCol()))
				case 'i':
					history.Do(NewInsertColRightCommand(data.CurrentCol()))
				case 'I':
					history.Do(NewInsertColLeftCommand(data.CurrentRow(), data.CurrentCol()))
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
		SetText(data.GetCurrentCell().GetText()).
		SetDoneFunc(func(key tcell.Key) {
			app.SetFocus(table)
			// Push cursor down, if possible.
			if data.CurrentRow() < data.GetRowCount()-1 {
				data.SetCurrentRow(data.CurrentRow() + 1)
			}
			table.Select(data.CurrentRow(), data.CurrentCol())
		}).
		SetChangedFunc(func(text string) {
			history.Do(NewChangeCellValueCommand(data.CurrentRow(), data.CurrentCol(), text))
			data.highlight = data.GetCurrentCell().Calculate()
		},
		)
}
func buildModal() {
	// Returns a new primitive which puts the provided primitive in the center and
	// sets its size to the given width and height.
	modal = func(p tview.Primitive, width, height int) tview.Primitive {
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

}

var csvFile *string
var data = NewData()
var table = tview.NewTable()
var app = tview.NewApplication()
var cellInput = tview.NewInputField()
var pages = tview.NewPages()
var modal func(p tview.Primitive, width, height int) tview.Primitive
var modalContents = tview.NewBox()
var bottomBar = tview.NewTextView()
var history = NewHistory()
var formulas []Formula
var floatFormat = "%.2f"

func main() {
	// Parse cli arguments.
	csvFile = flag.String("file", "", "path to csv file")
	flag.Parse()
	if *csvFile == "" {
		log.Fatal("-file not specified")
	}

	// Load csv file data.
	readCsvFile(*csvFile, data)

	// Configure available formulas.
	formulas = append(formulas, NewSumFormula())

	// Set cursor to the first cell.
	data.SetCurrentRow(1)
	data.SetCurrentCol(1)

	buildCellInput()
	buildTableWidget()
	buildModal()

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
	row int
}

func NewInsertRowBelowCommand(row int) *InsertRowBelowCommand {
	return &InsertRowBelowCommand{row: row + 1}
}
func (cmd *InsertRowBelowCommand) Execute() {
	data.InsertRow(cmd.row)
}

func (cmd *InsertRowBelowCommand) Unexecute() {
	data.RemoveRow(cmd.row)
}

// InsertRowAboveCommand.
type InsertRowAboveCommand struct {
	row int
	col int
}

func NewInsertRowAboveCommand(row int, col int) *InsertRowAboveCommand {
	return &InsertRowAboveCommand{row: row, col: col}
}
func (cmd *InsertRowAboveCommand) Execute() {
	data.InsertRow(cmd.row)
	data.SetCurrentRow(cmd.row + 1)
	table.Select(cmd.row+1, cmd.col)
}

func (cmd *InsertRowAboveCommand) Unexecute() {
	data.RemoveRow(cmd.row)
	data.SetCurrentRow(cmd.row)
	table.Select(cmd.row, cmd.col)
}

// InsertColRightCommand.
type InsertColRightCommand struct {
	col int
}

func NewInsertColRightCommand(col int) *InsertColRightCommand {
	return &InsertColRightCommand{col: col + 1}
}

func (cmd *InsertColRightCommand) Execute() {
	data.InsertColumn(cmd.col)
}

func (cmd *InsertColRightCommand) Unexecute() {
	data.RemoveColumn(cmd.col)
}

// InsertColLeftCommand.
type InsertColLeftCommand struct {
	col int
	row int
}

func NewInsertColLeftCommand(row int, col int) *InsertColLeftCommand {
	return &InsertColLeftCommand{row: row, col: col}
}

func (cmd *InsertColLeftCommand) Execute() {
	data.InsertColumn(cmd.col)
	data.SetCurrentCol(cmd.col + 1)
	table.Select(cmd.row, cmd.col+1)
}

func (cmd *InsertColLeftCommand) Unexecute() {
	data.RemoveColumn(cmd.col)
	data.SetCurrentCol(cmd.col)
	table.Select(cmd.row, cmd.col)
}

// SortColStrDescCommand is the command used to sort a column in descending string order.
type SortColStrDescCommand struct {
	col               int
	originalOrder     [][]*Cell // to remember the order before sorting
	originalSortedCol int
	originalSortOrder string
}

// NewSortColStrDescCommand creates a new SortColStrDescCommand with the given column.
func NewSortColStrDescCommand(col int) *SortColStrDescCommand {
	return &SortColStrDescCommand{
		col:           col,
		originalOrder: nil, // will be set during the first execution
	}
}

// Execute executes the SortColStrDescCommand, sorting the column in descending order.
func (cmd *SortColStrDescCommand) Execute() {
	if cmd.originalOrder == nil {
		cmd.originalSortedCol = data.sortedCol
		cmd.originalSortOrder = data.sortOrder
		// Capture the current order before sorting
		cmd.originalOrder = make([][]*Cell, len(data.cells))
		for i, row := range data.cells {
			cmd.originalOrder[i] = make([]*Cell, len(row))
			copy(cmd.originalOrder[i], row)
		}
	}

	// Now sort the column in descending order
	data.SortColStrDesc(cmd.col)
}

// Unexecute restores the column to the original order before sorting.
func (cmd *SortColStrDescCommand) Unexecute() {
	if cmd.originalOrder != nil {
		// Restore the original cell order
		for i, row := range cmd.originalOrder {
			for j, cell := range row {
				data.cells[i][j] = cell
			}
		}
	}
	data.sortedCol = cmd.originalSortedCol
	data.sortOrder = cmd.originalSortOrder
	data.drawXYCoordinates()
}

// SortColStrAscCommand is the command used to sort a column in ascending string order.
type SortColStrAscCommand struct {
	col               int
	originalOrder     [][]*Cell // to remember the order before sorting
	originalSortedCol int
	originalSortOrder string
}

// NewSortColStrAscCommand creates a new SortColStrAscCommand with the given column.
func NewSortColStrAscCommand(col int) *SortColStrAscCommand {
	return &SortColStrAscCommand{
		col:           col,
		originalOrder: nil, // will be set during the first execution
	}
}

// Execute executes the SortColStrAscCommand, sorting the column in ascending order.
func (cmd *SortColStrAscCommand) Execute() {
	if cmd.originalOrder == nil {
		cmd.originalSortedCol = data.sortedCol
		cmd.originalSortOrder = data.sortOrder
		// Capture the current order before sorting
		cmd.originalOrder = make([][]*Cell, len(data.cells))
		for i, row := range data.cells {
			cmd.originalOrder[i] = make([]*Cell, len(row))
			copy(cmd.originalOrder[i], row)
		}
	}

	// Now sort the column in ascending order
	data.SortColStrAsc(cmd.col)
}

// Unexecute restores the column to the original order before sorting.
func (cmd *SortColStrAscCommand) Unexecute() {
	if cmd.originalOrder != nil {
		// Restore the original cell order
		for i, row := range cmd.originalOrder {
			for j, cell := range row {
				data.cells[i][j] = cell
			}
		}
	}
	data.sortedCol = cmd.originalSortedCol
	data.sortOrder = cmd.originalSortOrder
	data.drawXYCoordinates()
}

type DecreaseColWidthCommand struct {
	col int
}

func NewDecreaseColWidthCommand(col int) *DecreaseColWidthCommand {
	return &DecreaseColWidthCommand{col: col}
}

func (cmd *DecreaseColWidthCommand) Execute() {
	for rowIdx := range data.cells {
		cell := data.cells[rowIdx][cmd.col]
		if cell.MaxWidth == 1 {
			break
		}
		cell.SetMaxWidth(cell.MaxWidth - 1)
	}
}

func (cmd *DecreaseColWidthCommand) Unexecute() {
	for rowIdx := range data.cells {
		cell := data.cells[rowIdx][data.CurrentCol()]
		cell.SetMaxWidth(cell.MaxWidth + 1)
	}
}

type IncreaseColWidthCommand struct {
	col int
}

func NewIncreaseColWidthCommand(col int) *IncreaseColWidthCommand {
	return &IncreaseColWidthCommand{col: col}
}

func (cmd *IncreaseColWidthCommand) Execute() {
	for rowIdx := range data.cells {
		cell := data.cells[rowIdx][data.CurrentCol()]
		cell.SetMaxWidth(cell.MaxWidth + 1)
	}
}

func (cmd *IncreaseColWidthCommand) Unexecute() {
	for rowIdx := range data.cells {
		cell := data.cells[rowIdx][cmd.col]
		if cell.MaxWidth == 1 {
			break
		}
		cell.SetMaxWidth(cell.MaxWidth - 1)
	}
}

type DeleteColumnCommand struct {
	deletedCol []*Cell // to remember the order before sorting
	row        int
	col        int
}

func NewDeleteColumnCommand(row int, col int) *DeleteColumnCommand {
	return &DeleteColumnCommand{row: row, col: col}
}

func (cmd *DeleteColumnCommand) Execute() {
	// Capture the current column before deleting.
	if cmd.deletedCol == nil {
		for i := 0; i < data.GetRowCount(); i++ {
			cmd.deletedCol = append(cmd.deletedCol, data.cells[i][cmd.col])
		}
	}
	data.RemoveColumn(cmd.col)
	if cmd.col == data.GetColumnCount() { // Last column deleted, shift selection left.
		if data.GetColumnCount() > 0 {
			table.Select(cmd.row, data.GetColumnCount()-1)
		}
	}
}

func (cmd *DeleteColumnCommand) Unexecute() {
	// This is last column (special case)
	if cmd.col == data.GetColumnCount() {
		data.InsertColumn(data.GetColumnCount() - 1)
		// Paste back deleted cells.
		for row := range data.cells {
			data.cells[row][cmd.col] = cmd.deletedCol[row]
		}
		table.Select(cmd.row, data.GetColumnCount()-1)
		return
	}

	data.InsertColumn(cmd.col)
	// Paste back deleted cells.
	for row := range data.cells {
		data.cells[row][cmd.col] = cmd.deletedCol[row]
	}
}

type DeleteRowCommand struct {
	deletedRow []*Cell // to remember the order before sorting
	row        int
	col        int
}

func NewDeleteRowCommand(row int, col int) *DeleteRowCommand {
	return &DeleteRowCommand{row: row, col: col}
}

func (cmd *DeleteRowCommand) Execute() {
	// Capture the current row before deleting.
	if cmd.deletedRow == nil {
		for i := 0; i < data.GetColumnCount(); i++ {
			cmd.deletedRow = append(cmd.deletedRow, data.cells[cmd.row][i])
		}
	}

	data.RemoveRow(cmd.row)

	if cmd.row == data.GetRowCount() { // Last row deleted, shift selection up.
		if data.GetRowCount() > 0 {
			table.Select(data.GetRowCount()-1, cmd.col)
		}
	}
}

func (cmd *DeleteRowCommand) Unexecute() {
	// This is last column (special case)
	if cmd.row == data.GetRowCount() {
		data.InsertRow(data.GetRowCount() - 1)
		// Paste back deleted cells.
		for col := 0; col < data.GetColumnCount(); col++ {
			data.cells[data.GetRowCount()-1][col] = cmd.deletedRow[col]
		}
		table.Select(data.GetRowCount()-1, cmd.col)
		return
	}

	data.InsertRow(cmd.row)
	// Paste back deleted cells.
	for col := 0; col < data.GetColumnCount(); col++ {
		data.cells[cmd.row][col] = cmd.deletedRow[col]
	}
}

type ChangeCellValueCommand struct {
	row     int
	col     int
	prevVal string
	newVal  string
}

func NewChangeCellValueCommand(row int, col int, text string) *ChangeCellValueCommand {
	return &ChangeCellValueCommand{row: row, col: col, newVal: text}
}

func (cmd *ChangeCellValueCommand) Execute() {
	cmd.prevVal = data.cells[cmd.row][cmd.col].GetText()
	data.cells[cmd.row][cmd.col].SetText(cmd.newVal)
}

func (cmd *ChangeCellValueCommand) Unexecute() {
	data.cells[cmd.row][cmd.col].SetText(cmd.prevVal)
}
