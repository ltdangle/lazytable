package data

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var d *Data
var formulas []Formula
var floatFormat = "%.2f"

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
func (cell *Cell) Calculate() *highlight {
	if cell.IsFormula() {
		cell.text = strings.ReplaceAll(cell.text, " ", "") // remove spaces
		fText := cell.text[1:]                             // strip leading =
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

type SumFormula struct {
}

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
	if startCol < 0 || startRow < 0 || endRow >= len(d.Cells) || endCol >=
		len(d.Cells[0]) {
		return 0, fmt.Errorf("coordinates out of bounds")
	}

	// Sum cells in the range [startX:endX, startY:endY]
	for y := startRow; y <= endRow; y++ {
		for x := startCol; x <= endCol; x++ {
			val, err := strconv.ParseFloat(d.Cells[y][x].TableCell.Text, 64)
			if err != nil {
				return 0, fmt.Errorf("%d,%d is not a number", y-1, x-1)
			}
			sum += val
		}
	}
	return sum, nil
}

// Highlighted cells region.
// TODO: use getter and setter and check validity in setter
type highlight struct {
	startRow int
	startCol int
	endRow   int
	endCol   int
}

func NewHighlight() *highlight {
	return &highlight{}
}

// Data type.
type Data struct {
	Cells      [][]*Cell
	currentRow int
	currentCol int
	sortedCol  int
	sortOrder  string
	Highlight  *highlight
}

func NewData(frmls []Formula) *Data {
	formulas = frmls
	d = &Data{sortedCol: -1, sortOrder: ""}
	return d
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
	d.Cells = nil
}
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
func (d *Data) InsertRow(row int) {
	if row > d.GetRowCount() {
		return
	}

	d.Cells = append(d.Cells, nil)       // Extend by one.
	copy(d.Cells[row+1:], d.Cells[row:]) // Shift down.
	d.Cells[row] = d.createRow()         // New row is initialized.
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
	d.Cells = append(d.Cells, dataRow)
}

// TODO: manage cell attibutes (color, etc) explicitly here
func (d *Data) GetCell(row, column int) *tview.TableCell {
	// Coordinates are outside our table.
	if row > d.GetRowCount()-1 || column > d.GetColumnCount()-1 {
		return nil
	}

	cell := d.Cells[row][column]
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

	d.highlightCell(row, column, cell)

	return cell.TableCell
}

// Highlight cell range.
func (d *Data) highlightCell(row int, column int, cell *Cell) {
	if d.Highlight == nil {
		cell.SetAttributes(tcell.AttrNone)
		return
	}

	cellIsHighlighted := row >= d.Highlight.startRow+1 && column >= d.Highlight.startCol+1 && row <= d.Highlight.endRow+1 && column <= d.Highlight.endCol+1
	if cellIsHighlighted {
		cell.SetTextColor(tcell.ColorGreen)
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

	return d.Cells[d.CurrentRow()][d.CurrentCol()]
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
	sort.SliceStable(d.Cells[2:], func(i, j int) bool {
		return sorter(d.Cells[i+2][col], d.Cells[j+2][col])
	})
}
func (d *Data) DrawXYCoordinates() {
	for rowIdx := range d.Cells {
		d.Cells[rowIdx][0].SetText(fmt.Sprintf("%d", rowIdx-1))
	}
	for colIdx, col := range d.Cells[0] {
		colText := fmt.Sprintf("%d", colIdx-1)
		if d.sortedCol != -1 {
			if colIdx == d.sortedCol {
				colText = colText + d.sortOrder
			}
		}
		col.SetText(colText)
	}

	d.Cells[0][0].SetText("")
}
