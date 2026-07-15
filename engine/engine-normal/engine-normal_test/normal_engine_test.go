package engine_test

import (
	"strings"
	"testing"

	engine "github.com/ri5hii/ChaoSwap/engine/engine-normal"
)

func sq2(file, rank int) engine.Square { return engine.Square{File: file, Rank: rank} }

func cloneBoard(b *engine.BoardState) *engine.BoardState {
	nb := &engine.BoardState{
		SideToMove:      b.SideToMove,
		CastlingRights:  b.CastlingRights,
		EnPassantSquare: b.EnPassantSquare,
		GameState:       b.GameState,
	}
	nb.ChessBoard = b.ChessBoard
	return nb
}

func boardsEqual(a, b *engine.BoardState) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.SideToMove != b.SideToMove {
		return false
	}
	if a.CastlingRights != b.CastlingRights {
		return false
	}
	if !a.EnPassantSquare.Equals(b.EnPassantSquare) {
		return false
	}
	if a.GameState != b.GameState {
		return false
	}
	for r := 0; r < 8; r++ {
		for f := 0; f < 8; f++ {
			ap := a.ChessBoard[r][f]
			bp := b.ChessBoard[r][f]
			if ap.Type != bp.Type || ap.Color != bp.Color {
				return false
			}
		}
	}
	return true
}

func requireUndoRestores(t *testing.T, before, after *engine.BoardState, undo *engine.Undo) {
	t.Helper()
	engine.UndoMove(after, *undo)
	if !boardsEqual(before, after) {
		t.Fatalf("UndoMove did not restore board state")
	}
}

func mustTryMove(t *testing.T, b *engine.BoardState, from, to engine.Square, promo engine.PieceType) *engine.Undo {
	t.Helper()
	p := b.PieceAt(from)
	if p.Type == engine.None {
		t.Fatalf("no piece at %v", from)
	}
	u, err := b.TryMove(p, from, to, promo)
	if err != nil {
		t.Fatalf("expected TryMove to succeed %v->%v, got %v", from, to, err)
	}
	return u
}

func mustMakeMovePromo(t *testing.T, b *engine.BoardState, from, to engine.Square, promo engine.PieceType) *engine.Undo {
	t.Helper()
	p := b.PieceAt(from)
	if p.Type == engine.None {
		t.Fatalf("no piece at %v", from)
	}
	u, err := b.MakeMove(p, from, to, promo)
	if err != nil {
		t.Fatalf("expected MakeMove to succeed %v->%v, got %v", from, to, err)
	}
	return u
}

func mustRejectMakeMove(t *testing.T, b *engine.BoardState, from, to engine.Square, promo engine.PieceType) {
	t.Helper()
	p := b.PieceAt(from)
	if p.Type == engine.None {
		t.Fatalf("no piece at %v", from)
	}
	_, err := b.MakeMove(p, from, to, promo)
	if err == nil {
		t.Fatalf("expected MakeMove to reject %v->%v", from, to)
	}
}

func placeBareKings(b *engine.BoardState) {
	b.SetPiece(sq2(4, 7), engine.Piece{Type: engine.King, Color: engine.White})
	b.SetPiece(sq2(4, 0), engine.Piece{Type: engine.King, Color: engine.Black})
}

func TestTryMoveDoesNotEnforceKingSafetyButMakeMoveDoes(t *testing.T) {
	b := engine.NewEmptyPosition()
	placeBareKings(b)

	// Put a black rook giving check along the e-file to white king on e1 (engine coords e1 == (4,7)).
	b.SetPiece(sq2(4, 5), engine.Piece{Type: engine.Rook, Color: engine.Black})

	// White pawn move that does not resolve check.
	from := sq2(0, 6)
	to := sq2(0, 5)
	b.SetPiece(from, engine.Piece{Type: engine.Pawn, Color: engine.White})
	b.SideToMove = engine.White

	before := cloneBoard(b)

	// TryMove can succeed (piece-move rules only).
	u := mustTryMove(t, b, from, to, engine.None)
	// Undo restores.
	requireUndoRestores(t, before, b, u)

	// MakeMove must reject because king remains in check.
	mustRejectMakeMove(t, b, from, to, engine.None)

	// Ensure state unchanged after rejected MakeMove.
	if !boardsEqual(before, b) {
		t.Fatalf("rejected MakeMove mutated the board")
	}
}

