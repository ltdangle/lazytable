package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Table struct {
	cells [][]*Cell
}

func NewTable() *Table {
	return &Table{
		cells: [][]*Cell{
			{&Cell{}, &Cell{}, &Cell{}, &Cell{}, &Cell{}},
			{&Cell{}, &Cell{}, &Cell{}, &Cell{}, &Cell{}},
			{&Cell{}, &Cell{}, &Cell{}, &Cell{}, &Cell{}},
		},
	}
}
func (m *Table) Init() tea.Cmd {
	return nil
}

func (m *Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *Table) View() string {
	var builder strings.Builder

	for _, row := range m.cells {
		builder.WriteString("| ")
		for _, cell := range row {
			builder.WriteString(cell.View())
			builder.WriteString(" | ")
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
