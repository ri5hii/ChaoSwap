package app

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	engineChaos "github.com/ri5hii/ChaoSwap/engine/engine-chaos"
	engineNormal "github.com/ri5hii/ChaoSwap/engine/engine-normal"
)

// NewModel initialises a Model with a standard starting position, seeded RNG, default mode, and welcome status.
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
		helpLine: "Press <tab> to switch mode. S is used to confirm swap in chaos mode",

		moveLog: []MoveRecord{},
	}
}

// Init satisfies tea.Model; no initial command is needed.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles key presses for mode switching, move entry, and chaos swap execution.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case msg.String() == "ctrl+c":
			return m, tea.Quit
		case msg.String() == "tab":
			if m.mode == "Chaos" {
				m.mode = "Normal"
				m.status = "Switched to Normal mode."
			} else {
				m.mode = "Chaos"
				m.status = "Switched to Chaos mode."
			}
			return m, nil
		case msg.String() == "enter":
			raw := strings.TrimSpace(m.input)
			if raw == "" {
				return m, nil
			}
			valid, from, to, promo, err := engineNormal.IsMoveNotationValid(raw)
			if err != nil {
				m.status = err.Error()
				return m, nil
			}
			if !valid {
				m.status = "Invalid move notation."
				return m, nil
			}

			piece := m.chessBoard.PieceAt(from)
			var movePlayed string
			if promo != engineNormal.None {
				movePlayed = fmt.Sprintf("(%s -> %s)%s", engineNormal.SquareNotation(from), engineNormal.SquareNotation(to), engineNormal.PieceIcon(engineNormal.Piece{
					Type:  promo,
					Color: piece.Color,
				}))
			} else {
				movePlayed = fmt.Sprintf("(%s -> %s)", engineNormal.SquareNotation(from), engineNormal.SquareNotation(to))
			}

			_, moveErr := m.chessBoard.MakeMove(piece, from, to, promo)
			if moveErr != nil {
				m.status = "Error: " + moveErr.Error()
			} else {
				m.status = "Played: " + movePlayed
				m.moveLog = append(m.moveLog, MoveRecord{
					Ply:       len(m.moveLog) + 1,
					Side:      piece.Color,
					pieceType: piece.Type,
					From:      from,
					To:        to,
				})
			}
			m.input = ""
			return m, nil
		case msg.String() == "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil
		case msg.String() == "s", msg.String() == "S":
			if m.mode == "Chaos" {
				return m.handleChaoSwap()
			}
			m.input += msg.Text
			return m, nil
		default:
			if m.mode == "Normal" {
				if len(m.input) < 5 {
					m.input += msg.Text
				}
				return m, nil
			} else {
				m.status = "Switch to Normal mode to apply move."
			}
		}
	}
	return m, nil
}

// handleChaoSwap picks a random legal swap and applies it, logging the result.
func (m *Model) handleChaoSwap() (tea.Model, tea.Cmd) {
	swap, ok := m.chaos.PickSwapPair()
	if !ok {
		m.status = "No legal swap available."
		return m, nil
	}

	m.chaos.TrySwap(swap)

	m.moveLog = append(m.moveLog, MoveRecord{
		isSwap: true,
		Ply:    len(m.moveLog) + 1,
		Side:   1 - m.chessBoard.SideToMove,
		From:   swap.A,
		To:     swap.B,
	})
	m.status = fmt.Sprintf("Swap: %s -> %s", engineNormal.SquareNotation(swap.A), engineNormal.SquareNotation(swap.B))
	return m, nil
}
