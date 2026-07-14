package app

import (
	"fmt"
	"strings"

	engineNormal "github.com/ri5hii/ChaoSwap/engine/engine-normal"
)

// View renders the full screen for the Bubble Tea TUI.
//
// Layout (top to bottom):
//   - Board (left) and move log (right), rendered side-by-side.
//   - Current turn and selected mode.
//   - Input prompt showing the current buffer.
//   - Status line for errors and informational messages.
//
// Summary mode:
// If `model.summary.Active` is true, View switches to a summary screen that displays the
// final result and the move list without accepting further moves.
func (model *Model) View() string {
	if model != nil && model.summary.Active {
		return model.viewSummary()
	}

	boardLines := renderBoard(model.board)
	logLines := renderMoveLog(model.moveLog, 20)

	top := BoardSideByMoveSide(boardLines, logLines, "   ")

	turnLine := fmt.Sprintf("Turn: %s | Mode: %s", sideToMoveString(model.board), model.mode)
	inputLine := fmt.Sprintf("> Enter Move/Command: %s", model.input)
	statusLine := fmt.Sprintf("Status: %s", model.status)

	parts := []string{
		strings.Join(top, "\n"),
		turnLine,
		inputLine,
		statusLine,
	}

	return strings.Join(parts, "\n") + "\n"
}

// viewSummary renders an end-of-game summary screen.
//
// It reuses the standard board + move log layout, and adds a result banner plus a short
// instruction footer. Navigation through historical positions is intentionally not
// implemented here yet; the summary is a stable, read-only view.
func (model *Model) viewSummary() string {
	boardLines := renderBoard(model.board)
	logLines := renderMoveLog(model.moveLog, 40)

	top := BoardSideByMoveSide(boardLines, logLines, "   ")

	resultLine := "Result: (unknown)"
	if model.summary.Result != "" {
		resultLine = "Result: " + model.summary.Result
	}

	reasonLine := ""
	if model.summary.Reason != "" {
		reasonLine = "Reason: " + model.summary.Reason
	}

	turnLine := fmt.Sprintf("Final turn (would be): %s | Mode: %s", sideToMoveString(model.board), model.mode)

	helpLine := "Summary view. Use :new to play again, :quit to exit."
	inputLine := fmt.Sprintf("> Enter Command: %s", model.input)
	parts := []string{
		strings.Join(top, "\n"),
		resultLine,
		reasonLine,
		turnLine,
		helpLine,
		inputLine,
	}

	// Avoid an empty "Reason:" line when no reason was set.
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			continue
		}
		out = append(out, p)
	}

	return strings.Join(out, "\n") + "\n"
}

// sideToMoveString turns engine turn state into a user-facing label.
func sideToMoveString(board *engineNormal.BoardState) string {
	if board == nil {
		return "?"
	}
	if board.SideToMove == engineNormal.White {
		return "White"
	}
	return "Black"
}

// renderBoard returns an ASCII board suitable for terminal display.
//
// Invariant:
//   - The engine board is indexed [rank][file], where rank increases from '1' to '8'.
//   - This renderer prints ranks from 8 down to 1, matching standard chess diagrams.
func renderBoard(board *engineNormal.BoardState) []string {
	if board == nil {
		return []string{
			"8 . . . . . . . .",
			"7 . . . . . . . .",
			"6 . . . . . . . .",
			"5 . . . . . . . .",
			"4 . . . . . . . .",
			"3 . . . . . . . .",
			"2 . . . . . . . .",
			"1 . . . . . . . .",
			"  a b c d e f g h",
		}
	}

	lines := make([]string, 0, 9)

	for rank := 7; rank >= 0; rank-- {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("%d ", rank+1))

		for file := 0; file <= 7; file++ {
			p := board.ChessBoard[rank][file]
			b.WriteString(engineNormal.PieceIcon(p))
			b.WriteString(" ")
		}
		lines = append(lines, b.String())
	}

	lines = append(lines, "  a b c d e f g h")
	return lines
}

