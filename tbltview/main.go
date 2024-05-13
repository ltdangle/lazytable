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
	lgr "tblview/logger"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var csvFile *string
var dta *data.Data
var table = tview.NewTable()
var app = tview.NewApplication()
var cellInput = tview.NewInputField()
var commandInput = tview.NewInputField()
var pages = tview.NewPages()
var modal func(p tview.Primitive, width, height int) tview.Primitive
var modalContents = tview.NewBox()
var history = NewHistory()
var logger = lgr.NewLogger("tmp/log.txt")

const MODE_NORMAL = "n"
const MODE_VISUAL = "v"
const MODE_VISUAL_LINE = "V"
const MODE_VISUAL_BLOCK = "Ctrl+V"

var selection *data.Selection
var mode string = MODE_NORMAL

var clmCommands []ClmCommand

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
	dta = data.NewData(frmls, logger)
	// Init selection.
	selection = data.NewSelection(dta)
	// Build clm command.
	clmCommands = append(clmCommands, NewSortColStrAscClmCommand(), NewReplaceClmCommand())

	// Load csv file data.
	readCsvFile(*csvFile, dta)

	// Set cursor to the first cell.
	dta.SetCurrentRow(1)
	dta.SetCurrentCol(1)

	buildCellInput()
	buildTable()
	buildModal()
	buildCommandInput()

	// Configure layout.
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cellInput, 1, 0, false).
		AddItem(table, 0, 1, false).
		AddItem(commandInput, 1, 0, false)

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

	dta.DrawXYCoordinates()
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
		SetSelectionChangedFunc(wrapSelectionChangedFunc()).
		SetInputCapture(wrapInputCapture())
}

func buildCellInput() {
	cellInput.
		SetFieldStyle(tcell.StyleDefault).
		SetLabel(fmt.Sprintf("%d:%d ", dta.CurrentRow()-1, dta.CurrentCol()-1)).
		SetText(dta.GetCurrentCell().GetText()).
		SetDoneFunc(
			func(key tcell.Key) {
				logger.Info(fmt.Sprintf("cellInput.SetDoneFunc: %v", key))
				app.SetFocus(table)
				// Push cursor down, if possible.
				if dta.CurrentRow() < dta.GetRowCount()-1 {
					dta.SetCurrentRow(dta.CurrentRow() + 1)
				}
				table.Select(dta.CurrentRow(), dta.CurrentCol())
			}).
		SetChangedFunc(wrapChangedFunc())
}

