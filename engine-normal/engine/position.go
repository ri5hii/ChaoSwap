package engine

import "fmt"

type BoardState struct {
	ChessBoard      Board
	SideToMove      PieceColor
	CastlingRights  uint8
	EnPassantSquare Square
}

const (
	WhiteKingSide  uint8 = 1 << 0
	WhiteQueenSide uint8 = 1 << 1
	BlackKingSide  uint8 = 1 << 2
	BlackQueenSide uint8 = 1 << 3
)

func NewEmptyPosition() *BoardState {
	return &BoardState{
		SideToMove:      White,
		EnPassantSquare: NoSquare,
		CastlingRights:  WhiteKingSide | WhiteQueenSide | BlackKingSide | BlackQueenSide,
	}
}

func NewGamePosition() *BoardState {
	board := NewEmptyPosition()

	backRank := []PieceType{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}

	for file := 0; file <= 7; file++ {
		board.SetPiece(Square{file, 0}, Piece{Type: backRank[file], Color: Black})
		board.SetPiece(Square{file, 1}, Piece{Type: Pawn, Color: Black})
		board.SetPiece(Square{file, 6}, Piece{Type: Pawn, Color: White})
		board.SetPiece(Square{file, 7}, Piece{Type: backRank[file], Color: White})
	}

	return board
}

func (boardState *BoardState) SetPiece(square Square, piece Piece) {
	if !square.IsValid() {
		return
	}
	boardState.ChessBoard[square.Rank][square.File] = piece
}

func (boardState *BoardState) PieceAt(square Square) Piece {
	if !square.IsValid() {
		return Piece{Type: None}
	}
	return boardState.ChessBoard[square.Rank][square.File]
}

func PrintBoard(boardState *BoardState) {
	for rank := 0; rank <= 7; rank++ {
		fmt.Print(8-rank, " ")
		for file := 0; file <= 7; file++ {
			fmt.Print(boardState.ChessBoard[rank][file].PieceIcon(), " ")
		}
		fmt.Println()
	}
	fmt.Println("  a b c d e f g h")
}

func (piece Piece) PieceIcon() string {
	if piece.Type == None {
		return "."
	}
	if piece.Color == White {
		switch piece.Type {
		case Pawn:
			return "♙"
		case Rook:
			return "♖"
		case Knight:
			return "♘"
		case Bishop:
			return "♗"
		case Queen:
			return "♕"
		case King:
			return "♔"
		}
	}
	if piece.Color == Black {
		switch piece.Type {
		case Pawn:
			return "♟"
		case Rook:
			return "♜"
		case Knight:
			return "♞"
		case Bishop:
			return "♝"
		case Queen:
			return "♛"
		case King:
			return "♚"
		}
	}
	return "?"
}

func SquareNotation(square Square) string {
	if !square.IsValid() {
		return "--"
	}
	return string([]byte{
		byte('a' + square.File),
		byte('1' + square.Rank),
	})
}

func ParseSquareNotation(notation string) (Square, bool) {
	if len(notation) != 2 {
		return NoSquare, false
	}
	file := int(notation[0] - 'a')
	rank := int(notation[1] - '1')
	square := Square{File: file, Rank: rank}
	if !square.IsValid() {
		return NoSquare, false
	}
	return square, true
}
