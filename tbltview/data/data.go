package data

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"tblview/logger"

	"github.com/gdamore/tcell/v2"
	"github.com/jinzhu/copier"
	"github.com/rivo/tview"
)

var d *Data
var formulas []Formula

const (
	ascIndicator  = "↑"
	descIndicator = "↓"
)

type Cell struct {
	IsSelected bool
	IsInRange  bool // Cell is part of the formula range.
	Text       string

	// tview.TableCell attributes
	Attributes tcell.AttrMask
	Align      int
	Width      int
}

func NewCell() *Cell {
	cell := &Cell{Width: 10}
	return cell
}

func (cell *Cell) Calculate() (displayText string, rangeFormula *FormulaRange, err error) {
	if cell.IsFormula() {
		text := strings.ReplaceAll(cell.Text, " ", "") // remove spaces
		fText := text[1:]                              // strip leading =
		for _, formula := range formulas {
			isMatch, _ := formula.Match(fText)
			if isMatch {
				calculated, formulaRange, err := formula.Calculate(d, fText)
				if err != nil {
					return "", nil, err
				}
				return calculated, formulaRange, nil
			}
		}
		return "", nil, errors.New("no formula")
	}
	return cell.Text, nil, nil
}

func (cell *Cell) IsFormula() bool {
	return strings.HasPrefix(cell.Text, "=")
}

type Formula interface {
	// Checks if provided text matches the formula.
	Match(text string) (ok bool, matches []string)
	// Calculates the formula.
	Calculate(data *Data, text string) (string, *FormulaRange, error)
}

// TODO: use getter and setter and check validity in setter
type FormulaRange struct {
	StartRow int
	StartCol int
	EndRow   int
	EndCol   int
}

func NewFormulaRange() *FormulaRange {
	return &FormulaRange{}
}

type Selection struct {
	data *Data

	startRow int // Inital selction point.
	startCol int // Initial selection point.

	topRow    int
	leftCol   int
	bottomRow int
	rightCol  int
}

func NewSelection(data *Data) *Selection {
	return &Selection{
		data: data,
	}
}
func (s *Selection) Clear() {
	s.startRow = 0
	s.startCol = 0
	s.topRow = 0
	s.leftCol = 0
	s.bottomRow = 0
	s.rightCol = 0
}
func (s *Selection) SetCoordintates(startRow int, startCol int, endRow int, endCol int) {
	s.startRow = startRow
	s.startCol = startCol

	s.topRow = startRow
	s.leftCol = startCol
	s.bottomRow = endRow
	s.rightCol = endCol

}

func (s *Selection) IsRowSelected() bool {
	if s.leftCol == 1 && s.rightCol == s.data.GetColumnCount()-1 {
		return true
	}
	return false
}

func (s *Selection) IsColumnSelected() bool {
	if s.topRow == 1 && s.bottomRow == s.data.GetRowCount()-1 {
		return true
	}
	return false
}
func (s *Selection) Update(endRow int, endCol int) {
	s.bottomRow = endRow
	s.rightCol = endCol

	if s.bottomRow <= s.startRow {
		s.topRow = s.bottomRow
		s.bottomRow = s.startRow
	}

	if s.rightCol <= s.startCol {
		s.leftCol = s.rightCol
		s.rightCol = s.startCol
	}
}
func (s *Selection) GetTopRow() int {
	return s.topRow
}
func (s *Selection) GetLeftCol() int {
	return s.leftCol
}
func (s *Selection) GetBottomRow() int {
	return s.bottomRow
}
func (s *Selection) GetRightCol() int {
	return s.rightCol
}

// Data type.
type Data struct {
	cells      [][]*Cell
	currentRow int
	currentCol int
	sortedCol  int
	sortOrder  string
	logger     *logger.Logger
}

