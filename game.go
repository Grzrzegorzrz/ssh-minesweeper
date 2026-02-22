package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type gameStatus int

const (
	playing gameStatus = iota
	won
	lost
)

type cell struct {
	isBomb    bool
	revealed  bool
	flagged   bool
	neighbors int
}

type gameState struct {
	grid       [][]cell
	cursorX    int
	cursorY    int
	status     gameStatus
	bombsLeft  int
	cellsLeft  int
	firstClick bool
	gFlag      int
	width      int
	height     int
	numBombs   int
}

func newGameState(width, height, numBombs int) gameState {
	grid := make([][]cell, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]cell, width)
	}
	return gameState{
		grid:       grid,
		status:     playing,
		bombsLeft:  numBombs,
		cellsLeft:  width*height - numBombs,
		firstClick: true,
		width:      width,
		height:     height,
		numBombs:   numBombs,
	}
}

func (m gameState) Init() tea.Cmd {
	return nil
}

func (m gameState) Update(msg tea.Msg) (gameState, tea.Cmd) {
	if m.gFlag == 1 {
		m.gFlag = 2
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "w", "k":
			if m.cursorY > 0 {
				m.cursorY--
			}
		case "down", "s", "j":
			if m.cursorY < m.height-1 {
				m.cursorY++
			}
		case "left", "a", "h":
			if m.cursorX > 0 {
				m.cursorX--
			}
		case "right", "d", "l":
			if m.cursorX < m.width-1 {
				m.cursorX++
			}
		case " ", "enter":
			if m.status == playing {
				m.reveal(m.cursorX, m.cursorY)
			} else {
				return newGameState(m.width, m.height, m.numBombs), nil
			}
		case "G":
			m.cursorY = m.height - 1
		case "g":
			if m.gFlag == 2 {
				m.cursorY = 0
				m.gFlag = 0
			} else {
				m.gFlag = 1
			}
		case "$":
			m.cursorX = m.width - 1
		case "0":
			m.cursorX = 0
		case "f":
			if m.status == playing {
				m.toggleFlag(m.cursorX, m.cursorY)
			}
		case "r":
			return newGameState(m.width, m.height, m.numBombs), nil
		}
	}

	if m.gFlag == 2 {
		m.gFlag = 0
	}
	return m, nil
}

func (m *gameState) generateBombs(firstX, firstY int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	count := 0
	for count < m.numBombs {
		x := r.Intn(m.width)
		y := r.Intn(m.height)
		if (x == firstX && y == firstY) || m.grid[y][x].isBomb {
			continue
		}
		m.grid[y][x].isBomb = true
		count++
	}

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			if m.grid[y][x].isBomb {
				continue
			}
			m.grid[y][x].neighbors = m.countNeighbors(x, y)
		}
	}
}

func (m gameState) countNeighbors(x, y int) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < m.width && ny >= 0 && ny < m.height && m.grid[ny][nx].isBomb {
				count++
			}
		}
	}
	return count
}

func (m *gameState) reveal(x, y int) {
	if m.grid[y][x].revealed || m.grid[y][x].flagged {
		return
	}

	if m.firstClick {
		m.generateBombs(x, y)
		m.firstClick = false
	}

	m.grid[y][x].revealed = true

	if m.grid[y][x].isBomb {
		m.status = lost
		return
	}

	m.cellsLeft--
	if m.cellsLeft == 0 {
		m.status = won
		return
	}

	if m.grid[y][x].neighbors == 0 {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < m.width && ny >= 0 && ny < m.height {
					m.reveal(nx, ny)
				}
			}
		}
	}
}

func (m *gameState) toggleFlag(x, y int) {
	if m.grid[y][x].revealed {
		return
	}
	if m.grid[y][x].flagged {
		m.grid[y][x].flagged = false
		m.bombsLeft++
	} else {
		m.grid[y][x].flagged = true
		m.bombsLeft--
	}
}

func (m gameState) View() string {
	var board strings.Builder

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			c := m.grid[y][x]
			var char string
			var style lipgloss.Style

			if c.revealed {
				if c.isBomb {
					char = "*"
					style = bombStyle
				} else if c.neighbors > 0 {
					char = fmt.Sprintf("%d", c.neighbors)
					style = numStyles[c.neighbors-1]
				} else {
					char = "."
					style = hiddenStyle
				}
			} else if c.flagged {
				char = "F"
				style = flagStyle
			} else {
				char = "#"
				style = hiddenStyle
			}

			rendered := style.Render(fmt.Sprintf(" %s ", char))
			if x == m.cursorX && y == m.cursorY {
				rendered = cursorStyle.Render(fmt.Sprintf(" %s ", char))
			}
			board.WriteString(rendered)
		}
		if y < m.height-1 {
			board.WriteString("\n")
		}
	}

	gameView := boardStyle.Render(board.String())

	var s strings.Builder
	s.WriteString(titleStyle.Render("Minesweeper Over SSH!"))
	s.WriteString("\n")
	s.WriteString(gameView)
	s.WriteString(statusStyle.Render(fmt.Sprintf("\nBombs Hidden: %d | Remaining: %d | Status: %s", m.numBombs, m.cellsLeft, getStatusText(m))))

	if m.status != playing {
		var extraMsg string
		if m.status == lost {
			extraMsg = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("\n\nGAME OVER! Press r to restart or q to quit.")
		} else {
			extraMsg = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render("\n\nYOU WIN! Press r to restart or q to quit.")
		}
		s.WriteString(extraMsg)
	} else {
		s.WriteString("\n\nMove: arrows/wasd/hjkl\nFirst/Last row: gg/G\nFirst/Last column 0/$:\n\nReveal: space/enter | Flag: f | Restart: r | Quit: q")
	}

	return s.String()
}

func getStatusText(m gameState) string {
	switch m.status {
	case lost:
		return "Defeat"
	case won:
		return "Victory"
	default:
		return "Sweeping..."
	}
}

var (
	cursorStyle = lipgloss.NewStyle().Background(lipgloss.Color("7"))
	bombStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	flagStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	numStyles   = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("13")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("14")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("0")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	}
	hiddenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).MarginBottom(1)
	statusStyle = lipgloss.NewStyle().Italic(true).MarginTop(1)
	boardStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)
)
