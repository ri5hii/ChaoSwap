package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	engine "github.com/ri5hii/ChaoSwap/engine/engine-normal"
)

// NewModel constructs the initial Bubble Tea model for a fresh game session.
//
// Why: the model is the single owner of UI state (input buffer, status text, mode)
// and the engine position pointer, so creating it in one place makes startup and
// resets consistent.
func NewModel() *Model {
	return &Model{
		board:    engine.NewGamePosition(),
		mode:     "Normal",
		input:    "",
		status:   "Welcome to ChaoSwap! Type :help for commands.",
		moveLog:  []MoveRecord{},
		quitting: false,
	}
}

// Init returns the initial Bubble Tea command.
//
// We don't need any startup I/O (timers, async loads, etc.), so this is nil.
func (model *Model) Init() tea.Cmd {
	return nil
}

// Update is the Bubble Tea reducer for the app.
//
// Inputs:
//   - Commands start with ':' (e.g. ':help', ':quit').
//   - Otherwise we treat the input as coordinate move notation (e.g. 'e2e4' or 'e7e8q').
//
// Promotion invariant:
// If the user enters a 4-char move that appears to be a pawn reaching the last rank,
// the UI switches into `awaitingPromotion` and the next Enter is interpreted as a
// promotion choice: 'q', 'r', 'b', or 'n'. This lets users type either `e7e8q`
// directly or `e7e8` then choose interactively.
func (model *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			model.quitting = true
			return model, tea.Quit

		case tea.KeyEnter:
			raw := strings.TrimSpace(model.input)
			parsed := engine.ParseInput(raw)

			if model.awaitingPromotion {
				return model.handlePromotionSelection(parsed)
			}

			if strings.HasPrefix(raw, ":") {
				return model.handleCommand(raw)
			}

			// Summary-mode invariant: once the game ends we switch into a read-only summary screen.
			// Moves are not accepted in summary mode (only commands like :quit or a future :new).
			if model.summary.Active {
				model.status = "Game over. You are in summary view. Use :quit to exit."
				model.input = ""
				return model, nil
			}

			return model.handleMoveInput(parsed)

		case tea.KeyBackspace:
			if len(model.input) > 0 {
				model.input = model.input[:len(model.input)-1]
			}
			return model, nil

		default:
			if len(msg.Runes) > 0 {
				model.input += string(msg.Runes)
				return model, nil
			}
		}
	}

	return model, nil
}

// handlePromotionSelection consumes a promotion choice while `awaitingPromotion` is true.
//
// It reconstructs the full move notation from the stored from/to squares plus the
// selected suffix, then applies it through the usual move pipeline.
//
// Commands:
//   - q/r/b/n: apply promotion
//   - cancel (or 'c'): exit the promotion flow without changing the engine state
func (model *Model) handlePromotionSelection(parsed string) (tea.Model, tea.Cmd) {
	if parsed == "" {
		model.status = "Choose promotion piece: q (queen), r (rook), b (bishop), n (knight)."
		model.input = ""
		return model, nil
	}

	switch parsed {
	case "q", "r", "b", "n":
		notation := fmt.Sprintf(
			"%s%s%s",
			engine.SquareNotation(model.promotionFrom),
			engine.SquareNotation(model.promotionTo),
			parsed,
		)

		err := model.ApplyMoveNotation(notation)
		if err != nil {
			model.status = err.Error()
			model.input = ""
			return model, nil
		}

		model.awaitingPromotion = false
		model.promotionFrom = engine.NoSquare
		model.promotionTo = engine.NoSquare
		model.input = ""
		return model, nil

	case "c", "cancel":
		model.awaitingPromotion = false
		model.promotionFrom = engine.NoSquare
		model.promotionTo = engine.NoSquare
		model.status = "Promotion cancelled. Enter the move again with a suffix, e.g. e7e8q."
		model.input = ""
		return model, nil

	default:
		model.status = "Invalid promotion choice. Use q/r/b/n (or 'cancel')."
		model.input = ""
		return model, nil
	}
}

// handleCommand executes ':'-prefixed user commands.
//
// Input is normalized via engine.ParseInput so users can type variants like
// ':Help', ': help', or ':he-lp'.
func (model *Model) handleCommand(raw string) (tea.Model, tea.Cmd) {
	cmd := strings.TrimSpace(strings.TrimPrefix(raw, ":"))
	cmd = engine.ParseInput(cmd)

	switch cmd {
	case "", "help", "h":
		model.status = `
Commands (prefix with ':'):
  :help | :h             Show this help
  :description | :d      Show game description
  :quit | :q             Quit
  :normal                Set Normal mode
  :chaos                 Set Chaos mode

Moves:
  e2e4                   Coordinate move input
  e7e8q / e7e8r / ...     Promotion suffix: q r b n

Notes:
  - Commands must start with ':'.
  - Promotion without a suffix will prompt you to choose q/r/b/n.
`
		model.input = ""
		return model, nil

	case "description", "d":
		model.status = `
Modes:
  Normal: Standard chess rules.
  Chaos:  (WIP in this TUI) A twist mode for swapping pieces (rules to be implemented).

Promotion:
  Use e7e8q (or r/b/n) to choose the promotion piece, or enter e7e8 and select when prompted.
`
		model.input = ""
		return model, nil

	case "quit", "q":
		model.quitting = true
		return model, tea.Quit

	case "normal":
		model.mode = "Normal"
		model.status = "Switched to Normal mode."
		model.input = ""
		return model, nil

	case "chaos":
		model.mode = "Chaos"
		model.status = "Switched to Chaos mode."
		model.input = ""
		return model, nil

	default:
		model.status = "Unknown command. Try :help."
		model.input = ""
		return model, nil
	}
}

