package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

var (
	inputStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	modeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
)

func (m *Model) View() tea.View {
	banner := ""

	boardArea := ""

	inputLine := inputStyle.Render("> " + m.input + " ")
	modeLine := modeStyle.Render("Mode: " + m.mode)
	inputCumModeLine := fmt.Sprintf("%s | %s", inputLine, modeLine)
	statusLine := statusStyle.Render(m.status)
	helpLine := helpStyle.Render(m.helpLine)

	body := lipgloss.JoinHorizontal(lipgloss.Top, boardArea, "")

	screen := lipgloss.JoinVertical(lipgloss.Left, banner, body, inputCumModeLine, statusLine, helpLine)

	view := tea.NewView(screen)

	return view
}
