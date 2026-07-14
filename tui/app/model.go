package app

import (
	engineChaos "github.com/ri5hii/ChaoSwap/engine/engine-chaos"
	engineNormal "github.com/ri5hii/ChaoSwap/engine/engine-normal"
)

// MoveRecord captures a single played ply in a UI-friendly form.
//
// It is intentionally separate from engine types so the TUI can render move history
// (including long algebraic formatting) without re-deriving facts from board history.
//
// Fields such as castling and promotion are stored explicitly so the log can be
// rendered consistently even if the engine state later changes (e.g., after undo).
type MoveRecord struct {
	Ply       int
	Side      engineNormal.PieceColor
	PieceType engineNormal.PieceType
	From      engineNormal.Square
	To        engineNormal.Square

	IsCapture bool

	IsCastleKingSide  bool
	IsCastleQueenSide bool

	Promotion engineNormal.PieceType
	Raw       string
}

// SummaryState stores end-of-game presentation state for the TUI.
//
// Why: after the game ends (checkmate/stalemate/draw), we want to present a stable
// summary and optionally allow navigation through the move history without mutating
// the live position.
type SummaryState struct {
	Active bool

	// Result is a short, user-facing statement (e.g., "1-0 Checkmate", "1/2-1/2 Stalemate").
	Result string

	// Reason provides a more descriptive explanation suitable for the status line.
	Reason string

	// Cursor is the move-log index currently selected for viewing (0..len(moveLog)).
	// When Cursor == len(moveLog), the view corresponds to the final position.
	Cursor int
}

// Model holds all state required by the Bubble Tea program.
type Model struct {
	board    *engineNormal.BoardState
	chaos    *engineChaos.State
	mode     string
	input    string
	status   string
	moveLog  []MoveRecord
	quitting bool

	awaitingPromotion bool
	promotionFrom     engineNormal.Square
	promotionTo       engineNormal.Square

	// summary holds UI state for presenting a post-game summary and history navigation.
	summary SummaryState
}
