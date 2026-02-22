package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type setupModel struct {
	width    int
	height   int
	bombs    int
	done     bool
	cursor   int // for menu selection
	subStage int // 0 = menu, 1 = custom dimensions
}

var presets = []struct {
	name   string
	width  int
	height int
	bombs  int
}{
	{"Easy (9x9; 10 mines)", 9, 9, 10},
	{"Intermediate (16x16; 40 mines)", 16, 16, 40},
	{"Expert (16x30; 99 mines)", 30, 16, 99},
	{"Custom", 0, 0, 0},
}

func initialSetupModel() setupModel {
	return setupModel{
		width:    10,
		height:   10,
		bombs:    15,
		subStage: 0,
		cursor:   0,
	}
}

func (m setupModel) Init() tea.Cmd {
	return nil
}

func (m setupModel) Update(msg tea.Msg) (setupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.subStage == 0 {
			switch msg.String() {
			case "up", "w", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "s", "j":
				if m.cursor < len(presets)-1 {
					m.cursor++
				}
			case "enter":
				p := presets[m.cursor]
				if p.name == "Custom" {
					m.subStage = 1
				} else {
					m.width, m.height, m.bombs = p.width, p.height, p.bombs
					m.done = true
				}
			}
			return m, nil
		}

		// SubStage 1: Custom Dimensions
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
		case "esc", "backspace":
			m.subStage = 0
		}
	}
	return m, nil
}

func (m setupModel) View() string {
	var body strings.Builder

	if m.subStage == 0 {
		body.WriteString("Select Difficulty:\n\n")
		for i, p := range presets {
			cursor := "  "
			style := lipgloss.NewStyle()
			if m.cursor == i {
				cursor = "> "
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
			}
			body.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(p.name)))
		}
		body.WriteString("\nArrows/WASD to move, Enter to select.")
	} else {
		body.WriteString("Custom Dimensions:\n\n")
		body.WriteString(fmt.Sprintf("Width  : %d\n", m.width))
		body.WriteString(fmt.Sprintf("Height : %d\n", m.height))
		body.WriteString(fmt.Sprintf("Bombs  : %d\n", m.bombs))
		body.WriteString("\nArrows/WASD to adjust, +/- for bombs.\nEnter to start, Esc to go back.")
	}

	content := boardStyle.Render(body.String())

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render("Minesweeper Setup"),
		content,
	)
}
