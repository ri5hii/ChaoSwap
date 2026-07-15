package app

import (
	"math/rand/v2"
	"time"

	tea "charm.land/bubbletea/v2"

	engineChaos "github.com/ri5hii/ChaoSwap/engine/engine-chaos"
	engineNormal "github.com/ri5hii/ChaoSwap/engine/engine-normal"
)

func NewModel() *Model {
	board := engineNormal.NewGamePosition()
	return &Model{
		chessBoard: board,
		chaos: &engineChaos.State{
			Base: board,
			RNG:  rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano()>>32))),
		},

		mode:     "Normal",
		input:    "",
		status:   "Welcome to ChaoSwap! Type :help for commands.",
		helpLine: "Press :chaos or :normal to switch mode. S is used to confirm swap in chaos mode",

		moveLog: []MoveRecord{},
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}