func TestMakeMoveRejectsMovingPinnedPiece(t *testing.T) {
	b := engine.NewEmptyPosition()
	placeBareKings(b)

	// White king e1, white rook e2 pinned by black rook e8.
	wKing := sq2(4, 7)
	wRook := sq2(4, 6)
	bRook := sq2(4, 0)

	b.SetPiece(wKing, engine.Piece{Type: engine.King, Color: engine.White})
	b.SetPiece(wRook, engine.Piece{Type: engine.Rook, Color: engine.White})
	b.SetPiece(bRook, engine.Piece{Type: engine.Rook, Color: engine.Black})

	// Clear squares between e2 and e8 except e1 king and e2 rook already set; we have empty by default.
	b.SideToMove = engine.White

	// Attempt to move pinned rook away: e2 -> f2 should be illegal as it exposes king to rook.
	mustRejectMakeMove(t, b, wRook, sq2(5, 6), engine.None)
}

func TestUndoRestoresAfterCastlingKingSide(t *testing.T) {
	b := engine.NewEmptyPosition()
	b.SideToMove = engine.White
	b.CastlingRights = engine.WhiteKingSide | engine.WhiteQueenSide

	// White pieces for castling.
	kingFrom := sq2(4, 7) // e1
	rookFrom := sq2(7, 7) // h1
	kingTo := sq2(6, 7)   // g1

	b.SetPiece(kingFrom, engine.Piece{Type: engine.King, Color: engine.White})
	b.SetPiece(rookFrom, engine.Piece{Type: engine.Rook, Color: engine.White})
	b.SetPiece(sq2(4, 0), engine.Piece{Type: engine.King, Color: engine.Black})

	before := cloneBoard(b)
	u := mustMakeMovePromo(t, b, kingFrom, kingTo, engine.None)

	// Sanity: king moved.
	if got := b.PieceAt(kingTo); got.Type != engine.King || got.Color != engine.White {
		t.Fatalf("expected king on g1 after castling, got %+v", got)
	}

	requireUndoRestores(t, before, b, u)
}

func TestUndoRestoresAfterEnPassant(t *testing.T) {
	b := engine.NewEmptyPosition()
	placeBareKings(b)

	// White pawn on e5, black pawn on d5, EP target at d6? (engine ranks: white forward is -1).
	// Existing engine_test uses: white pawn (4,3) captures to (3,2) with EP target (3,2).
	whitePawnFrom := sq2(4, 3)
	blackPawnOn := sq2(3, 3)
	epTo := sq2(3, 2)

	b.SetPiece(whitePawnFrom, engine.Piece{Type: engine.Pawn, Color: engine.White})
	b.SetPiece(blackPawnOn, engine.Piece{Type: engine.Pawn, Color: engine.Black})
	b.EnPassantSquare = epTo
	b.SideToMove = engine.White

	before := cloneBoard(b)
	u := mustMakeMovePromo(t, b, whitePawnFrom, epTo, engine.None)

	// Sanity: pawn moved, captured pawn removed.
	if got := b.PieceAt(epTo); got.Type != engine.Pawn || got.Color != engine.White {
		t.Fatalf("expected white pawn on EP destination, got %+v", got)
	}
	if got := b.PieceAt(blackPawnOn); got.Type != engine.None {
		t.Fatalf("expected captured pawn removed, got %+v", got)
	}

	requireUndoRestores(t, before, b, u)
}

func TestUndoRestoresAfterPromotionToKnight(t *testing.T) {
	b := engine.NewEmptyPosition()
	placeBareKings(b)

	// White pawn one step from promotion. In this engine, white promotes on rank 0.
	from := sq2(0, 1)
	to := sq2(0, 0)

	b.SetPiece(from, engine.Piece{Type: engine.Pawn, Color: engine.White})
	b.SideToMove = engine.White

	before := cloneBoard(b)
	u := mustMakeMovePromo(t, b, from, to, engine.Knight)

	// Sanity: piece is now a knight at destination.
	if got := b.PieceAt(to); got.Type != engine.Knight || got.Color != engine.White {
		t.Fatalf("expected promoted knight at %v, got %+v", to, got)
	}
	if got := b.PieceAt(from); got.Type != engine.None {
		t.Fatalf("expected from square empty after promotion, got %+v", got)
	}

	requireUndoRestores(t, before, b, u)
}

