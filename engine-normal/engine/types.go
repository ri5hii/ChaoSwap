package engine

type PieceColor int

const (
	White PieceColor = iota
	Black
)

type PieceType int

const (
	None PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

// Piece represents a chess piece with a type and a color.
type Piece struct {
	Type  PieceType
	Color PieceColor
}

// Square identifies a position on the board using 0-based file and rank coordinates.
// File 0 corresponds to 'a', file 7 to 'h'. Rank 0 corresponds to '1', rank 7 to '8'.
type Square struct {
	File int
	Rank int
}

// NoSquare represents an invalid square value.
var NoSquare = Square{File: -1, Rank: -1}

// IsValid reports whether the square lies within the bounds of an 8x8 board.
func (square Square) IsValid() bool {
	return square.File >= 0 && square.File <= 7 &&
		square.Rank >= 0 && square.Rank <= 7
}

// Equals reports whether two squares refer to the same file and rank.
func (square Square) Equals(other Square) bool {
	return square.File == other.File && square.Rank == other.Rank
}

// Board stores pieces indexed by [rank][file].
type Board [8][8]Piece
