package main

import (
	"fmt"
	"math/rand/v2"
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

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

type gameState struct {
	grid        [][]cell
	cursorX     int
	cursorY     int
	status      gameStatus
	bombsLeft   int
	cellsLeft   int
	firstClick  bool
	gFlag       int
	width       int
	height      int
	numBombs    int
	triggeredX  int
	triggeredY  int
	timeElapsed int
	timerActive bool
}

func newGameState(width, height, numBombs int) gameState {
	grid := make([][]cell, height)
	for y := range height {
		grid[y] = make([]cell, width)
	}
	return gameState{
		grid:        grid,
		status:      playing,
		bombsLeft:   numBombs,
		cellsLeft:   width*height - numBombs,
		firstClick:  true,
		width:       width,
		height:      height,
		numBombs:    numBombs,
		triggeredX:  -1,
		triggeredY:  -1,
		timeElapsed: 0,
		timerActive: false,
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
	case tickMsg:
		if m.status == playing && m.timerActive {
			m.timeElapsed++
			return m, tick()
		}
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
				var cmd tea.Cmd
				if m.grid[m.cursorY][m.cursorX].revealed {
					m.chord(m.cursorX, m.cursorY)
				} else {
					if m.firstClick {
						m.timerActive = true
						cmd = tick()
					}
					m.reveal(m.cursorX, m.cursorY)
				}
				return m, cmd
			} else {
				return newGameState(m.width, m.height, m.numBombs), nil
			}
		case "c":
			if m.status == playing {
				m.chord(m.cursorX, m.cursorY)
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
	type pos struct{ x, y int }
	var dangerZone []pos
	var safeZone []pos

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			if x == firstX && y == firstY {
				continue
			}
			if x >= firstX-1 && x <= firstX+1 && y >= firstY-1 && y <= firstY+1 {
				safeZone = append(safeZone, pos{x, y})
			} else {
				dangerZone = append(dangerZone, pos{x, y})
			}
		}
	}

	rand.Shuffle(len(dangerZone), func(i, j int) {
		dangerZone[i], dangerZone[j] = dangerZone[j], dangerZone[i]
	})
	rand.Shuffle(len(safeZone), func(i, j int) {
		safeZone[i], safeZone[j] = safeZone[j], safeZone[i]
	})

	allPotentials := append(dangerZone, safeZone...)
	allPotentials = append(allPotentials, pos{firstX, firstY})

	for i := 0; i < m.numBombs && i < len(allPotentials); i++ {
		p := allPotentials[i]
		m.grid[p.y][p.x].isBomb = true
	}

	for y := range m.height {
		for x := range m.width {
			if !m.grid[y][x].isBomb {
				m.grid[y][x].neighbors = m.countNeighbors(x, y)
			}
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
		m.timerActive = false
		m.triggeredX = x
		m.triggeredY = y
		for y := range m.height {
			for x := range m.width {
				m.grid[y][x].revealed = true
			}
		}
		return
	}

	m.cellsLeft--
	if m.cellsLeft == 0 {
		m.status = won
		m.timerActive = false
		for y := range m.height {
			for x := range m.width {
				if m.grid[y][x].isBomb == true {
					m.grid[y][x].flagged = true
				} else {
					m.grid[y][x].revealed = true
				}
			}
		}
		m.numBombs = 0
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

func (m *gameState) chord(x, y int) {
	if !m.grid[y][x].revealed || m.grid[y][x].neighbors == 0 {
		return
	}

	flagCount := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < m.width && ny >= 0 && ny < m.height && m.grid[ny][nx].flagged {
				flagCount++
			}
		}
	}

	if flagCount == m.grid[y][x].neighbors {
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
				if c.isBomb {//󰚑
					char = ""
					if x == m.triggeredX && y == m.triggeredY {
						char = ""
						style = triggeredBombStyle
					} else {
						style = bombStyle
					}
				} else if c.neighbors > 0 {
					char = fmt.Sprintf("%d", c.neighbors)
					style = numStyles[c.neighbors-1]
				} else {
					char = "·"
					style = hiddenStyle
				}
			} else if c.flagged {
				char = "⚑"
				style = flagStyle
			} else {
				char = "■"
				style = hiddenStyle
			}

			// We use a fixed width of 3 characters for each cell
			content := fmt.Sprintf(" %s ", char)
			if x == m.cursorX && y == m.cursorY {
				// Apply cursor style to the entire 3-character cell
				board.WriteString(cursorStyle.Render(content))
			} else {
				board.WriteString(style.Render(content))
			}
		}
		if y < m.height-1 {
			board.WriteString("\n")
		}
	}

	gameView := boardStyle.Render(board.String())

	// Status bar in Lualine style
	pinkLabel := lipgloss.NewStyle().Background(lipgloss.Color("212")).Foreground(lipgloss.Color("235")).Bold(true).Padding(0, 1)
	purpleLabel := lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("255")).Bold(true).Padding(0, 1)
	cyanLabel := lipgloss.NewStyle().Background(lipgloss.Color("81")).Foreground(lipgloss.Color("235")).Bold(true).Padding(0, 1)
	darkVal := lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("255")).Padding(0, 1)

	bombsStr := lipgloss.JoinHorizontal(lipgloss.Left,
		pinkLabel.Render("BOMBS"),
		darkVal.Render(fmt.Sprintf("%02d", m.bombsLeft)),
	)

	statusStr := lipgloss.JoinHorizontal(lipgloss.Left,
		purpleLabel.Render("STATUS"),
		darkVal.Render(strings.ToUpper(getStatusText(m))),
	)

	timeStr := lipgloss.JoinHorizontal(lipgloss.Left,
		cyanLabel.Render("TIME"),
		darkVal.Render(fmt.Sprintf("%03ds", m.timeElapsed)),
	)

	fullStatus := lipgloss.JoinHorizontal(lipgloss.Top,
		bombsStr, "  ", statusStr, "  ", timeStr,
	)

	var helpStr string
	if m.status != playing {
		if m.status == lost {
			helpStr = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("GAME OVER • Press R to Restart or Q to Quit")
		} else {
			helpStr = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render("YOU WIN • Press R to Restart or Q to Quit")
		}
	} else {
		helpStr = lipgloss.NewStyle().Foreground(lipgloss.Color("248")).Render("Arrows:Move • Spc:Reveal • F:Flag • C:Chord • R:Restart")
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		gameView,
		lipgloss.NewStyle().MarginTop(1).Render(fullStatus),
		lipgloss.NewStyle().MarginTop(1).Render(helpStr),
	)
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
	cursorStyle        = lipgloss.NewStyle().Background(lipgloss.Color("81")).Foreground(lipgloss.Color("255"))
	bombStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
	triggeredBombStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("9")).Bold(true)
	flagStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true) // Red flag
	numStyles          = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("33")),  // 1: light blue
		lipgloss.NewStyle().Foreground(lipgloss.Color("41")),  // 2: green
		lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // 3: red
		lipgloss.NewStyle().Foreground(lipgloss.Color("99")),  // 4: deep purple
		lipgloss.NewStyle().Foreground(lipgloss.Color("160")), // 5: maroon
		lipgloss.NewStyle().Foreground(lipgloss.Color("37")),  // 6: cyan
		lipgloss.NewStyle().Foreground(lipgloss.Color("248")), // 7: black (dark gray)
		lipgloss.NewStyle().Foreground(lipgloss.Color("243")), // 8: gray
	}
	hiddenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).MarginBottom(1)
	statusStyle = lipgloss.NewStyle().Italic(true).MarginTop(1)
	boardStyle  = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("248")).Padding(0, 1)
)
