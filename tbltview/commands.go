package main

import (
	"fmt"
	"github.com/qdm12/reprint"
	"strings"
	"tblview/data"
)

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

// TODO: log history actions
func (h *History) Do(cmd Command) {
	// Execute command and que table update.
	go func() {
		app.QueueUpdateDraw(
			func() {
				cmd.Execute()
			},
		)
	}()
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
	// Execute command and que table update.
	go func() {
		app.QueueUpdateDraw(
			func() {
				cmd.Unexecute()
			},
		)
	}()
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
	// Execute command and que table update.
	go func() {
		app.QueueUpdateDraw(
			func() {
				cmd.Execute()
			},
		)
	}()
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

type DeleteRowsCommand struct {
	cellSnapshot  [][]*data.Cell
	selectedCells data.Selection
}

func NewDeleteRowsCommand(s data.Selection) *DeleteRowsCommand {
	return &DeleteRowsCommand{selectedCells: s}
}

func (cmd *DeleteRowsCommand) Execute() {
	cmd.cellSnapshot = dta.SnapShotCells()

	dta.RemoveRows(cmd.selectedCells.GetTopRow(), cmd.selectedCells.GetBottomRow())

	logger.Info(fmt.Sprintf("DeleteRowCommand: delete from row %d to row %d", cmd.selectedCells.GetTopRow(), cmd.selectedCells.GetBottomRow()))
}

func (cmd *DeleteRowsCommand) Unexecute() {
	dta.RestoreSnapshot(cmd.cellSnapshot)
	logger.Info(fmt.Sprintf("undo deleted rows %v", cmd.selectedCells))
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

type ReplaceTextCommand struct {
	selection     *data.Selection
	search        string
	replace       string
	originalCells [][]*data.Cell
}

func NewReplaceTextCommand(selection *data.Selection, search string, replace string) *ReplaceTextCommand {
	return &ReplaceTextCommand{selection: selection, search: search, replace: replace}
}

func (cmd *ReplaceTextCommand) Execute() {
	// Clear cell selection first.
	dta.ClearSelection(cmd.selection)
	selection.Clear()

	// Mirror originalCells dimensions to data.
	cmd.originalCells = make([][]*data.Cell, dta.GetRowCount())
	for row := range cmd.originalCells {
		cmd.originalCells[row] = make([]*data.Cell, dta.GetColumnCount())
	}

	var replaced bool
	for row := cmd.selection.GetTopRow(); row <= cmd.selection.GetBottomRow(); row++ {
		for col := cmd.selection.GetLeftCol(); col <= cmd.selection.GetRightCol(); col++ {
			cell := dta.GetDataCell(row, col)
			// (Deep) copy cell (even unexported fields).
			err := reprint.FromTo(&cell, &cmd.originalCells[row][col])
			if err != nil {
				logger.Error(err.Error())
			}
			// Replace cell text.
			newText := strings.ReplaceAll(cell.GetText(), cmd.search, cmd.replace)
			cell.SetText(newText)
			replaced = true

			logger.Info(fmt.Sprintf("cell %d:%d replaced %s with %s", row, col, cmd.search, cmd.replace))
		}
	}

	if !replaced {
		logger.Info(fmt.Sprintf("did not replace %s with %s in selection %v", cmd.search, cmd.replace, cmd.selection))
		return
	}
}

func (cmd *ReplaceTextCommand) Unexecute() {
	// Restore copied cells.
	for row := cmd.selection.GetTopRow(); row <= cmd.selection.GetBottomRow(); row++ {
		for col := cmd.selection.GetLeftCol(); col <= cmd.selection.GetRightCol(); col++ {
			cell := cmd.originalCells[row][col]
			dta.SetDataCell(row, col, cell)
		}
	}
	logger.Info(fmt.Sprintf("undo replace %s with %s in selection %v", cmd.search, cmd.replace, cmd.selection))
}
