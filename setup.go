package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type setupModel struct {
	width  int
	height int
	bombs  int
	done   bool
}

func initialSetupModel() setupModel {
	return setupModel{
		width:  10,
		height: 10,
		bombs:  15,
	}
}

func (m setupModel) Init() tea.Cmd {
	return nil
}

func (m setupModel) Update(msg tea.Msg) (setupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "a", "h":
			if m.width > 5 {
				m.width--
			}
		case "right", "d", "l":
			if m.width < 40 {
				m.width++
			}
		case "up", "w", "k":
			if m.height < 24 {
				m.height++
			}
		case "down", "s", "j":
			if m.height > 5 {
				m.height--
			}
		case "+":
			if m.bombs < m.width*m.height-1 {
				m.bombs++
			}
		case "-":
			if m.bombs > 1 {
				m.bombs--
			}
		case "r":
			m.width, m.height, m.bombs = 10, 10, 15
		case "enter":
			m.done = true
		}
	}
	return m, nil
}

func (m setupModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Minesweeper Setup"))
	b.WriteString("\n\nUse arrows/wasd to set dimensions, +/- for bombs. Enter to start.\n\n")
	b.WriteString(fmt.Sprintf("Width  : %d\n", m.width))
	b.WriteString(fmt.Sprintf("Height : %d\n", m.height))
	b.WriteString(fmt.Sprintf("Bombs  : %d\n", m.bombs))
	return boardStyle.Render(b.String())
}