// renderMoveLog formats a list of plies into full-move lines ("1. ...").
//
// Why this exists:
//   - The engine is concerned with legality and state transitions.
//   - The TUI stores a UI-friendly `MoveRecord` so it can render history without
//     re-deriving details after the board changes.
//
// `max` limits how many plies are shown (0 or negative means show all).
func renderMoveLog(moves []MoveRecord, max int) []string {
	lines := []string{"Moves:"}

	if len(moves) == 0 {
		lines = append(lines, "(none)")
		return lines
	}

	start := 0
	if max > 0 && len(moves) > max {
		start = len(moves) - max
	}

	visible := moves[start:]
	if len(visible) == 0 {
		lines = append(lines, "(none)")
		return lines
	}

	// Keep the log aligned to "White, Black" pairs. If the first visible ply is
	// Black, drop it rather than printing a partial line.
	if len(visible) > 0 && visible[0].Side == engineNormal.Black {
		visible = visible[1:]
	}

	for i := 0; i < len(visible); i += 2 {
		white := visible[i]
		black := MoveRecord{}
		hasBlack := i+1 < len(visible)
		if hasBlack {
			black = visible[i+1]
		}

		moveNumber := (white.Ply + 1) / 2
		wStr := formatLongAlgebraic(white)
		bStr := ""
		if hasBlack {
			bStr = formatLongAlgebraic(black)
		}

		if bStr == "" {
			lines = append(lines, fmt.Sprintf("%2d. %-10s", moveNumber, wStr))
		} else {
			lines = append(lines, fmt.Sprintf("%2d. %-10s %-10s", moveNumber, wStr, bStr))
		}
	}

	return lines
}

// formatLongAlgebraic renders a single `MoveRecord` using a long-algebraic-like format.
//
// The output is intentionally simple and stable for a TUI move list:
//   - Castling: "O-O" or "O-O-O"
//   - Otherwise: "<piece><from>< - or x ><to>[=<promotionPiece>]"
func formatLongAlgebraic(m MoveRecord) string {
	if len(m.Raw) >= 2 && m.Raw[:2] == "S(" {
		return m.Raw
	}

	if m.IsCastleKingSide {
		return "O-O"
	}
	if m.IsCastleQueenSide {
		return "O-O-O"
	}

	pieceIcon := engineNormal.PieceIcon(engineNormal.Piece{Type: m.PieceType, Color: m.Side})
	from := engineNormal.SquareNotation(m.From)
	to := engineNormal.SquareNotation(m.To)

	sep := "-"
	if m.IsCapture {
		sep = "x"
	}

	s := pieceIcon + from + sep + to

	if m.Promotion != engineNormal.None {
		s += "=" + engineNormal.PieceIcon(engineNormal.Piece{Type: m.Promotion, Color: m.Side})
	}

	return s
}

// BoardSideByMoveSide stitches the board and move log into a side-by-side layout.
//
// It pads the left block to a uniform width (computed by rune count, not bytes)
// so the right block aligns cleanly even with non-ASCII characters.
func BoardSideByMoveSide(left []string, right []string, gap string) []string {
	if gap == "" {
		gap = "|"
	}

	leftWidth := 0
	for _, l := range left {
		if w := runeWidth(l); w > leftWidth {
			leftWidth = w
		}
	}

	height := len(left)
	if len(right) > height {
		height = len(right)
	}

	out := make([]string, 0, height)
	for i := 0; i < height; i++ {
		var l, r string
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}

		leftPadded := padRight(l, leftWidth)
		if r == "" {
			out = append(out, leftPadded)
		} else {
			out = append(out, leftPadded+gap+r)
		}
	}

	return out
}

// padRight pads a string using spaces until it reaches the desired rune width.
func padRight(s string, width int) string {
	w := runeWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// runeWidth returns the number of runes (not bytes) in s for basic terminal alignment.
func runeWidth(s string) int {
	return len([]rune(s))
}
