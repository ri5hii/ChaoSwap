package main

import (
	"fmt"
	"os"

	"github.com/ri5hii/ChaoSwap/engine-normal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

func printDescription() {
	fmt.Print(`ChaoSwap: 
		
		ChaoSwap offers a unique twist on traditional chess, providing players with new opportunities for creativity and tactical play.
		
		Rules:
		1. The game is played on a standard 8x8 chessboard with the same pieces and initial setup as traditional chess.
		2. Players take turns making moves, just like in regular chess.
		3. However, instead of moving a piece to an empty square, players can choose to swap the positions of two random pieces of their own color.
		4. The objective of the game remains the same: to checkmate the opponent's king.
		
		Modes:
		1. Normal Mode: Players proceed to play the game as per the rules of traditional chess.
		2. Chaos Mode: At the start of each turn, the player can choose to swap two random pieces of their own color instead of making a regular move.`)
}

func printHelp() {
	fmt.Print(`ChaoSwap: 
		
		Usage:
		ChaoSwap [command]
		
		Commands:
		description - Print a description of the game and its rules.
		start       - Start the game.
		help        - Print this help message.`)
}

func startGame() {
	// Create and run bubble tea model for the game here

}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		return
	}

	cmd := args[0]
	switch cmd {
	case "description":
		printDescription()
	case "start":
		startGame()
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s \n", cmd)
		printDescription()
		fmt.Print("\n")
		printHelp()
	}

}
