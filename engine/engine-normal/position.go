package engine

import "fmt"

// BoardState holds all mutable state required to validate and apply moves for a single game.
type BoardState struct {
	ChessBoard      Board
	SideToMove      PieceColor
	CastlingRights  uint8
	EnPassantSquare Square
	GameState       GameState
}

const (
	WhiteKingSide  uint8 = 1 << 0
	WhiteQueenSide uint8 = 1 << 1
	BlackKingSide  uint8 = 1 << 2
	BlackQueenSide uint8 = 1 << 3
)

// NewEmptyPosition returns a board state with an empty board and default rights/state for a new game.
func NewEmptyPosition() *BoardState {
	return &BoardState{
		SideToMove:      White,
		EnPassantSquare: NoSquare,
		CastlingRights:  WhiteKingSide | WhiteQueenSide | BlackKingSide | BlackQueenSide,
	}
}

// NewGamePosition returns a board state initialized to the standard chess starting position.
func NewGamePosition() *BoardState {
	board := NewEmptyPosition()

	backRank := []PieceType{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}
	for file := 0; file <= 7; file++ {
		board.SetPiece(Square{File: file, Rank: 0}, Piece{Type: backRank[file], Color: White})
		board.SetPiece(Square{File: file, Rank: 1}, Piece{Type: Pawn, Color: White})
		board.SetPiece(Square{File: file, Rank: 6}, Piece{Type: Pawn, Color: Black})
		board.SetPiece(Square{File: file, Rank: 7}, Piece{Type: backRank[file], Color: Black})
	}

	return board
}

// SetPiece places a piece on the given square.
// If the square is invalid, the operation is ignored.
//
// This is a low-level mutator used by move application and tests.
func (boardState *BoardState) SetPiece(square Square, piece Piece) {
	if !square.IsValid() {
		return
	}
	boardState.ChessBoard[square.Rank][square.File] = piece
}

// PieceAt returns the piece located at square.
// If the square is invalid, it returns a None piece.
func (boardState *BoardState) PieceAt(square Square) Piece {
	if !square.IsValid() {
		return Piece{Type: None}
	}
	return boardState.ChessBoard[square.Rank][square.File]
}

// PrintBoard prints an ASCII view of the board to stdout.
// This is intended for debugging, not for TUI rendering.
func PrintBoard(boardState *BoardState) {
	for rank := 0; rank <= 7; rank++ {
		fmt.Print(8-rank, " ")
		for file := 0; file <= 7; file++ {
			fmt.Print(PieceIcon(boardState.ChessBoard[rank][file]), " ")
		}
		fmt.Println()
	}
	fmt.Println("  a b c d e f g h")
}

// PieceIcon returns a single-character representation of a piece.
// It is used for simple console rendering and debugging.
func PieceIcon(piece Piece) string {
	if piece.Type == None {
		return "."
	}

	if piece.Color == White {
		switch piece.Type {
		case Pawn:
			return "p"
		case Rook:
			return "r"
		case Knight:
			return "n"
		case Bishop:
			return "b"
		case Queen:
			return "q"
		case King:
			return "k"
		}
	}

	if piece.Color == Black {
		switch piece.Type {
		case Pawn:
			return "P"
		case Rook:
			return "R"
		case Knight:
			return "N"
		case Bishop:
			return "B"
		case Queen:
			return "Q"
		case King:
			return "K"
		}
	}

	return "?"
}

// SquareNotation converts a square to coordinate notation like "e2".
// If the square is invalid, it returns "--".
func SquareNotation(square Square) string {
	if !square.IsValid() {
		return "--"
	}
	return string([]byte{
		byte('a' + square.File),
		byte('1' + square.Rank),
	})
}

// ParseNotationToSquare parses a coordinate like "e2" into a Square.
// It returns the parsed square, a boolean indicating validity, and an error for malformed input.
func ParseNotationToSquare(notation string) (Square, bool, error) {
	if len(notation) != 2 {
		return NoSquare, false, fmt.Errorf("Invalid square notation: %s", notation)
	}

	file := int(notation[0] - 'a')
	rank := int(notation[1] - '1')
	square := Square{File: file, Rank: rank}
	if !square.IsValid() {
		return NoSquare, false, fmt.Errorf("Invalid square notation: %s", notation)
	}

	return square, true, nil
}
