package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"tblview/data"
)

// Undo / redo functionality.
type Command interface {
	Execute() error
	Unexecute() error
}

type History struct {
	UndoStack []Command
	RedoStack []Command
}

func NewHistory() *History {
	return &History{}
}

func (h *History) Do(cmd Command) {
	err := cmd.Execute()
	if err != nil {
		go func() {
			app.QueueUpdateDraw(func() {
				showModal(err.Error())
			})
		}()
	}
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
	err := cmd.Unexecute()
	if err != nil {
		go func() {
			app.QueueUpdateDraw(func() {
				showModal(err.Error())
			})
		}()
	}
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
	err := cmd.Execute()
	if err != nil {
		go func() {
			app.QueueUpdateDraw(func() {
				showModal(err.Error())
			})
		}()
	}
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
func (cmd *InsertRowBelowCommand) Execute() error {
	dta.InsertRow(cmd.row)
	logger.Info(fmt.Sprintf("inserted row %d below", cmd.row))
	return nil
}

func (cmd *InsertRowBelowCommand) Unexecute() error {
	dta.RemoveRow(cmd.row)
	logger.Info(fmt.Sprintf("undo inserted row %d below", cmd.row))
	return nil
}

// InsertRowAboveCommand.
type InsertRowAboveCommand struct {
	row int
	col int
}

func NewInsertRowAboveCommand(row int, col int) *InsertRowAboveCommand {
	return &InsertRowAboveCommand{row: row, col: col}
}
func (cmd *InsertRowAboveCommand) Execute() error {
	dta.InsertRow(cmd.row)
	dta.CurrentRow = cmd.row + 1
	table.Select(cmd.row+1, cmd.col)
	logger.Info(fmt.Sprintf("inserted row %d above", cmd.row))
	return nil
}

func (cmd *InsertRowAboveCommand) Unexecute() error {
	dta.RemoveRow(cmd.row)
	dta.CurrentRow = cmd.row
	table.Select(cmd.row, cmd.col)
	logger.Info(fmt.Sprintf("undo inserted row %d above", cmd.row))
	return nil
}

// InsertColRightCommand.
type InsertColRightCommand struct {
	col int
}

func NewInsertColRightCommand(col int) *InsertColRightCommand {
	return &InsertColRightCommand{col: col + 1}
}

func (cmd *InsertColRightCommand) Execute() error {
	dta.InsertColumn(cmd.col)
	logger.Info(fmt.Sprintf("inserted col %d right", cmd.col))
	return nil
}

func (cmd *InsertColRightCommand) Unexecute() error {
	dta.RemoveColumn(cmd.col)
	logger.Info(fmt.Sprintf("undo inserted col %d right", cmd.col))
	return nil
}

// InsertColLeftCommand.
type InsertColLeftCommand struct {
	col int
	row int
}

func NewInsertColLeftCommand(row int, col int) *InsertColLeftCommand {
	return &InsertColLeftCommand{row: row, col: col}
}

func (cmd *InsertColLeftCommand) Execute() error {
	dta.InsertColumn(cmd.col)
	dta.CurrentCol = cmd.col + 1
	table.Select(cmd.row, cmd.col+1)
	logger.Info(fmt.Sprintf("inserted col %d left", cmd.col))
	return nil
}

func (cmd *InsertColLeftCommand) Unexecute() error {
	dta.RemoveColumn(cmd.col)
	dta.CurrentCol = cmd.col
	table.Select(cmd.row, cmd.col)
	logger.Info(fmt.Sprintf("undo inserted col %d left", cmd.col))
	return nil
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
func (cmd *SortColStrDescCommand) Execute() error {
	if cmd.originalOrder == nil {
		cmd.originalSortedCol = dta.SortedCol
		cmd.originalSortOrder = dta.SortOrder
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
	return nil
}

// Unexecute restores the column to the original order before sorting.
func (cmd *SortColStrDescCommand) Unexecute() error {
	if cmd.originalOrder != nil {
		// Restore the original cell order
		for i, row := range cmd.originalOrder {
			for j, cell := range row {
				dta.SetDataCell(i, j, cell)
			}
		}
	}
	dta.SortedCol = cmd.originalSortedCol
	dta.SortOrder = cmd.originalSortOrder
	dta.DrawXYCoordinates()
	logger.Info(fmt.Sprintf("undo sorted %d col by string desc", cmd.col))
	return nil
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
func (cmd *SortColStrAscCommand) Execute() error {
	if cmd.originalOrder == nil {
		cmd.originalSortedCol = dta.SortedCol
		cmd.originalSortOrder = dta.SortOrder
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
	return nil
}

// Unexecute restores the column to the original order before sorting.
func (cmd *SortColStrAscCommand) Unexecute() error {
	if cmd.originalOrder != nil {
		// Restore the original cell order
		for i, row := range cmd.originalOrder {
			for j, cell := range row {
				dta.SetDataCell(i, j, cell)
			}
		}
	}
	dta.SortedCol = cmd.originalSortedCol
	dta.SortOrder = cmd.originalSortOrder
	dta.DrawXYCoordinates()
	logger.Info(fmt.Sprintf("undo sorted %d col by string asc", cmd.col))
	return nil
}

type DecreaseColWidthCommand struct {
	col int
}

func NewDecreaseColWidthCommand(col int) *DecreaseColWidthCommand {
	return &DecreaseColWidthCommand{col: col}
}

func (cmd *DecreaseColWidthCommand) Execute() error {
	for rowIdx := range dta.GetCells() {
		cell := dta.GetDataCell(rowIdx, cmd.col)
		if cell.Width == 1 {
			break
		}
		cell.Width = cell.Width - 1
		logger.Info(fmt.Sprintf("decreased column %d width to %d", dta.CurrentCol, cell.Width))
	}
	return nil
}

func (cmd *DecreaseColWidthCommand) Unexecute() error {
	for rowIdx := range dta.GetCells() {
		cell := dta.GetDataCell(rowIdx, cmd.col)
		cell.Width = cell.Width + 1
		logger.Info(fmt.Sprintf("increased column %d width to %d", dta.CurrentCol, cell.Width))
	}
	return nil
}

type IncreaseColWidthCommand struct {
	col int
}

func NewIncreaseColWidthCommand(col int) *IncreaseColWidthCommand {
	return &IncreaseColWidthCommand{col: col}
}

func (cmd *IncreaseColWidthCommand) Execute() error {
	for rowIdx := range dta.GetCells() {
		cell := dta.GetDataCell(rowIdx, cmd.col)
		cell.Width = cell.Width + 1
		logger.Info(fmt.Sprintf("increased column %d width to %d", dta.CurrentCol, cell.Width))
	}
	return nil
}

func (cmd *IncreaseColWidthCommand) Unexecute() error {
	for rowIdx := range dta.GetCells() {
		cell := dta.GetDataCell(rowIdx, cmd.col)
		if cell.Width == 1 {
			break
		}
		cell.Width = cell.Width - 1
		logger.Info(fmt.Sprintf("decreased column %d width to %d", dta.CurrentCol, cell.Width))
	}
	return nil
}

type DeleteColumnCommand struct {
	deletedCol []*data.Cell // to remember the order before sorting
	row        int
	col        int
}

func NewDeleteColumnCommand(row int, col int) *DeleteColumnCommand {
	return &DeleteColumnCommand{row: row, col: col}
}

func (cmd *DeleteColumnCommand) Execute() error {
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
	return nil
}

func (cmd *DeleteColumnCommand) Unexecute() error {
	// This is last column (special case)
	if cmd.col == dta.GetColumnCount() {
		dta.InsertColumn(dta.GetColumnCount() - 1)
		// Paste back deleted cells.
		for row := range dta.GetCells() {
			dta.SetDataCell(row, cmd.col, cmd.deletedCol[row])
		}
		table.Select(cmd.row, dta.GetColumnCount()-1)
		return nil
	}

	dta.InsertColumn(cmd.col)
	// Paste back deleted cells.
	for row := range dta.GetCells() {
		dta.SetDataCell(row, cmd.col, cmd.deletedCol[row])
	}
	logger.Info(fmt.Sprintf("undo deleted column %d", cmd.col))
	return nil
}

type DeleteRowsCommand struct {
	cellSnapshot  [][]data.Cell
	selectedCells data.Selection
}

func NewDeleteRowsCommand(s data.Selection) *DeleteRowsCommand {
	return &DeleteRowsCommand{selectedCells: s}
}

func (cmd *DeleteRowsCommand) Execute() error {
	cmd.cellSnapshot = dta.SnapShotCells()

	dta.RemoveRows(cmd.selectedCells.GetTopRow(), cmd.selectedCells.GetBottomRow())

	logger.Info(fmt.Sprintf("DeleteRowCommand: delete from row %d to row %d", cmd.selectedCells.GetTopRow(), cmd.selectedCells.GetBottomRow()))
	return nil
}

func (cmd *DeleteRowsCommand) Unexecute() error {
	dta.RestoreSnapshot(cmd.cellSnapshot)
	logger.Info(fmt.Sprintf("undo deleted rows %v", cmd.selectedCells))
	return nil
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

func (cmd *ChangeCellValueCommand) Execute() error {
	cmd.prevVal = dta.GetDataCell(cmd.row, cmd.col).Text
	dta.GetDataCell(cmd.row, cmd.col).Text = cmd.newVal
	logger.Info(fmt.Sprintf("%d:%d changed value from %s to %s", cmd.row, cmd.col, cmd.prevVal, cmd.newVal))
	return nil
}

func (cmd *ChangeCellValueCommand) Unexecute() error {
	dta.GetDataCell(cmd.row, cmd.col).Text = cmd.prevVal
	logger.Info(fmt.Sprintf("%d:%d undo value from %s to %s", cmd.row, cmd.col, cmd.newVal, cmd.prevVal))
	return nil
}

type ReplaceTextCommand struct {
	selection    *data.Selection
	search       string
	replace      string
	cellSnapshot [][]data.Cell
}

func NewReplaceTextCommand(selection *data.Selection, search string, replace string) *ReplaceTextCommand {
	return &ReplaceTextCommand{selection: selection, search: search, replace: replace}
}

func (cmd *ReplaceTextCommand) Execute() error {
	// Create cell snapshot.
	cmd.cellSnapshot = dta.SnapShotCells()

	var replaced bool
	for row := cmd.selection.GetTopRow(); row <= cmd.selection.GetBottomRow(); row++ {
		for col := cmd.selection.GetLeftCol(); col <= cmd.selection.GetRightCol(); col++ {
			cell := dta.GetDataCell(row, col)
			// Replace cell text.
			newText := strings.ReplaceAll(cell.Text, cmd.search, cmd.replace)
			cell.Text = newText
			replaced = true

			logger.Info(fmt.Sprintf("cell %d:%d replaced %s with %s", row, col, cmd.search, cmd.replace))
		}
	}

	if !replaced {
		logger.Error(fmt.Sprintf("did not replace %s with %s in selection %v", cmd.search, cmd.replace, cmd.selection))
		return fmt.Errorf("did not replace %s with %s in selection %v", cmd.search, cmd.replace, cmd.selection)
	}

	// Clear cell selection.
	dta.ClearSelection()
	selection.Clear()

	return nil
}

func (cmd *ReplaceTextCommand) Unexecute() error {
	// Restore copied cells.
	dta.RestoreSnapshot(cmd.cellSnapshot)
	logger.Info(fmt.Sprintf("undo replace %s with %s in selection %v", cmd.search, cmd.replace, cmd.selection))
	return nil
}

type WriteFileCommand struct {
	filePath string
}

func NewWriteFileCommand(filePath string) *WriteFileCommand {
	return &WriteFileCommand{filePath: filePath}
}

// TODO: remove panics
func (cmd *WriteFileCommand) Execute() error {
	file, err := os.Create(cmd.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	err = encoder.Encode(dta)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("wrote to file %s", cmd.filePath))
	return nil
}

func (cmd *WriteFileCommand) Unexecute() error {
	logger.Info(fmt.Sprintf("[NOT IMPLEMENTED] undo wrote to file %s", cmd.filePath))
	return nil
}

type LoadFileCommand struct {
	filePath string
}

func NewLoadFileCommand(filePath string) *LoadFileCommand {
	return &LoadFileCommand{filePath: filePath}
}

func (cmd *LoadFileCommand) Execute() error {
	file, err := os.Open(cmd.filePath)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(dta)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("loaded file %s", cmd.filePath))
	return nil
}

func (cmd *LoadFileCommand) Unexecute() error {
	logger.Info(fmt.Sprintf("[NOT IMPLEMENTED] undo loaded file %s", cmd.filePath))
	return nil
}
