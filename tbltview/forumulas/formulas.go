package formulas

import (
	"fmt"
	"regexp"
	"strconv"
	d "tblview/data"
)

var floatFormat = "%.2f"

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

func (f *SumFormula) Calculate(data *d.Data, text string) (string, *d.FormulaRange, error) {
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
	total, err := f.sum(data, startRow+1, startCol+1, endRow+1, endCol+1)
	if err != nil {
		return "", nil, err
	}

	highlight := d.NewFormulaRange()
	highlight.StartRow = startRow
	highlight.StartCol = startCol
	highlight.EndRow = endRow
	highlight.EndCol = endCol

	return fmt.Sprintf(floatFormat, total), highlight, nil
}

func (f *SumFormula) sum(data *d.Data, startRow, startCol, endRow, endCol int) (float64, error) {
	sum := 0.0

	// Validate the coordinates
	if startCol > endCol || startRow > endRow {
		return 0, fmt.Errorf("start coordinates must be less than or equal to end coordinates")
	}
	if startCol < 0 || startRow < 0 || endRow >= data.GetRowCount() || endCol >= len(data.GetRow(0)) {
		return 0, fmt.Errorf("coordinates out of bounds")
	}

	// Sum cells in the range [startX:endX, startY:endY]
	for y := startRow; y <= endRow; y++ {
		for x := startCol; x <= endCol; x++ {
			val, err := strconv.ParseFloat(data.GetDataCell(y, x).Text, 64)
			if err != nil {
				return 0, fmt.Errorf("%d,%d is not a number", y-1, x-1)
			}
			sum += val
		}
	}
	return sum, nil
}