func TestPromotionToRookBishopQueenAllWorkAndUndoRestores(t *testing.T) {
	cases := []struct {
		name  string
		promo engine.PieceType
	}{
		{"queen", engine.Queen},
		{"rook", engine.Rook},
		{"bishop", engine.Bishop},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := engine.NewEmptyPosition()
			placeBareKings(b)

			from := sq2(0, 1)
			to := sq2(0, 0)

			b.SetPiece(from, engine.Piece{Type: engine.Pawn, Color: engine.White})
			b.SideToMove = engine.White

			before := cloneBoard(b)
			u := mustMakeMovePromo(t, b, from, to, tc.promo)

			if got := b.PieceAt(to); got.Type != tc.promo || got.Color != engine.White {
				t.Fatalf("expected promoted %v at %v, got %+v", tc.promo, to, got)
			}

			requireUndoRestores(t, before, b, u)
		})
	}
}

func TestPromotionCapturePromotionUndo(t *testing.T) {
	b := engine.NewEmptyPosition()
	placeBareKings(b)

	// White pawn captures on last rank and promotes.
	from := sq2(0, 1)
	to := sq2(1, 0)

	b.SetPiece(from, engine.Piece{Type: engine.Pawn, Color: engine.White})
	b.SetPiece(to, engine.Piece{Type: engine.Rook, Color: engine.Black})
	b.SideToMove = engine.White

	before := cloneBoard(b)
	u := mustMakeMovePromo(t, b, from, to, engine.Queen)

	if got := b.PieceAt(to); got.Type != engine.Queen || got.Color != engine.White {
		t.Fatalf("expected promoted queen after capture-promotion, got %+v", got)
	}

	requireUndoRestores(t, before, b, u)
}

func TestIllegalKingMoveIntoCheckRejected(t *testing.T) {
	b := engine.NewEmptyPosition()

	// White king e1.
	wk := sq2(4, 7)
	b.SetPiece(wk, engine.Piece{Type: engine.King, Color: engine.White})
	// Black king somewhere safe.
	b.SetPiece(sq2(0, 0), engine.Piece{Type: engine.King, Color: engine.Black})
	// Black rook attacks e-file squares.
	b.SetPiece(sq2(4, 0), engine.Piece{Type: engine.Rook, Color: engine.Black})

	b.SideToMove = engine.White

	// King tries to move along file into rook attack (e1 -> e2).
	mustRejectMakeMove(t, b, wk, sq2(4, 6), engine.None)
}

func TestIsCheckMateAndIsStaleMateNotBothTrue(t *testing.T) {
	b := engine.NewEmptyPosition()
	b.SetPiece(sq2(7, 7), engine.Piece{Type: engine.King, Color: engine.White})
	b.SetPiece(sq2(0, 0), engine.Piece{Type: engine.King, Color: engine.Black})
	b.SetPiece(sq2(6, 7), engine.Piece{Type: engine.Rook, Color: engine.Black})
	b.SetPiece(sq2(7, 6), engine.Piece{Type: engine.Rook, Color: engine.Black})
	b.SetPiece(sq2(6, 6), engine.Piece{Type: engine.Queen, Color: engine.Black})
	b.SideToMove = engine.White

	mate, err := b.IsCheckMate()
	if err != nil {
		t.Fatalf("IsCheckMate error: %v", err)
	}
	stale, err := b.IsStaleMate()
	if err != nil {
		t.Fatalf("IsStaleMate error: %v", err)
	}
	if !mate {
		t.Fatalf("expected mate in this position")
	}
	if stale {
		t.Fatalf("expected not stalemate when checkmated")
	}
}

func TestIsStaleMateBasic(t *testing.T) {
	b := engine.NewEmptyPosition()

	// Classic stalemate: White king a1 trapped by black queen and king.
	b.SetPiece(sq2(0, 7), engine.Piece{Type: engine.King, Color: engine.White})
	b.SetPiece(sq2(7, 0), engine.Piece{Type: engine.King, Color: engine.Black})
	b.SetPiece(sq2(1, 5), engine.Piece{Type: engine.Queen, Color: engine.Black})
	b.SetPiece(sq2(2, 6), engine.Piece{Type: engine.King, Color: engine.Black})
	b.SideToMove = engine.White

	inCheck, err := b.InCheck(engine.White)
	if err != nil {
		t.Fatalf("InCheck error: %v", err)
	}
	if inCheck {
		t.Fatalf("expected not in check")
	}

	stale, err := b.IsStaleMate()
	if err != nil {
		t.Fatalf("IsStaleMate error: %v", err)
	}
	if !stale {
		t.Fatalf("expected stalemate")
	}

	mate, err := b.IsCheckMate()
	if err != nil {
		t.Fatalf("IsCheckMate error: %v", err)
	}
	if mate {
		t.Fatalf("expected not checkmate")
	}
}

