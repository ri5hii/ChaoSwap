package engine_test

import (
	"testing"

	engine "github.com/ri5hii/ChaoSwap/engine-normal/engine"
)

func sq(file, rank int) engine.Square {
	return engine.Square{File: file, Rank: rank}
}

func placeKings(board *engine.BoardState) {
	board.SetPiece(sq(4, 7), engine.Piece{Type: engine.King, Color: engine.White})
	board.SetPiece(sq(4, 0), engine.Piece{Type: engine.King, Color: engine.Black})
}

func mustMakeMove(t *testing.T, board *engine.BoardState, from, to engine.Square) {
	t.Helper()
	piece := board.PieceAt(from)
	if piece.Type == engine.None {
		t.Fatalf("no piece at source square %v", from)
	}
	if board.MakeMove(piece, from, to) == nil {
		t.Fatalf("expected move to be valid from %v to %v", from, to)
	}
}

func mustRejectMove(t *testing.T, board *engine.BoardState, from, to engine.Square) {
	t.Helper()
	piece := board.PieceAt(from)
	if piece.Type == engine.None {
		t.Fatalf("no piece at source square %v", from)
	}
	if board.MakeMove(piece, from, to) != nil {
		t.Fatalf("expected move to be invalid from %v to %v", from, to)
	}
}

func TestInitialSetup(t *testing.T) {
	board := engine.NewGamePosition()

	whiteBack := []engine.PieceType{engine.Rook, engine.Knight, engine.Bishop, engine.Queen, engine.King, engine.Bishop, engine.Knight, engine.Rook}
	for file := 0; file < 8; file++ {
		square := sq(file, 7)
		piece := board.PieceAt(square)
		if piece.Type != whiteBack[file] || piece.Color != engine.White {
			t.Fatalf("expected white %v at file %d, rank 7, got %v %v", whiteBack[file], file, piece.Color, piece.Type)
		}
	}

	for file := 0; file < 8; file++ {
		square := sq(file, 6)
		piece := board.PieceAt(square)
		if piece.Type != engine.Pawn || piece.Color != engine.White {
			t.Fatalf("expected white pawn at file %d, rank 6, got %v %v", file, piece.Color, piece.Type)
		}
	}

	blackBack := []engine.PieceType{engine.Rook, engine.Knight, engine.Bishop, engine.Queen, engine.King, engine.Bishop, engine.Knight, engine.Rook}
	for file := 0; file < 8; file++ {
		square := sq(file, 0)
		piece := board.PieceAt(square)
		if piece.Type != blackBack[file] || piece.Color != engine.Black {
			t.Fatalf("expected black %v at file %d, rank 0, got %v %v", blackBack[file], file, piece.Color, piece.Type)
		}
	}

	for file := 0; file < 8; file++ {
		square := sq(file, 1)
		piece := board.PieceAt(square)
		if piece.Type != engine.Pawn || piece.Color != engine.Black {
			t.Fatalf("expected black pawn at file %d, rank 1, got %v %v", file, piece.Color, piece.Type)
		}
	}
}

func TestPawnMoves(t *testing.T) {
	board := engine.NewGamePosition()

	mustMakeMove(t, board, sq(0, 6), sq(0, 5))
	if board.PieceAt(sq(0, 5)).Type != engine.Pawn || board.PieceAt(sq(0, 5)).Color != engine.White {
		t.Fatalf("expected white pawn at destination after move")
	}
	if board.PieceAt(sq(0, 6)).Type != engine.None {
		t.Fatalf("expected source square to be empty after move")
	}

	board = engine.NewGamePosition()
	mustMakeMove(t, board, sq(1, 6), sq(1, 4))
}

func TestIllegalPawnBackward(t *testing.T) {
	board := engine.NewGamePosition()
	mustRejectMove(t, board, sq(0, 6), sq(0, 7))
}

func TestKnightMove(t *testing.T) {
	board := engine.NewGamePosition()
	mustMakeMove(t, board, sq(1, 7), sq(0, 5))
}

func TestBishopBlocked(t *testing.T) {
	board := engine.NewGamePosition()
	mustRejectMove(t, board, sq(2, 7), sq(6, 3))
}

func TestRookBlocked(t *testing.T) {
	board := engine.NewGamePosition()
	mustRejectMove(t, board, sq(0, 7), sq(0, 5))
}

func TestCaptureOwnPieceInvalid(t *testing.T) {
	board := engine.NewGamePosition()
	mustRejectMove(t, board, sq(0, 7), sq(0, 6))
}

