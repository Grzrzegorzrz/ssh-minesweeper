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
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

const (
	host = "0.0.0.0"
	port = 2222
)

type appModel struct {
	setup setupModel
	game  gameState
	stage int // 0 = setup, 1 = game
}

func (a appModel) Init() tea.Cmd {
	return nil
}

func (a appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	}

	if a.stage == 0 {
		var cmd tea.Cmd
		a.setup, cmd = a.setup.Update(msg)
		if a.setup.done {
			a.game = newGameState(a.setup.width, a.setup.height, a.setup.bombs)
			a.stage = 1
		}
		return a, cmd
	}

	var cmd tea.Cmd
	a.game, cmd = a.game.Update(msg)
	return a, cmd
}

func (a appModel) View() string {
	if a.stage == 0 {
		return a.setup.View()
	}
	return a.game.View()
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