func TestCastlingRightsClearedWhenRookMovesAndUndoRestores(t *testing.T) {
	b := engine.NewEmptyPosition()
	placeBareKings(b)

	// Place white rook on a1 and allow castling rights initially.
	rookFrom := sq2(0, 7)
	rookTo := sq2(0, 6)
	b.SetPiece(rookFrom, engine.Piece{Type: engine.Rook, Color: engine.White})
	b.CastlingRights = engine.WhiteKingSide | engine.WhiteQueenSide | engine.BlackKingSide | engine.BlackQueenSide
	b.SideToMove = engine.White

	before := cloneBoard(b)
	u := mustMakeMovePromo(t, b, rookFrom, rookTo, engine.None)

	// White queenside right should be cleared when rook moves from a1.
	if (b.CastlingRights & engine.WhiteQueenSide) != 0 {
		t.Fatalf("expected white queenside castling right to be cleared after rook move")
	}

	requireUndoRestores(t, before, b, u)
}

func TestEnPassantClearsWhenNotUsedNextPly(t *testing.T) {
	b := engine.NewGamePosition()

	// White: e2e4 (double push sets EP square)
	u1 := mustMakeMovePromo(t, b, sq2(4, 6), sq2(4, 4), engine.None)
	if !b.EnPassantSquare.IsValid() {
		t.Fatalf("expected en passant square set after double pawn push")
	}

	// Black plays a non-EP move: a7a6, which should clear EP target in MakeMove if unchanged.
	u2 := mustMakeMovePromo(t, b, sq2(0, 1), sq2(0, 2), engine.None)
	if b.EnPassantSquare.IsValid() {
		t.Fatalf("expected en passant square cleared after a non-EP response move")
	}

	// Undo back to verify undo stack correctness for EP square.
	before := engine.NewGamePosition()
	engine.UndoMove(b, *u2)
	engine.UndoMove(b, *u1)
	if !boardsEqual(before, b) {
		t.Fatalf("undo did not restore starting position")
	}
}

func TestCastlingIllegalWhileInCheck(t *testing.T) {
	b := engine.NewEmptyPosition()
	b.SideToMove = engine.White
	b.CastlingRights = engine.WhiteKingSide

	kingFrom := sq2(4, 7) // e1
	rookFrom := sq2(7, 7) // h1
	kingTo := sq2(6, 7)   // g1

	b.SetPiece(kingFrom, engine.Piece{Type: engine.King, Color: engine.White})
	b.SetPiece(rookFrom, engine.Piece{Type: engine.Rook, Color: engine.White})
	b.SetPiece(sq2(4, 0), engine.Piece{Type: engine.King, Color: engine.Black})

	// Put white king in check (rook on e-file).
	b.SetPiece(sq2(4, 5), engine.Piece{Type: engine.Rook, Color: engine.Black})

	mustRejectMakeMove(t, b, kingFrom, kingTo, engine.None)
}

func TestCastlingIllegalIntoCheck(t *testing.T) {
	b := engine.NewEmptyPosition()
	b.SideToMove = engine.White
	b.CastlingRights = engine.WhiteKingSide

	kingFrom := sq2(4, 7) // e1
	rookFrom := sq2(7, 7) // h1
	kingTo := sq2(6, 7)   // g1

	b.SetPiece(kingFrom, engine.Piece{Type: engine.King, Color: engine.White})
	b.SetPiece(rookFrom, engine.Piece{Type: engine.Rook, Color: engine.White})
	b.SetPiece(sq2(4, 0), engine.Piece{Type: engine.King, Color: engine.Black})

	// Attack the destination square g1 with a black rook on g8.
	b.SetPiece(sq2(6, 0), engine.Piece{Type: engine.Rook, Color: engine.Black})

	mustRejectMakeMove(t, b, kingFrom, kingTo, engine.None)
}

func TestCastlingQueenSideAndUndoRestores(t *testing.T) {
	b := engine.NewEmptyPosition()
	b.SideToMove = engine.White
	b.CastlingRights = engine.WhiteQueenSide

	kingFrom := sq2(4, 7) // e1
	rookFrom := sq2(0, 7) // a1
	kingTo := sq2(2, 7)   // c1

	b.SetPiece(kingFrom, engine.Piece{Type: engine.King, Color: engine.White})
	b.SetPiece(rookFrom, engine.Piece{Type: engine.Rook, Color: engine.White})
	b.SetPiece(sq2(4, 0), engine.Piece{Type: engine.King, Color: engine.Black})

	before := cloneBoard(b)
	u := mustMakeMovePromo(t, b, kingFrom, kingTo, engine.None)

	if got := b.PieceAt(kingTo); got.Type != engine.King || got.Color != engine.White {
		t.Fatalf("expected white king on c1 after queen-side castling, got %+v", got)
	}

	requireUndoRestores(t, before, b, u)
}

