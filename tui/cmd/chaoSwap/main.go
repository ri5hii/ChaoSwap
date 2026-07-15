package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	app "github.com/ri5hii/ChaoSwap/tui/app"
)

// main is the CLI entrypoint for the ChaoSwap binary.
//
// This command routes subcommands to either the Bubble Tea TUI (`start`) or to
// informational output (`help`, `description`). The TUI itself is implemented
// under `tui/app`.
func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		return
	}

	switch args[0] {
	case "description":
		printDescription()
	case "start":
		startGame()
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n\n", args[0])
		printHelp()
	}
}

// startGame launches the interactive Bubble Tea TUI.
func startGame() {
	program := tea.NewProgram(app.NewModel())
	_, err := program.Run()
	if err != nil {
		fmt.Println("Error running the game:", err)
	}
}

// printHelp prints CLI usage and available subcommands.
func printHelp() {
	fmt.Print(`
ChaoSwap:

Usage:
  chaoSwap <command>

Commands:
  start        Start the TUI game.
  description  Print a short description of the game.
  help         Print this help text.
`)
}

// printDescription prints a short, user-facing description of the game and its input conventions.
func printDescription() {
	fmt.Print(`
ChaoSwap: A chess variant where players can swap one of their pieces with an opponent's piece instead of making a normal move.

ChaoSwap provides a chess TUI with two modes:

  Normal: Standard chess rules.
  Chaos:  A twist mode (work in progress in this TUI).

TUI Tips:
  - Commands start with ':' (e.g. :help, :quit, :normal, :chaos, :description).
  - Moves use coordinate notation (e.g. e2e4).
  - Promotions should be specified with a suffix: e7e8q (or r/b/n).
`)
}
