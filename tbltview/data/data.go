package data

import (
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
	*tview.TableCell
	isSelected bool
	isInRange  bool // Cell is part of the formula range.
	isError    bool // Cell formula has error.
	text       string
}

func NewCell() *Cell {
	cell := &Cell{TableCell: tview.NewTableCell("")}
	cell.SetMaxWidth(10)
	return cell
}
func (cell *Cell) SetText(text string) {
	cell.text = text
	cell.Calculate()
}
func (cell *Cell) ShowError(text string) {
	cell.TableCell.SetText("#ERR:" + text)
	cell.isError = true
}
func (cell *Cell) Calculate() *FormulaRange {
	cell.isInRange = false
	cell.isError = false
	if cell.IsFormula() {
		cell.text = strings.ReplaceAll(cell.text, " ", "") // remove spaces
		fText := cell.text[1:]                             // strip leading =
		for _, formula := range formulas {
			isMatch, _ := formula.Match(fText)
			if isMatch {
				calculated, formulaRange, err := formula.Calculate(d, fText)
				if err != nil {
					cell.ShowError(err.Error())
					return nil
				}
				cell.TableCell.SetText(calculated)
				return formulaRange
			}
		}
		cell.ShowError("no formula")
		return nil
	}
	cell.TableCell.SetText(cell.text)
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
	startRow int // Inital selction point.
	startCol int // Initial selection point.

	topRow    int
	leftCol   int
	bottomRow int
	rightCol  int
}

func NewSelection(startRow int, startCol int, endRow int, endCol int) *Selection {
	return &Selection{
		startRow: startRow,
		startCol: startCol,

		topRow:    startRow,
		leftCol:   startCol,
		bottomRow: endRow,
		rightCol:  endCol,
	}
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
	cell.Calculate()

	cell.TableCell.SetTextColor(tcell.ColorWhite)

	if cell.isInRange {
		cell.TableCell.SetTextColor(tcell.ColorGreen)
	}
	if cell.IsFormula() {
		cell.TableCell.SetTextColor(tcell.ColorGreen)
	}
	if cell.isSelected {
		cell.TableCell.SetTextColor(tcell.ColorBlue)
	}
	if cell.isError {
		cell.TableCell.SetTextColor(tcell.ColorRed)
	}

	return cell.TableCell

}

func (d *Data) HighlightFormulaRange(h *FormulaRange) {
	for row := h.StartRow + 1; row <= h.EndRow+1; row++ {
		for col := h.StartCol + 1; col <= h.EndCol+1; col++ {
			d.GetDataCell(row, col).isInRange = true
		}
	}
}

func (d *Data) ClearFormulaRange(h *FormulaRange) {
	for row := h.StartRow + 1; row <= h.EndRow+1; row++ {
		for col := h.StartCol + 1; col <= h.EndCol+1; col++ {
			d.GetDataCell(row, col).isInRange = false
		}
	}
}

func (d *Data) SelectCells(s *Selection) {
	d.logger.Info(fmt.Sprintf("Data.SelectCells: %v", s))
	for row := s.topRow; row <= s.bottomRow; row++ {
		for col := s.leftCol; col <= s.rightCol; col++ {
			d.GetDataCell(row, col).isSelected = true

		}
	}
}

func (d *Data) ClearCellSelect(s *Selection) {
	d.logger.Info(fmt.Sprintf("Data.ClearCellSelect: %v", s))
	if s == nil {
		return
	}
	for row := s.topRow; row <= s.bottomRow; row++ {
		for col := s.leftCol; col <= s.rightCol; col++ {
			d.GetDataCell(row, col).isSelected = false
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
		return a.TableCell.Text < b.TableCell.Text // Compare the text of the cells for ascending order.
	})
	d.sortedCol = col
	d.sortOrder = ascIndicator
	d.DrawXYCoordinates()
}

func (d *Data) SortColStrDesc(col int) {
	d.sortColumn(col, func(a, b *Cell) bool {
		return a.TableCell.Text > b.TableCell.Text // Compare the text of the cells for descending order.
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

func (d *Data) DrawXYCoordinates() {
	// Write row numbers.
	for rowIdx := range d.cells {
		cell := d.cells[rowIdx][0]
		cell.SetText(fmt.Sprintf("%d", rowIdx-1))
		cell.SetAttributes(tcell.AttrDim)
		cell.SetAlign(1) //AlignCenter
		if rowIdx == d.CurrentRow() {
			cell.SetAttributes(tcell.AttrBold)
			cell.SetAttributes(tcell.AttrUnderline)
		}
	}
	// Write column numbers.
	for colIdx, cell := range d.cells[0] {
		colText := fmt.Sprintf("%d", colIdx-1)
		cell.SetAttributes(tcell.AttrDim)
		cell.SetAlign(1) //AlignCenter
		if colIdx == d.CurrentCol() {
			cell.SetAttributes(tcell.AttrBold)
			cell.SetAttributes(tcell.AttrUnderline)
		}
		if d.sortedCol != -1 {
			if colIdx == d.sortedCol {
				colText = colText + d.sortOrder
			}
		}
		cell.SetText(colText)
	}

	d.cells[0][0].SetText("")
}
