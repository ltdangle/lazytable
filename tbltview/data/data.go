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
	IsInRange bool   `json:"isInRange"`
	Text      string `json:"text"`

	// tview.TableCell attributes
	Attributes tcell.AttrMask `json:"attributes"`
	Align      int            `json:"align"`
	Width      int            `json:"width"`
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
	Cells      [][]*Cell `json:"cells"`
	CurrentRow int       `json:"currentRow"`
	CurrentCol int       `json:"currentCol"`
	SortedCol  int       `json:"sortedCol"`
	SortOrder  string    `json:"sortOrder"`
	Selection  *Selection
	logger     *logger.Logger
}

func NewData(frmls []Formula, logger *logger.Logger) *Data {
	formulas = frmls
	d = &Data{SortedCol: -1, SortOrder: "", logger: logger}
	d.Selection = NewSelection(d)
	return d
}
func (d *Data) GetCells() [][]*Cell {
	return d.Cells
}
func (d *Data) GetDataCell(row int, col int) *Cell {
	return d.Cells[row][col]
}

// TODO: mutator
func (d *Data) SetDataCell(row int, col int, cell *Cell) {
	d.Cells[row][col] = cell
}

func (d *Data) CopyDataCell(row int, col int, cell *Cell) {
	_ = copier.Copy(d.Cells[row][col], cell)
}
func (d *Data) GetRow(row int) []*Cell {
	return d.Cells[row]
}
func (d *Data) GetCol(column int) []*Cell {
	var col []*Cell
	for row := range d.Cells {
		col = append(col, d.Cells[row][column])
	}
	return col
}
// TODO: mutator
func (d *Data) Clear() {
	d.Cells = nil
}
// TODO: mutator
func (d *Data) InsertColumn(column int) {
	for row := range d.Cells {
		if column > len(d.Cells[row]) {
			continue
		}
		d.Cells[row] = append(d.Cells[row], nil)             // Extend by one.
		copy(d.Cells[row][column+1:], d.Cells[row][column:]) // Shift to the right.
		d.Cells[row][column] = NewCell()
		d.DrawXYCoordinates()
	}
}
// TODO: mutator
func (d *Data) InsertRow(row int) {
	if row > d.GetRowCount() {
		return
	}

	d.Cells = append(d.Cells, nil)       // Extend by one.
	copy(d.Cells[row+1:], d.Cells[row:]) // Shift down.
	d.Cells[row] = d.createRow()         // New row is initialized.
	d.DrawXYCoordinates()
}

// TODO: mutator
func (d *Data) AddDataRow(dataRow []*Cell) {
	d.Cells = append(d.Cells, dataRow)
}

