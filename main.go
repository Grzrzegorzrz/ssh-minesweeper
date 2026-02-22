package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

var (
	windowStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("81")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(true)

	navStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("248")).
			Padding(1, 2)

	activeNavStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("81")).
			Padding(1, 2)
)

const (
	host = "0.0.0.0"
	port = 2222
)

type appModel struct {
	setup  setupModel
	game   gameState
	stage  int // 0 = setup, 1 = game
	width  int
	height int
}

func (a appModel) Init() tea.Cmd {
	return nil
}

func (a appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "q", "esc":
			if a.stage == 1 {
				a.stage = 0
				a.setup.done = false
				return a, nil
			} else {
				return a, tea.Quit
			}
		}
	}

	if a.stage == 0 {
		var cmd tea.Cmd
		a.setup, cmd = a.setup.Update(msg)
		if a.setup.done {
			a.game = newGameState(a.setup.width, a.setup.height, a.setup.bombs)
			a.stage = 1
			// Return a tick command to start the timer if firstClick is handled
			// Actually tick is started on first click in game.go
		}
		return a, cmd
	}

	var cmd tea.Cmd
	a.game, cmd = a.game.Update(msg)
	return a, cmd
}

func (a appModel) View() string {
	var content string

	if a.stage == 0 {
		content = a.setup.View()
	} else {
		content = a.game.View()
	}

	// App title
	title := headerStyle.Render("Minesweeper Via SSH!")

	fullView := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"\n",
		content,
	)

	// Wrap in window border
	windowed := windowStyle.Render(fullView)

	if a.width == 0 || a.height == 0 {
		return windowed
	}

	// Dynamic centering
	return lipgloss.Place(
		a.width,
		a.height,
		lipgloss.Center,
		lipgloss.Center,
		windowed,
	)
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	m := appModel{
		setup: initialSetupModel(),
		stage: 0,
	}
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, fmt.Sprintf("%d", port))),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			logging.Middleware(),
		),
	)
	if err != nil {
		fmt.Printf("Error creating server: %v", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	fmt.Printf("Starting SSH server on %s:%d\n", host, port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			fmt.Printf("Error starting server: %v", err)
			done <- nil
		}
	}()

	<-done
	fmt.Println("Stopping SSH server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		fmt.Printf("Error stopping server: %v", err)
	}
}