func TestEnPassantWhite(t *testing.T) {
	board := engine.NewEmptyPosition()
	placeKings(board)

	whitePawnSquare := sq(4, 3)
	blackPawnSquare := sq(3, 3)
	board.SetPiece(whitePawnSquare, engine.Piece{Type: engine.Pawn, Color: engine.White})
	board.SetPiece(blackPawnSquare, engine.Piece{Type: engine.Pawn, Color: engine.Black})

	enPassantSquare := sq(3, 2)
	board.EnPassantSquare = enPassantSquare
	board.SideToMove = engine.White

	mustMakeMove(t, board, whitePawnSquare, enPassantSquare)

	if board.PieceAt(enPassantSquare).Type != engine.Pawn || board.PieceAt(enPassantSquare).Color != engine.White {
		t.Fatalf("expected white pawn to land on en passant target square")
	}
	if board.PieceAt(blackPawnSquare).Type != engine.None {
		t.Fatalf("expected black pawn to be captured via en passant")
	}
}

func TestCastlingWhiteKingSide(t *testing.T) {
	board := engine.NewEmptyPosition()
	board.SideToMove = engine.White
	board.CastlingRights = engine.WhiteKingSide | engine.WhiteQueenSide

	kingFrom := sq(4, 7)
	rookFrom := sq(7, 7)
	kingTo := sq(6, 7)
	rookTo := sq(5, 7)

	board.SetPiece(kingFrom, engine.Piece{Type: engine.King, Color: engine.White})
	board.SetPiece(rookFrom, engine.Piece{Type: engine.Rook, Color: engine.White})
	board.SetPiece(sq(4, 0), engine.Piece{Type: engine.King, Color: engine.Black})

	mustMakeMove(t, board, kingFrom, kingTo)

	if board.PieceAt(kingTo).Type != engine.King || board.PieceAt(kingTo).Color != engine.White {
		t.Fatalf("expected white king to be on g1 after castling")
	}
	if board.PieceAt(rookTo).Type != engine.Rook || board.PieceAt(rookTo).Color != engine.White {
		t.Fatalf("expected white rook to be on f1 after castling")
	}
	if board.PieceAt(rookFrom).Type != engine.None {
		t.Fatalf("expected original rook square to be empty after castling")
	}
}

func TestEnPassantInvalidWhenNoTarget(t *testing.T) {
	board := engine.NewEmptyPosition()
	placeKings(board)

	whitePawnSquare := sq(4, 3)
	blackPawnSquare := sq(3, 3)
	board.SetPiece(whitePawnSquare, engine.Piece{Type: engine.Pawn, Color: engine.White})
	board.SetPiece(blackPawnSquare, engine.Piece{Type: engine.Pawn, Color: engine.Black})
	board.SideToMove = engine.White
	board.EnPassantSquare = engine.NoSquare

	mustRejectMove(t, board, whitePawnSquare, sq(3, 2))
}

func TestCastlingBlockedByPiece(t *testing.T) {
	board := engine.NewEmptyPosition()
	board.SideToMove = engine.White
	board.CastlingRights = engine.WhiteKingSide | engine.WhiteQueenSide

	kingFrom := sq(4, 7)
	rookFrom := sq(7, 7)
	kingTo := sq(6, 7)

	board.SetPiece(kingFrom, engine.Piece{Type: engine.King, Color: engine.White})
	board.SetPiece(rookFrom, engine.Piece{Type: engine.Rook, Color: engine.White})
	board.SetPiece(sq(5, 7), engine.Piece{Type: engine.Knight, Color: engine.White})
	board.SetPiece(sq(4, 0), engine.Piece{Type: engine.King, Color: engine.Black})

	mustRejectMove(t, board, kingFrom, kingTo)
}

func TestCastlingThroughCheckInvalid(t *testing.T) {
	board := engine.NewEmptyPosition()
	board.SideToMove = engine.White
	board.CastlingRights = engine.WhiteKingSide | engine.WhiteQueenSide

	kingFrom := sq(4, 7)
	rookFrom := sq(7, 7)
	kingTo := sq(6, 7)

	board.SetPiece(kingFrom, engine.Piece{Type: engine.King, Color: engine.White})
	board.SetPiece(rookFrom, engine.Piece{Type: engine.Rook, Color: engine.White})
	board.SetPiece(sq(4, 0), engine.Piece{Type: engine.King, Color: engine.Black})

	attackerSquare := sq(5, 0)
	board.SetPiece(attackerSquare, engine.Piece{Type: engine.Rook, Color: engine.Black})

	mustRejectMove(t, board, kingFrom, kingTo)
}