func buildCommandInput() {
	commandInput.SetFieldStyle(tcell.StyleDefault)
	commandInput.SetDoneFunc(
		func() func(key tcell.Key) {
			var error string
			return func(key tcell.Key) {
				if error != "" {
					error = ""
					commandInput.SetLabel("")
					commandInput.SetText("")
					app.SetFocus(table)
					return
				}

				text := commandInput.GetText()
				logger.Info(fmt.Sprintf("commandInput.SetDoneFunc: key: %v, text: %s", key, text))
				for _, clmCommand := range clmCommands {
					match, commandError, command := clmCommand.Match(text)
					if match {
						if commandError != nil {
							error = commandError.Error()
							commandInput.SetText(error)
							return
						}
						history.Do(command)
						commandInput.SetLabel("")
						commandInput.SetText("")
						mode = MODE_NORMAL
						app.SetFocus(table)
						return
					}
				}

				error = "command not found. press ENTER to continue"
				commandInput.SetText(error)
			}
		}(),
	).
		SetChangedFunc(func(text string) {
			logger.Info(fmt.Sprintf("commandInput.SetChangedFunc: %s", text))
		})
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
func wrapSelectionChangedFunc() func(row, col int) {
	var hihglight *data.FormulaRange
	return func(row, col int) {
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

		// Clear previos highlights.
		if hihglight != nil {
			dta.ClearFormulaRange(hihglight)
		}

		// Highlight cells for the formula.
		hihglight = dta.GetCurrentCell().Calculate()
		if hihglight != nil {
			dta.HighlightFormulaRange(hihglight)
		}
		switch mode {
		case MODE_VISUAL:
			dta.ClearSelection(selection)
			selection.Update(row, col)
			dta.SelectCells(selection)
		case MODE_VISUAL_LINE:
			dta.ClearSelection(selection)
			selection.Update(row, dta.GetColumnCount()-1)
			dta.SelectCells(selection)
		case MODE_VISUAL_BLOCK:
			dta.ClearSelection(selection)
			selection.Update(dta.GetRowCount()-1, col)
			dta.SelectCells(selection)
		}
		dta.DrawXYCoordinates()
	}
}

func wrapChangedFunc() func(text string) {
	var hihglight *data.FormulaRange
	return func(text string) {
		logger.Info(fmt.Sprintf("cellInput.SetChangedFunc: %v", text))
		// This function is called whenever cursor changes position.
		// So we need to check if the value actually changed.
		prevVal := dta.GetCurrentCell().GetText()
		if prevVal != text {
			history.Do(NewChangeCellValueCommand(dta.CurrentRow(), dta.CurrentCol(), text))
		}

		// Clear previos highlights.
		if hihglight != nil {
			dta.ClearFormulaRange(hihglight)
		}

		// Highlight cells for the formula.
		hihglight = dta.GetCurrentCell().Calculate()
		if hihglight != nil {
			dta.HighlightFormulaRange(hihglight)
		}
	}
}

func wrapInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		logger.Info(fmt.Sprintf("table.SetInputCapture: rune - %v, key - %v, modifier - %v, name - %v", event.Rune(), event.Key(), event.Modifiers(), event.Name()))
		row, col := table.GetSelection()

		// Normal mode.
		switch event.Name() {
		case "Rune[:]":
			commandInput.SetLabel(":")
			app.SetFocus(commandInput)
		case "Rune[i]":
			app.SetFocus(cellInput)
		case "Rune[v]":
			mode = MODE_VISUAL
			selection.SetCoordintates(row, col, row, col)
			dta.SelectCells(selection)
			logger.Info("visual mode")
		case "Rune[V]":
			mode = MODE_VISUAL_LINE
			selection.SetCoordintates(row, 1, row, dta.GetColumnCount()-1)
			dta.SelectCells(selection)
			logger.Info("visual line mode")
		case "Ctrl+V":
			mode = MODE_VISUAL_BLOCK
			selection.SetCoordintates(1, col, dta.GetRowCount()-1, col)
			dta.SelectCells(selection)
			logger.Info("visual block mode")
		case "Esc":
			table.SetSelectable(true, true)
			mode = MODE_NORMAL
			dta.ClearSelection(selection)
			logger.Info("normal mode")
		case "Rune[d]":
			commandInput.SetText(fmt.Sprintf("Row selected: %v, col selected: %v", selection.IsRowSelected(), selection.IsColumnSelected()))
			switch mode {
			case MODE_VISUAL_LINE:
				if selection.IsRowSelected() {
					history.Do(NewDeleteRowCommand(row, col))
					dta.ClearSelection(selection)
					mode = MODE_NORMAL
				}
			case MODE_VISUAL_BLOCK:
				if selection.IsColumnSelected() {
					history.Do(NewDeleteColumnCommand(row, col))
					dta.ClearSelection(selection)
					mode = MODE_NORMAL
				}

			}
		case "Rune[>]": // Increase column width.
			history.Do(NewIncreaseColWidthCommand(dta.CurrentCol()))
		case "Rune[<]": // Decrease column width.
			history.Do(NewDecreaseColWidthCommand(dta.CurrentCol()))
		case "Rune[f]": // Sort string values asc.
			history.Do(NewSortColStrAscCommand(dta.CurrentCol()))
		case "Rune[F]": // Sort string values desc.
			history.Do(NewSortColStrDescCommand(dta.CurrentCol()))
		case "Rune[o]": // Insert row below.
			history.Do(NewInsertRowBelowCommand(dta.CurrentRow()))
		case "Rune[O]": // Insert row above.
			history.Do(NewInsertRowAboveCommand(dta.CurrentRow(), dta.CurrentCol()))
		case "Rune[a]":
			history.Do(NewInsertColRightCommand(dta.CurrentCol()))
		case "Rune[I]":
			history.Do(NewInsertColLeftCommand(dta.CurrentRow(), dta.CurrentCol()))
		case "Rune[u]":
			history.Undo()
		case "Ctrl+R":
			history.Redo()
		}

		return event
	}
}
