package app

import (
	engineChaos "github.com/ri5hii/ChaoSwap/engine/engine-chaos"
	engineNormal "github.com/ri5hii/ChaoSwap/engine/engine-normal"
)

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

type MoveRecord struct {
	pieceType engineNormal.PieceType
	Side      engineNormal.PieceColor
	To        engineNormal.Square
	From      engineNormal.Square

	Ply int

	isCapture         bool
	isQueenSideCastle bool
	isKingSideCastle  bool

	isPromotion   bool
	PromotionType engineNormal.PieceType
}

type SummaryState struct{}

type Banner struct {
	gameBanner string
}