func (d *Data) GetCell(row, column int) *tview.TableCell {
	// Coordinates are outside our table.
	if row > d.GetRowCount()-1 || column > d.GetColumnCount()-1 {
		return nil
	}

	cell := d.Cells[row][column]
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
	// Check if the cell is selected and format accordingly.
	if row >= d.Selection.topRow && row <= d.Selection.bottomRow && column >= d.Selection.leftCol && column <= d.Selection.rightCol {
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

func (d *Data) SetCell(row, column int, cell *tview.TableCell) {
	// Part of the tview.TableContent interface.
}

func (d *Data) GetRowCount() int {
	return len(d.Cells)
}

func (d *Data) GetColumnCount() int {
	return len(d.Cells[0])
}

// TODO: mutator
func (d *Data) RemoveRow(row int) {
	if d.GetRowCount() == 2 {
		return
	}
	if row <= 0 || row >= d.GetRowCount() {
		return // Invalid row index
	}
	d.Cells = append(d.Cells[:row], d.Cells[row+1:]...)
	d.DrawXYCoordinates()
}

// TODO: mutator
func (d *Data) RemoveRows(fromRow int, toRow int) {
	if d.GetRowCount() == 2 {
		return
	}
	if fromRow <= 0 || toRow >= d.GetRowCount() {
		return // Invalid row index
	}
	d.Cells = append(d.Cells[:fromRow], d.Cells[toRow+1:]...)
	d.DrawXYCoordinates()
}

// TODO: mutator
func (d *Data) RemoveColumn(col int) {
	if d.GetColumnCount() == 2 {
		return
	}
	if col <= 0 || col >= d.GetColumnCount() {
		return // Invalid column index
	}
	for i := range d.Cells {
		d.Cells[i] = append(d.Cells[i][:col], d.Cells[i][col+1:]...)
	}
	d.DrawXYCoordinates()
}

// TODO: mutator
func (d *Data) createRow() []*Cell {
	var row []*Cell
	for i := 0; i < d.GetColumnCount(); i++ {
		row = append(row, NewCell())
	}
	return row
}

func (d *Data) GetCurrentCell() *Cell {
	// Check of out of bounds values.
	if d.CurrentRow >= d.GetRowCount() {
		return NewCell()
	}
	if d.CurrentCol >= d.GetColumnCount() {
		return NewCell()
	}

	return d.Cells[d.CurrentRow][d.CurrentCol]
}

// Sort column  string values.
// TODO: mutator
func (d *Data) SortColStrAsc(col int) {
	d.sortColumn(col, func(a, b *Cell) bool {
		aText, _, _ := a.Calculate()
		bText, _, _ := b.Calculate()
		return aText < bText
	})
	d.SortedCol = col
	d.SortOrder = ascIndicator
	d.DrawXYCoordinates()
}

// TODO: mutator
func (d *Data) SortColStrDesc(col int) {
	d.sortColumn(col, func(a, b *Cell) bool {
		aText, _, _ := a.Calculate()
		bText, _, _ := b.Calculate()
		return aText > bText
	})
	d.SortedCol = col
	d.SortOrder = descIndicator
	d.DrawXYCoordinates()
}

// Sorts column. Accept column index and a sorter function that
// takes slice of vertical column cells as an argument.
// TODO: mutator
func (d *Data) sortColumn(col int, sorter func(a, b *Cell) bool) {
	// Perform a stable sort to maintain the relative order of other elements.
	// Account for cols row and header row (+2)
	sort.SliceStable(d.Cells[2:], func(i, j int) bool {
		return sorter(d.Cells[i+2][col], d.Cells[j+2][col])
	})
}

func (d *Data) SnapShotCells() [][]Cell {
	var dest [][]Cell
	for _, row := range d.Cells {
		var rowCopy []Cell
		for _, cell := range row {
			rowCopy = append(rowCopy, *cell)
		}
		dest = append(dest, rowCopy)
	}
	return dest
}

func (d *Data) RestoreSnapshot(snapshot [][]Cell) {
	// Initialize cells with the required capacity.
	d.Cells = make([][]*Cell, len(snapshot))
	// Copy snapshot.
	for i, row := range snapshot {
		rowCopy := make([]*Cell, len(row))
		for j, cell := range row {
			cellCopy := cell       // create a copy of the cell
			rowCopy[j] = &cellCopy // assign the copy's address
		}
		d.Cells[i] = rowCopy // assign the row copy to the corresponding index in d.Cells
	}
	d.DrawXYCoordinates()
}

func (d *Data) DrawXYCoordinates() {
	// Write row numbers.
	for rowIdx := range d.Cells {
		cell := d.Cells[rowIdx][0]
		cell.Text = fmt.Sprintf("%d", rowIdx-1)
		cell.Attributes = tcell.AttrDim
		cell.Align = 1 //AlignCenter
		if rowIdx == d.CurrentRow {
			cell.Attributes = tcell.AttrBold
			cell.Attributes = tcell.AttrUnderline
		}
	}
	// Write column numbers.
	for colIdx, cell := range d.Cells[0] {
		colText := fmt.Sprintf("%d", colIdx-1)
		cell.Attributes = tcell.AttrDim
		cell.Align = 1 //AlignCenter
		if colIdx == d.CurrentCol {
			cell.Attributes = tcell.AttrBold
			cell.Attributes = tcell.AttrUnderline
		}
		if d.SortedCol != -1 {
			if colIdx == d.SortedCol {
				colText = colText + d.SortOrder
			}
		}
		cell.Text = colText
	}

	d.Cells[0][0].Text = ""
}