// handleMoveInput applies non-command input as a move attempt.
//
// It also detects the "promotion without suffix" case and transitions into
// `awaitingPromotion` rather than immediately failing or auto-promoting.
func (model *Model) handleMoveInput(parsed string) (tea.Model, tea.Cmd) {
	if parsed == "" {
		model.input = ""
		return model, nil
	}

	if len(parsed) == 4 && model.PromotionInputWithoutSuffix(parsed) {
		_, fromSq, toSq, _, err := engine.IsMoveNotationValid(parsed)
		if err != nil {
			model.status = err.Error()
			model.input = ""
			return model, nil
		}

		model.awaitingPromotion = true
		model.promotionFrom = fromSq
		model.promotionTo = toSq
		model.status = "Promotion! Choose: q (queen), r (rook), b (bishop), n (knight). (or 'cancel')"
		model.input = ""
		return model, nil
	}

	if err := model.ApplyMoveNotation(parsed); err != nil {
		model.status = err.Error()
	}

	model.input = ""
	return model, nil
}

// ApplyMoveNotation validates and applies a coordinate move notation to the engine state.
//
// Responsibility split:
//   - validation/parsing is delegated to engine.IsMoveNotationValid
//   - move application is delegated to board.MakeMove
//   - UI bookkeeping (status + moveLog) happens here
//
// Log invariant:
// MoveRecord is populated with enough info to render later without needing to
// re-derive everything from historical positions.
func (model *Model) ApplyMoveNotation(notation string) error {
	valid, fromSquare, toSquare, promotionType, err := engine.IsMoveNotationValid(notation)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("invalid move notation: use e2e4 (or e7e8q for promotion)")
	}

	piece := model.board.PieceAt(fromSquare)
	target := model.board.PieceAt(toSquare)

	_, moveErr := model.board.MakeMove(piece, fromSquare, toSquare, promotionType)
	if moveErr != nil {
		return moveErr
	}

	ply := len(model.moveLog) + 1
	record := MoveRecord{
		Ply:       ply,
		Side:      piece.Color,
		PieceType: piece.Type,
		From:      fromSquare,
		To:        toSquare,
		IsCapture: target.Type != engine.None,
		Promotion: promotionType,
		Raw:       notation,
	}

	fileDiff := toSquare.File - fromSquare.File
	rankDiff := toSquare.Rank - fromSquare.Rank
	if piece.Type == engine.King && rankDiff == 0 && fileDiff == 2 {
		record.IsCastleKingSide = true
	}
	if piece.Type == engine.King && rankDiff == 0 && fileDiff == -2 {
		record.IsCastleQueenSide = true
	}

	model.status = "Played: " + notation
	model.moveLog = append(model.moveLog, record)

	// End-of-game detection is done after applying the move to the engine.
	checkmate, cmErr := model.board.IsCheckMate()
	if cmErr != nil {
		return cmErr
	}
	if checkmate {
		// In checkmate, SideToMove is the checkmated side (the side that cannot move).
		winner := "White"
		if model.board.SideToMove == engine.Black {
			winner = "Black"
		}

		model.summary.Active = true
		model.summary.Result = winner + " wins"
		model.summary.Reason = "Checkmate"
		model.summary.Cursor = len(model.moveLog)
		model.status = "Game over by checkmate."
		return nil
	}

	stalemate, smErr := model.board.IsStaleMate()
	if smErr != nil {
		return smErr
	}
	if stalemate {
		model.summary.Active = true
		model.summary.Result = "Draw"
		model.summary.Reason = "Stalemate"
		model.summary.Cursor = len(model.moveLog)
		model.status = "Game over by stalemate."
		return nil
	}

	// Draw detection: insufficient material.
	//
	// Note: your engine draw detection uses DrawByInsufficentMaterial (spelled as-is).
	if model.board.DrawByInsufficentMaterial() {
		model.summary.Active = true
		model.summary.Result = "Draw"
		model.summary.Reason = "Insufficient material"
		model.summary.Cursor = len(model.moveLog)
		model.status = "Game over by insufficient material."
		return nil
	}

	return nil
}

// PromotionInputWithoutSuffix reports whether a 4-character notation looks like a pawn
// move reaching the last rank, which requires a promotion choice.
//
// This is intentionally "UI heuristic" logic: the engine still performs the real
// legality checks when the move is applied.
func (model *Model) PromotionInputWithoutSuffix(notation string) bool {
	if len(notation) != 4 {
		return false
	}

	_, fromSquare, toSquare, _, err := engine.IsMoveNotationValid(notation)
	if err != nil {
		return false
	}

	piece := model.board.PieceAt(fromSquare)
	if piece.Type != engine.Pawn {
		return false
	}

	if piece.Color == engine.White && toSquare.Rank == 7 {
		return true
	}
	if piece.Color == engine.Black && toSquare.Rank == 0 {
		return true
	}

	return false
}