func NewData(frmls []Formula, logger *logger.Logger) *Data {
	formulas = frmls
	d = &Data{sortedCol: -1, sortOrder: "", logger: logger}
	return d
}
func (d *Data) GetCells() [][]*Cell {
	return d.cells
}
func (d *Data) GetDataCell(row int, col int) *Cell {
	return d.cells[row][col]
}
func (d *Data) SetDataCell(row int, col int, cell *Cell) {
	d.cells[row][col] = cell
}
func (d *Data) CopyDataCell(row int, col int, cell *Cell) {
	_ = copier.Copy(d.cells[row][col], cell)
}
func (d *Data) GetRow(row int) []*Cell {
	return d.cells[row]
}
func (d *Data) GetCol(column int) []*Cell {
	var col []*Cell
	for row := range d.cells {
		col = append(col, d.cells[row][column])
	}
	return col
}
func (d *Data) SortedCol() int {
	return d.sortedCol
}
func (d *Data) SetSortedCol(sortedCol int) {
	d.sortedCol = sortedCol
}
func (d *Data) SortOrder() string {
	return d.sortOrder
}
func (d *Data) SetSortOrder(sortOrder string) {
	d.sortOrder = sortOrder
}
func (d *Data) Clear() {
	d.cells = nil
}
func (d *Data) InsertColumn(column int) {
	for row := range d.cells {
		if column > len(d.cells[row]) {
			continue
		}
		d.cells[row] = append(d.cells[row], nil)             // Extend by one.
		copy(d.cells[row][column+1:], d.cells[row][column:]) // Shift to the right.
		d.cells[row][column] = NewCell()
		d.DrawXYCoordinates()
	}
}
func (d *Data) InsertRow(row int) {
	if row > d.GetRowCount() {
		return
	}

	d.cells = append(d.cells, nil)       // Extend by one.
	copy(d.cells[row+1:], d.cells[row:]) // Shift down.
	d.cells[row] = d.createRow()         // New row is initialized.
	d.DrawXYCoordinates()
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
	tblCell := tview.NewTableCell("")
	tblCell.SetMaxWidth(cell.Width)

	displayedText, _, err := cell.Calculate()
	if err != nil {
		tblCell.SetText(err.Error())
	} else {
		tblCell.SetText(displayedText)
	}

	if cell.IsInRange {
		tblCell.SetTextColor(tcell.ColorGreen)
	}
	if cell.IsFormula() {
		tblCell.SetTextColor(tcell.ColorGreen)
	}
	if cell.IsSelected {
		tblCell.SetTextColor(tcell.ColorBlue)
	}
	if err != nil {
		tblCell.SetTextColor(tcell.ColorRed)
	}

	tblCell.SetAttributes(cell.Attributes)
	tblCell.SetAlign(cell.Align)
	tblCell.SetMaxWidth(cell.Width)

	return tblCell
}

func (d *Data) HighlightFormulaRange(h *FormulaRange) {
	for row := h.StartRow + 1; row <= h.EndRow+1; row++ {
		for col := h.StartCol + 1; col <= h.EndCol+1; col++ {
			d.GetDataCell(row, col).IsInRange = true
		}
	}
}

func (d *Data) ClearFormulaRange(h *FormulaRange) {
	for row := h.StartRow + 1; row <= h.EndRow+1; row++ {
		for col := h.StartCol + 1; col <= h.EndCol+1; col++ {
			d.GetDataCell(row, col).IsInRange = false
		}
	}
}

func (d *Data) SelectCells(s *Selection) {
	d.logger.Info(fmt.Sprintf("Data.SelectCells: %v", s))
	for row := s.topRow; row <= s.bottomRow; row++ {
		for col := s.leftCol; col <= s.rightCol; col++ {
			d.GetDataCell(row, col).IsSelected = true

		}
	}
}

func (d *Data) ClearSelection() {
	d.logger.Info("Data.ClearSelection: cleared selection")
	for _, row := range d.cells {
		for _, cell := range row {
			cell.IsSelected = false
		}
	}
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
	d.DrawXYCoordinates()
}

func (d *Data) RemoveRows(fromRow int, toRow int) {
	if d.GetRowCount() == 2 {
		return
	}
	if fromRow <= 0 || toRow >= d.GetRowCount() {
		return // Invalid row index
	}
	d.cells = append(d.cells[:fromRow], d.cells[toRow+1:]...)
	d.DrawXYCoordinates()
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
	d.DrawXYCoordinates()
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
		aText, _, _ := a.Calculate()
		bText, _, _ := b.Calculate()
		return aText < bText
	})
	d.sortedCol = col
	d.sortOrder = ascIndicator
	d.DrawXYCoordinates()
}

func (d *Data) SortColStrDesc(col int) {
	d.sortColumn(col, func(a, b *Cell) bool {
		aText, _, _ := a.Calculate()
		bText, _, _ := b.Calculate()
		return aText > bText
	})
	d.sortedCol = col
	d.sortOrder = descIndicator
	d.DrawXYCoordinates()
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
func (d *Data) SnapShotCells() [][]*Cell {
	var dest [][]*Cell
	for _, row := range d.cells {
		var rowCopy []*Cell
		for _, cell := range row {
			rowCopy = append(rowCopy, cell)
		}
		dest = append(dest, rowCopy)
	}
	return dest
}

func (d *Data) RestoreSnapshot(snapshot [][]*Cell) {
	d.cells = snapshot
	d.DrawXYCoordinates()
}

func (d *Data) DrawXYCoordinates() {
	// Write row numbers.
	for rowIdx := range d.cells {
		cell := d.cells[rowIdx][0]
		cell.Text = fmt.Sprintf("%d", rowIdx-1)
		cell.Attributes = tcell.AttrDim
		cell.Align = 1 //AlignCenter
		if rowIdx == d.CurrentRow() {
			cell.Attributes = tcell.AttrBold
			cell.Attributes = tcell.AttrUnderline
		}
	}
	// Write column numbers.
	for colIdx, cell := range d.cells[0] {
		colText := fmt.Sprintf("%d", colIdx-1)
		cell.Attributes = tcell.AttrDim
		cell.Align = 1 //AlignCenter
		if colIdx == d.CurrentCol() {
			cell.Attributes = tcell.AttrBold
			cell.Attributes = tcell.AttrUnderline
		}
		if d.sortedCol != -1 {
			if colIdx == d.sortedCol {
				colText = colText + d.sortOrder
			}
		}
		cell.Text = colText
	}

	d.cells[0][0].Text = ""
}
