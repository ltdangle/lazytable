  # TUI Spreadsheet

    A simple text-based user interface (TUI) spreadsheet application with VIM-
  like keybindings implemented in Go.

    ## Features

    - Edit cells with vim-like keybindings
    - Support for simple formulas (e.g., SUM)
    - Load and save CSV files
    - Row and column manipulation: add, delete, sort
    - Basic undo/redo functionality
    - Integrated modal system for user alerts

    ## Installation

    Before running the application, ensure you have Go installed on your
  system.

    ```bash
    go get github.com/gdamore/tcell/v2
    go get github.com/rivo/tview

  You can then clone the repository and build the application:

    git clone https://github.com/yourgithub/tui-spreadsheet.git
    cd tui-spreadsheet
    go build

  ## Usage

  To run the program with a provided CSV file:

    ./tui-spreadsheet -file=path/to/your/file.csv

  ## Keybindings

  •  j / k / h / l : navigate through cells
  •  : : enter command mode (similar to vim)
  •  i : insert mode to edit the current cell
  •  o : insert a new row below
  •  O : insert a new row above
  •  d : delete current row
  •  f / F : sort column in ascending/descending order
  •  V / CTRL+V : select entire row/column
  •  ESC : return to cell selection mode
  •  u / CTRL+R : undo/redo last action

  ## Contributing

  If you would like to contribute to the development of this TUI spreadsheet,
  please follow the standard fork-branch-PR workflow.

  ## License

  This TUI Spreadsheet is open-source software licensed under the MIT license.

