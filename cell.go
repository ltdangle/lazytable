package main

import tea "github.com/charmbracelet/bubbletea"

type Cell struct{}

func NewCell() *Cell {
	return &Cell{}
}

func (m *Cell) Init() tea.Cmd {
	return nil
}

func (m *Cell) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Cell) View() string {
	return "cell val"
}
