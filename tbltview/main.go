// This is simple tui spreadsheet program with vim keybindings.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"tblview/data"
	formulas "tblview/forumulas"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var csvFile *string
var dta *data.Data
var table = tview.NewTable()
var app = tview.NewApplication()
var cellInput = tview.NewInputField()
var pages = tview.NewPages()
var modal func(p tview.Primitive, width, height int) tview.Primitive
var modalContents = tview.NewBox()
var bottomBar = tview.NewTextView()
var history = NewHistory()
var logger = NewLogger("tmp/log.txt")

func main() {

	// Parse cli arguments.
	csvFile = flag.String("file", "", "path to csv file")
	flag.Parse()
	if *csvFile == "" {
		log.Fatal("-file not specified")
	}

	// Configure available frmls.
	var frmls []data.Formula
	frmls = append(frmls, formulas.NewSumFormula())

	// Init Data.
	dta = data.NewData(frmls)

	// Load csv file data.
	readCsvFile(*csvFile, dta)

	// Set cursor to the first cell.
	dta.SetCurrentRow(1)
	dta.SetCurrentCol(1)

	buildCellInput()
	buildTable()
	buildModal()

	// Configure layout.
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cellInput, 1, 0, false).
		AddItem(table, 0, 1, false).
		AddItem(bottomBar, 1, 0, false)

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

