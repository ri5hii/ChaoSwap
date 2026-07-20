package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	engineNormal "github.com/ri5hii/ChaoSwap/engine/engine-normal"
)

var (
	inputStyle  = lipgloss.NewStyle().Foreground(lipgloss.White)
	modeStyle   = lipgloss.NewStyle().Foreground(lipgloss.White)
	logStyle    = lipgloss.NewStyle().Foreground(lipgloss.White)
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.White)
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.White)

	boardPad = lipgloss.NewStyle().PaddingRight(4)

	placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	activeStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	hasInput = false
)

func (m *Model) View() tea.View {
	banner := ` ██████╗██╗  ██╗ █████╗  ██████╗ ███████╗██╗    ██╗ █████╗ ██████╗ 
██╔════╝██║  ██║██╔══██╗██╔═══██╗██╔════╝██║    ██║██╔══██╗██╔══██╗
██║     ███████║███████║██║   ██║███████╗██║ █╗ ██║███████║██████╔╝
██║     ██╔══██║██╔══██║██║   ██║╚════██║██║███╗██║██╔══██║██╔═══╝ 
╚██████╗██║  ██║██║  ██║╚██████╔╝███████║╚███╔███╔╝██║  ██║██║     
 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝ ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝     `

	boardArea := renderBoard(m.chessBoard)
	boardArea = boardPad.Render(boardArea)

	MoveLog := renderMoveLog(m.moveLog)

	inputLine := inputStyle.Render("> Enter your move: " + m.input + " ")
	modeLine := modeStyle.Render(fmt.Sprintf("Turn: %s | Mode: %s", sideToMoveString(m.chessBoard), m.mode))
	inputCumModeLine := fmt.Sprintf("%s | %s", inputLine, modeLine)
	statusLine := statusStyle.Render(m.status)
	helpLine := helpStyle.Render(m.helpLine)

	body := lipgloss.JoinHorizontal(lipgloss.Top, boardArea, MoveLog)

	screen := lipgloss.JoinVertical(lipgloss.Left, banner, "", body, "", inputCumModeLine, statusLine, helpLine)

	view := tea.NewView(screen)
	view.BackgroundColor = lipgloss.Color("0")

	cursorY := 41
	cursorX := len("> Enter your move: " + m.input)
	view.Cursor = tea.NewCursor(cursorX, cursorY)
	view.Cursor.Blink = true

	return view
}

func sideToMoveString(board *engineNormal.BoardState) string {
	if board == nil {
		return "?"
	}
	if board.SideToMove == engineNormal.White {
		return "White"
	}
	return "Black"
}

func renderMoveLog(moves []MoveRecord) string {
	lines := []string{"Moves: "}
	if len(moves) == 0 {
		lines = append(lines, "(none)")
	} else {
		for i := 0; i < len(moves); i += 2 {
			move := moves[i]
			line := fmt.Sprintf("%2d. %-7s", (move.Ply+1)/2, formatMoveRecord(move))
			if i+1 < len(moves) {
				line += fmt.Sprintf("\t %-7s", formatMoveRecord(moves[i+1]))
			}
			lines = append(lines, line)
		}
	}
	for len(lines) < 34 {
		lines = append(lines, "")
	}
	return logStyle.Render(strings.Join(lines, "\n"))
}

func formatMoveRecord(m MoveRecord) string {
	if m.isSwap {
		return fmt.Sprintf("S(%s -> %s)", engineNormal.SquareNotation(m.From), engineNormal.SquareNotation(m.To))
	}
	return fmt.Sprintf("%c%c -> %c%c", m.From.File+'a', m.From.Rank+'1', m.To.File+'a', m.To.Rank+'1')
}

const cellW = 7

var (
	topBorder = buildBorder('┌', '┬', '┐')
	midBorder = buildBorder('├', '┼', '┤')
	botBorder = buildBorder('└', '┴', '┘')
)

func buildBorder(left, mid, right rune) string {
	var b strings.Builder
	b.WriteRune(left)
	for i := 0; i < 8; i++ {
		for j := 0; j < cellW; j++ {
			b.WriteRune('─')
		}
		if i < 7 {
			b.WriteRune(mid)
		}
	}
	b.WriteRune(right)
	return b.String()
}

func spacerRow() string {
	var b strings.Builder
	b.WriteRune('│')
	for range 8 {
		b.WriteString(strings.Repeat(" ", cellW))
		b.WriteRune('│')
	}
	return b.String()
}

func pieceRow(board *engineNormal.BoardState, rank int) string {
	var b strings.Builder
	b.WriteRune('│')
	for file := 0; file < 8; file++ {
		b.WriteString(pieceCell(board.ChessBoard[rank][file]))
		b.WriteRune('│')
	}
	return b.String()
}

func pieceCell(piece engineNormal.Piece) string {
	s := engineNormal.PieceIcon(piece)
	if s == "." {
		return strings.Repeat(" ", cellW)
	}
	pad := (cellW - 1) / 2 // 3 for cellW=7
	return strings.Repeat(" ", pad) + s + strings.Repeat(" ", cellW-1-pad)
}

func fileLabelRow() string {
	var b strings.Builder
	b.WriteString("    ") // align past rank prefix area
	for file := 0; file < 8; file++ {
		pad := (cellW - 1) / 2
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteByte(byte('a' + file))
		b.WriteString(strings.Repeat(" ", cellW-pad))
	}
	return b.String()
}

func renderBoard(board *engineNormal.BoardState) string {
	var b strings.Builder
	b.Grow(68 * 33)

	b.WriteString("   " + topBorder + "\n")

	for rank := 7; rank >= 0; rank-- {
		b.WriteString("   " + spacerRow() + "\n")
		b.WriteString(fmt.Sprintf(" %d ", rank+1))
		b.WriteString(pieceRow(board, rank))
		b.WriteByte('\n')
		b.WriteString("   " + spacerRow() + "\n")

		if rank > 0 {
			b.WriteString("   " + midBorder + "\n")
		} else {
			b.WriteString("   " + botBorder + "\n")
		}
	}

	b.WriteString(fileLabelRow())
	return b.String()
}
