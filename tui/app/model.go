package app

import (
	engineChaos "github.com/ri5hii/ChaoSwap/engine/engine-chaos"
	engineNormal "github.com/ri5hii/ChaoSwap/engine/engine-normal"
)

// Model holds all UI state for the ChaoSwap TUI.
type Model struct {
	chessBoard *engineNormal.BoardState
	chaos      *engineChaos.State
	mode       string
	input      string
	status     string
	helpLine   string

	moveLog []MoveRecord

	summary SummaryState
}

// MoveRecord stores a single move or swap entry for the move log.
type MoveRecord struct {
	pieceType engineNormal.PieceType
	Side      engineNormal.PieceColor
	To        engineNormal.Square
	From      engineNormal.Square

	isSwap bool

	Ply int

	isCapture         bool
	isQueenSideCastle bool
	isKingSideCastle  bool

	isPromotion   bool
	PromotionType engineNormal.PieceType
}

// SummaryState holds end-of-game information for the summary screen. Currently unused.
type SummaryState struct{}

// Banner holds the ASCII art banner displayed at the top of the screen. Currently unused.
type Banner struct {
	gameBanner string
}
