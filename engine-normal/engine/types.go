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

type Piece struct {
	Type  PieceType
	Color PieceColor
}

// Square holds a board position as explicit File and Rank fields.
// File 0 = a-file, File 7 = h-file.
// Rank 0 = black's back rank, Rank 7 = white's back rank.
type Square struct {
	File int
	Rank int
}

var NoSquare = Square{File: -1, Rank: -1}

func (square Square) IsValid() bool {
	return square.File >= 0 && square.File <= 7 &&
		square.Rank >= 0 && square.Rank <= 7
}

func (square Square) Equals(other Square) bool {
	return square.File == other.File && square.Rank == other.Rank
}

// Board is indexed as [rank][file].
// Rank 0 is black's back rank, Rank 7 is white's back rank.
type Board [8][8]Piece