func TestCastlingRightsClearedWhenCornerRookCaptured(t *testing.T) {
	b := engine.NewEmptyPosition()
	placeBareKings(b)

	// White rook on h1 with rights; black bishop captures it.
	rookSq := sq2(7, 7) // h1
	b.SetPiece(rookSq, engine.Piece{Type: engine.Rook, Color: engine.White})
	b.CastlingRights = engine.WhiteKingSide | engine.WhiteQueenSide | engine.BlackKingSide | engine.BlackQueenSide

	// Put a black bishop on g2 that can capture h1 (diagonal).
	attacker := sq2(6, 6)
	b.SetPiece(attacker, engine.Piece{Type: engine.Bishop, Color: engine.Black})

	b.SideToMove = engine.Black

	u := mustMakeMovePromo(t, b, attacker, rookSq, engine.None)

	if (b.CastlingRights & engine.WhiteKingSide) != 0 {
		t.Fatalf("expected white kingside castling right cleared when rook on h1 is captured")
	}

	// Undo restores.
	engine.UndoMove(b, *u)
	if (b.CastlingRights & engine.WhiteKingSide) == 0 {
		// rights should be back to original after undo
		t.Fatalf("expected undo to restore castling rights after rook capture")
	}
}

func TestEnPassantCaptureThatExposesOwnKingIsRejected(t *testing.T) {
	b := engine.NewEmptyPosition()

	// Arrange a position where white pawn capturing en passant would open a rook attack to the king.
	//
	// White king on e1, black rook on e8 along the e-file. White pawn on e5 blocks the file.
	// EP target on d6 allows e5xd6 ep, which would vacate the e-file and expose the king to the rook.
	//
	// Using engine's coordinate system: king e1 == (4,7), black rook e8 == (4,0), pawn e5 == (4,3),
	// EP destination d6 == (3,2), captured pawn on d5 == (3,3).
	b.SetPiece(sq2(4, 7), engine.Piece{Type: engine.King, Color: engine.White})
	b.SetPiece(sq2(0, 0), engine.Piece{Type: engine.King, Color: engine.Black})
	b.SetPiece(sq2(4, 0), engine.Piece{Type: engine.Rook, Color: engine.Black})
	b.SetPiece(sq2(4, 3), engine.Piece{Type: engine.Pawn, Color: engine.White})
	b.SetPiece(sq2(3, 3), engine.Piece{Type: engine.Pawn, Color: engine.Black})
	b.EnPassantSquare = sq2(3, 2)
	b.SideToMove = engine.White

	// En passant capture should be rejected due to self-check exposure.
	mustRejectMakeMove(t, b, sq2(4, 3), sq2(3, 2), engine.None)
}

func TestParserIsMoveNotationValidCases(t *testing.T) {
	cases := []struct {
		name      string
		in        string
		wantOK    bool
		wantErr   bool
		wantPromo engine.PieceType
	}{
		{"simple move", "e2e4", true, false, engine.None},
		{"promotion queen", "e7e8q", true, false, engine.Queen},
		{"promotion rook", "e7e8r", true, false, engine.Rook},
		{"promotion bishop", "e7e8b", true, false, engine.Bishop},
		{"promotion knight", "e7e8n", true, false, engine.Knight},
		{"too short", "e2e", false, true, engine.None},
		{"bad square", "i2e4", false, true, engine.None},
		{"bad promo", "e7e8x", false, true, engine.None},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, from, to, promo, err := engine.IsMoveNotationValid(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.in, err)
			}
			if ok != tc.wantOK {
				t.Fatalf("ok mismatch for %q: got %v", tc.in, ok)
			}
			if !from.IsValid() || !to.IsValid() {
				t.Fatalf("expected valid squares for %q, got from=%v to=%v", tc.in, from, to)
			}
			if promo != tc.wantPromo {
				t.Fatalf("promotion mismatch for %q: got %v want %v", tc.in, promo, tc.wantPromo)
			}
		})
	}
}

func TestParserParseInputNormalization(t *testing.T) {
	got := engine.ParseInput("  E2-E4  ")
	if got != "e2e4" {
		t.Fatalf("expected ParseInput to normalize, got %q", got)
	}

	got = engine.ParseInput(" :He-lp ")
	if strings.TrimSpace(got) != ":help" && got != ":help" {
		// ParseInput is used by command pipeline; it should lowercase and remove spaces/hyphens.
		t.Fatalf("expected ParseInput to normalize command, got %q", got)
	}
}