func buildTable() {
	table.
		SetBorders(false).
		SetContent(dta).
		SetSelectable(true, true).
		SetFixed(2, 1).
		Select(1, 1).
		SetSelectedFunc(func(row, col int) {
			logger.Info(fmt.Sprintf("table.SetSelectedFunc: row %d, col %d", row, col))
			app.SetFocus(cellInput)
		}).
		SetSelectionChangedFunc(
			func(row, col int) {
				logger.Info(fmt.Sprintf("table.SetSelectionChangedFunc: row %d, col %d", row, col))
				// Don't select x,y coordinates.
				if row == 0 {
					dta.SetCurrentRow(1)
					table.Select(dta.CurrentRow(), col)
					return
				}
				if col == 0 {
					dta.SetCurrentCol(1)
					table.Select(row, dta.CurrentCol())
					return
				}

				// Select individual cell.
				dta.SetCurrentRow(row) // account for top coordinate row
				dta.SetCurrentCol(col) // account for leftmost coordinates col

				cellInput.SetLabel(fmt.Sprintf("%d:%d ", row-1, col-1))
				cellInput.SetText(dta.GetCurrentCell().GetText())

				dta.SetHighlight(dta.GetCurrentCell().Calculate())
			}).
		SetInputCapture(
			func(event *tcell.EventKey) *tcell.EventKey {
				logger.Info(fmt.Sprintf("table.SetInputCapture: rune - %v, key - %v, modifier - %v, name - %v", event.Rune(), event.Key(), event.Modifiers(), event.Name()))
				row, col := table.GetSelection()
				rowSelectable, colSelectable := table.GetSelectable()
				rowSelected := rowSelectable && !colSelectable
				colSelected := !rowSelectable && colSelectable

				rune := event.Rune()
				key := event.Key()

				switch rune {
				case 'i':
					app.SetFocus(cellInput)
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
					history.Do(NewIncreaseColWidthCommand(dta.CurrentCol()))
				case '<': // Decrease column width.
					history.Do(NewDecreaseColWidthCommand(dta.CurrentCol()))
				case 'f': // Sort string values asc.
					history.Do(NewSortColStrAscCommand(dta.CurrentCol()))
				case 'F': // Sort string values desc.
					history.Do(NewSortColStrDescCommand(dta.CurrentCol()))
				case 'o': // Insert row below.
					history.Do(NewInsertRowBelowCommand(dta.CurrentRow()))
				case 'O': // Insert row above.
					history.Do(NewInsertRowAboveCommand(dta.CurrentRow(), dta.CurrentCol()))
				case 'a':
					history.Do(NewInsertColRightCommand(dta.CurrentCol()))
				case 'I':
					history.Do(NewInsertColLeftCommand(dta.CurrentRow(), dta.CurrentCol()))
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
		SetLabel(fmt.Sprintf("%d:%d ", dta.CurrentRow()-1, dta.CurrentCol()-1)).
		SetText(dta.GetCurrentCell().GetText()).
		SetDoneFunc(func(key tcell.Key) {
			app.SetFocus(table)
			// Push cursor down, if possible.
			if dta.CurrentRow() < dta.GetRowCount()-1 {
				dta.SetCurrentRow(dta.CurrentRow() + 1)
			}
			table.Select(dta.CurrentRow(), dta.CurrentCol())
		}).
		SetChangedFunc(func(text string) {
			// This function is called whenever cursor changes position for some reason.
			// So we need to check if the value actually changed.
			prevVal := dta.GetCurrentCell().GetText()
			if prevVal != text {
				history.Do(NewChangeCellValueCommand(dta.CurrentRow(), dta.CurrentCol(), text))
			}
			dta.SetHighlight(dta.GetCurrentCell().Calculate())
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

func readCsvFile(fileName string, dataTbl *data.Data) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error opening file: %s", err.Error())
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
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
	dataTbl.GetCell(0, 0).SetText("")

	// Pretty-print table header.
	for _, headerCell := range dataTbl.GetRow(1) {
		headerCell.SetAttributes(tcell.AttrBold)
	}
}

func addRecordToDataTable(recordCount int, record []string, dataTbl *data.Data) {
	var dataRow []*data.Cell

	// Set col header.
	colHead := data.NewCell()
	colHead.SetText(fmt.Sprintf("%d", recordCount))
	dataRow = append(dataRow, colHead)

	// Add row (record) data.
	for _, val := range record {
		cell := data.NewCell()
		cell.SetText(val)
		cell.SetMaxWidth(10)
		dataRow = append(dataRow, cell)
	}

	dataTbl.AddDataRow(dataRow)
}

func convertDataToArr(dataTbl *data.Data) [][]string {
	var data [][]string
	for _, row := range dataTbl.GetCells()[1:] { // account for top col numbers row
		row = row[1:] // account for row numbers col
		stringRow := make([]string, len(row))
		for j, cell := range row {
			stringRow[j] = cell.GetText()
		}
		data = append(data, stringRow)
	}
	return data
}

func saveDataToFile(path string, dataDataTable *data.Data) {
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
	dta.InsertRow(cmd.row)
	logger.Info(fmt.Sprintf("inserted row %d below", cmd.row))
}

func (cmd *InsertRowBelowCommand) Unexecute() {
	dta.RemoveRow(cmd.row)
	logger.Info(fmt.Sprintf("undo inserted row %d below", cmd.row))
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
	dta.InsertRow(cmd.row)
	dta.SetCurrentRow(cmd.row + 1)
	table.Select(cmd.row+1, cmd.col)
	logger.Info(fmt.Sprintf("inserted row %d above", cmd.row))
}

func (cmd *InsertRowAboveCommand) Unexecute() {
	dta.RemoveRow(cmd.row)
	dta.SetCurrentRow(cmd.row)
	table.Select(cmd.row, cmd.col)
	logger.Info(fmt.Sprintf("undo inserted row %d above", cmd.row))
}

// InsertColRightCommand.
type InsertColRightCommand struct {
	col int
}

func NewInsertColRightCommand(col int) *InsertColRightCommand {
	return &InsertColRightCommand{col: col + 1}
}

func (cmd *InsertColRightCommand) Execute() {
	dta.InsertColumn(cmd.col)
	logger.Info(fmt.Sprintf("inserted col %d right", cmd.col))
}

func (cmd *InsertColRightCommand) Unexecute() {
	dta.RemoveColumn(cmd.col)
	logger.Info(fmt.Sprintf("undo inserted col %d right", cmd.col))
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
	dta.InsertColumn(cmd.col)
	dta.SetCurrentCol(cmd.col + 1)
	table.Select(cmd.row, cmd.col+1)
	logger.Info(fmt.Sprintf("inserted col %d left", cmd.col))
}

func (cmd *InsertColLeftCommand) Unexecute() {
	dta.RemoveColumn(cmd.col)
	dta.SetCurrentCol(cmd.col)
	table.Select(cmd.row, cmd.col)
	logger.Info(fmt.Sprintf("undo inserted col %d left", cmd.col))
}

// SortColStrDescCommand is the command used to sort a column in descending string order.
type SortColStrDescCommand struct {
	col               int
	originalOrder     [][]*data.Cell // to remember the order before sorting
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
		cmd.originalSortedCol = dta.SortedCol()
		cmd.originalSortOrder = dta.SortOrder()
		// Capture the current order before sorting
		cmd.originalOrder = make([][]*data.Cell, dta.GetRowCount())
		for i, row := range dta.GetCells() {
			cmd.originalOrder[i] = make([]*data.Cell, len(row))
			copy(cmd.originalOrder[i], row)
		}
	}

	// Now sort the column in descending order
	dta.SortColStrDesc(cmd.col)
	logger.Info(fmt.Sprintf("sorted %d col by string desc", cmd.col))
}

// Unexecute restores the column to the original order before sorting.
func (cmd *SortColStrDescCommand) Unexecute() {
	if cmd.originalOrder != nil {
		// Restore the original cell order
		for i, row := range cmd.originalOrder {
			for j, cell := range row {
				dta.SetDataCell(i, j, cell)
			}
		}
	}
	dta.SetSortedCol(cmd.originalSortedCol)
	dta.SetSortOrder(cmd.originalSortOrder)
	dta.DrawXYCoordinates()
	logger.Info(fmt.Sprintf("undo sorted %d col by string desc", cmd.col))
}

// SortColStrAscCommand is the command used to sort a column in ascending string order.
type SortColStrAscCommand struct {
	col               int
	originalOrder     [][]*data.Cell // to remember the order before sorting
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
		cmd.originalSortedCol = dta.SortedCol()
		cmd.originalSortOrder = dta.SortOrder()
		// Capture the current order before sorting
		cmd.originalOrder = make([][]*data.Cell, dta.GetRowCount())
		for i, row := range dta.GetCells() {
			cmd.originalOrder[i] = make([]*data.Cell, len(row))
			copy(cmd.originalOrder[i], row)
		}
	}

	// Now sort the column in ascending order
	dta.SortColStrAsc(cmd.col)
	logger.Info(fmt.Sprintf("sorted %d col by string asc", cmd.col))
}

// Unexecute restores the column to the original order before sorting.
func (cmd *SortColStrAscCommand) Unexecute() {
	if cmd.originalOrder != nil {
		// Restore the original cell order
		for i, row := range cmd.originalOrder {
			for j, cell := range row {
				dta.SetDataCell(i, j, cell)
			}
		}
	}
	dta.SetSortedCol(cmd.originalSortedCol)
	dta.SetSortOrder(cmd.originalSortOrder)
	dta.DrawXYCoordinates()
	logger.Info(fmt.Sprintf("undo sorted %d col by string asc", cmd.col))
}

type DecreaseColWidthCommand struct {
	col int
}

func NewDecreaseColWidthCommand(col int) *DecreaseColWidthCommand {
	return &DecreaseColWidthCommand{col: col}
}

func (cmd *DecreaseColWidthCommand) Execute() {
	for rowIdx := range dta.GetCells() {
		cell := dta.GetDataCell(rowIdx, cmd.col)
		if cell.MaxWidth == 1 {
			break
		}
		cell.SetMaxWidth(cell.MaxWidth - 1)
	}
	logger.Info(fmt.Sprintf("decreased column %d width", dta.CurrentCol()))
}

func (cmd *DecreaseColWidthCommand) Unexecute() {
	for rowIdx := range dta.GetCells() {
		cell := dta.GetCell(rowIdx, dta.CurrentCol())
		cell.SetMaxWidth(cell.MaxWidth + 1)
	}
	logger.Info(fmt.Sprintf("undo decreased column %d width", dta.CurrentCol()))
}

type IncreaseColWidthCommand struct {
	col int
}

func NewIncreaseColWidthCommand(col int) *IncreaseColWidthCommand {
	return &IncreaseColWidthCommand{col: col}
}

func (cmd *IncreaseColWidthCommand) Execute() {
	for rowIdx := range dta.GetCells() {
		cell := dta.GetCell(rowIdx, dta.CurrentCol())
		cell.SetMaxWidth(cell.MaxWidth + 1)
	}
	logger.Info(fmt.Sprintf("increased column %d width", dta.CurrentCol()))
}

func (cmd *IncreaseColWidthCommand) Unexecute() {
	for rowIdx := range dta.GetCells() {
		cell := dta.GetCell(rowIdx, cmd.col)
		if cell.MaxWidth == 1 {
			break
		}
		cell.SetMaxWidth(cell.MaxWidth - 1)
	}
	logger.Info(fmt.Sprintf("undo increased column %d width", dta.CurrentCol()))
}

type DeleteColumnCommand struct {
	deletedCol []*data.Cell // to remember the order before sorting
	row        int
	col        int
}

func NewDeleteColumnCommand(row int, col int) *DeleteColumnCommand {
	return &DeleteColumnCommand{row: row, col: col}
}

func (cmd *DeleteColumnCommand) Execute() {
	// Capture the current column before deleting.
	if cmd.deletedCol == nil {
		for i := 0; i < dta.GetRowCount(); i++ {
			cmd.deletedCol = append(cmd.deletedCol, dta.GetDataCell(i, cmd.col))
		}
	}
	dta.RemoveColumn(cmd.col)
	if cmd.col == dta.GetColumnCount() { // Last column deleted, shift selection left.
		if dta.GetColumnCount() > 0 {
			table.Select(cmd.row, dta.GetColumnCount()-1)
		}
	}
	logger.Info(fmt.Sprintf("deleted column %d", cmd.col))
}

func (cmd *DeleteColumnCommand) Unexecute() {
	// This is last column (special case)
	if cmd.col == dta.GetColumnCount() {
		dta.InsertColumn(dta.GetColumnCount() - 1)
		// Paste back deleted cells.
		for row := range dta.GetCells() {
			dta.SetDataCell(row, cmd.col, cmd.deletedCol[row])
		}
		table.Select(cmd.row, dta.GetColumnCount()-1)
		return
	}

	dta.InsertColumn(cmd.col)
	// Paste back deleted cells.
	for row := range dta.GetCells() {
		dta.SetDataCell(row, cmd.col, cmd.deletedCol[row])
	}
	logger.Info(fmt.Sprintf("undo deleted column %d", cmd.col))
}

type DeleteRowCommand struct {
	deletedRow []*data.Cell // to remember the order before sorting
	row        int
	col        int
}

func NewDeleteRowCommand(row int, col int) *DeleteRowCommand {
	return &DeleteRowCommand{row: row, col: col}
}

func (cmd *DeleteRowCommand) Execute() {
	// Capture the current row before deleting.
	if cmd.deletedRow == nil {
		for i := 0; i < dta.GetColumnCount(); i++ {
			cmd.deletedRow = append(cmd.deletedRow, dta.GetDataCell(cmd.row, i))
		}
	}

	dta.RemoveRow(cmd.row)

	if cmd.row == dta.GetRowCount() { // Last row deleted, shift selection up.
		if dta.GetRowCount() > 0 {
			table.Select(dta.GetRowCount()-1, cmd.col)
		}
	}

	logger.Info(fmt.Sprintf("deleted row %d", cmd.row))
}

func (cmd *DeleteRowCommand) Unexecute() {
	// This is last column (special case)
	if cmd.row == dta.GetRowCount() {
		dta.InsertRow(dta.GetRowCount() - 1)
		// Paste back deleted cells.
		for col := 0; col < dta.GetColumnCount(); col++ {
			dta.SetDataCell(dta.GetRowCount()-1, col, cmd.deletedRow[col])
		}
		table.Select(dta.GetRowCount()-1, cmd.col)
		return
	}

	dta.InsertRow(cmd.row)
	// Paste back deleted cells.
	for col := 0; col < dta.GetColumnCount(); col++ {
		dta.SetDataCell(cmd.row, col, cmd.deletedRow[col])
	}

	logger.Info(fmt.Sprintf("undo deleted row %d", cmd.row))
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
	cmd.prevVal = dta.GetDataCell(cmd.row, cmd.col).GetText()
	dta.GetDataCell(cmd.row, cmd.col).SetText(cmd.newVal)
	logger.Info(fmt.Sprintf("%d:%d changed value from %s to %s", cmd.row, cmd.col, cmd.prevVal, cmd.newVal))
}

func (cmd *ChangeCellValueCommand) Unexecute() {
	dta.GetDataCell(cmd.row, cmd.col).SetText(cmd.prevVal)
	logger.Info(fmt.Sprintf("%d:%d undo value from %s to %s", cmd.row, cmd.col, cmd.newVal, cmd.prevVal))
}

type Logger struct {
	file *os.File
}

func NewLogger(path string) *Logger {
	logger := &Logger{}
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	logger.file = file
	return logger
}

func (l *Logger) Info(msg string) {
	msg = msg + "\n"
	_, err := l.file.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
}
