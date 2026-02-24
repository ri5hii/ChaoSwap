package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	app "github.com/ri5hii/ChaoSwap/tui/app"
)

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

func startGame() {
	program := tea.NewProgram(app.NewModel())
	_, err := program.Run()
	if err != nil {
		fmt.Println("Error running the game:", err)
	}
}

func printHelp() {
	fmt.Print(`ChaoSwap

Usage:
  chaoSwap <command>

Commands:
  start        Start the TUI game.
  description  Print a short description of the game.
  help         Print this help text.
`)
}

func printDescription() {
	fmt.Print(`ChaoSwap

ChaoSwap provides a chess TUI with two modes:

  Normal: Standard chess rules.
  Chaos:  A twist mode (work in progress in this TUI).

TUI Tips:
  - Commands start with ':' (e.g. :help, :quit, :normal, :chaos, :description).
  - Moves use coordinate notation (e.g. e2e4).
  - Promotions can be specified with a suffix: e7e8q (or r/b/n).
`)
}
