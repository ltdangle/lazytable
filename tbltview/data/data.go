package data

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
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
	text string
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
	cell.TableCell.SetTextColor(tcell.ColorRed)
}
func (cell *Cell) Calculate() *Highlight {
	if cell.IsFormula() {
		cell.text = strings.ReplaceAll(cell.text, " ", "") // remove spaces
		fText := cell.text[1:]                             // strip leading =
		for _, formula := range formulas {
			isMatch, _ := formula.Match(fText)
			if isMatch {
				calculated, highlight, err := formula.Calculate(d, fText)
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
	Calculate(data *Data, text string) (string, *Highlight, error)
}

// Highlighted cells region.
// TODO: use getter and setter and check validity in setter
type Highlight struct {
	StartRow int
	StartCol int
	EndRow   int
	EndCol   int
}

func NewHighlight() *Highlight {
	return &Highlight{}
}

type Selection struct {
	StartRow int
	StartCol int
	EndRow   int
	EndCol   int
}

func NewSelection(startRow int, startCol int, endRow int, endCol int) *Selection {
	return &Selection{StartRow: startRow, StartCol: startCol, EndRow: endRow, EndCol: endCol}
}

// Data type.
type Data struct {
	cells      [][]*Cell
	currentRow int
	currentCol int
	sortedCol  int
	sortOrder  string
}

func NewData(frmls []Formula) *Data {
	formulas = frmls
	d = &Data{sortedCol: -1, sortOrder: ""}
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

// TODO: manage cell attibutes (color, etc) explicitly here
func (d *Data) GetCell(row, column int) *tview.TableCell {
	// Coordinates are outside our table.
	if row > d.GetRowCount()-1 || column > d.GetColumnCount()-1 {
		return nil
	}

	return d.cells[row][column].TableCell

}

func (d *Data) HighlightCells(h *Highlight) {
	for row := h.StartRow + 1; row <= h.EndRow+1; row++ {
		for col := h.StartCol + 1; col <= h.EndCol+1; col++ {
			d.GetDataCell(row, col).SetTextColor(tcell.ColorGreen)
		}
	}
}

func (d *Data) ClearHighlight(h *Highlight) {
	for row := h.StartRow + 1; row <= h.EndRow+1; row++ {
		for col := h.StartCol + 1; col <= h.EndCol+1; col++ {
			d.GetDataCell(row, col).SetTextColor(tcell.ColorDefault)
		}
	}
}

func (d *Data) SelectCells(s *Selection) {
	for row := s.StartRow ; row <= s.EndRow; row++ {
		for col := s.StartCol ; col <= s.EndCol; col++ {
			d.GetDataCell(row, col).SetAttributes(tcell.AttrReverse)
			// d.GetDataCell(row, col).SetTextColor(tcell.ColorBlue)
		}
	}
}

func (d *Data) ClearCellSelect(h *Selection) {
	for row := h.StartRow + 1; row <= h.EndRow+1; row++ {
		for col := h.StartCol + 1; col <= h.EndCol+1; col++ {
			d.GetDataCell(row, col).SetAttributes(tcell.AttrReverse)
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
