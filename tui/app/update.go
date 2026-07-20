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
		switch {
		case msg.String() == "ctrl+c":
			return m, tea.Quit
		case msg.String() == "enter":
			raw := strings.TrimSpace(m.input)
			if raw == "" {
				return m, nil
			}
			// if strings.HasPrefix(raw, ":") {
			// 	return m.handleCommand(raw)
			// }
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
		default:
			if len(m.input) < 5 {
				m.input += msg.Text
			}
			return m, nil
		}
	}
	return m, nil
}
