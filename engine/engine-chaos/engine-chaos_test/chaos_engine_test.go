package engine_test

import (
	"math/rand/v2"
	"testing"

	chaos "github.com/ri5hii/ChaoSwap/engine/engine-chaos"
	normal "github.com/ri5hii/ChaoSwap/engine/engine-normal"
)

func sq2(file, rank int) normal.Square { return normal.Square{File: file, Rank: rank} }

func newStateWithBoard(b *normal.BoardState) *chaos.State {
	return &chaos.State{
		Base: b,
		RNG:  rand.New(rand.NewPCG(42, 42)),
	}
}

func placeBareKings(b *normal.BoardState) {
	b.SetPiece(sq2(4, 0), normal.Piece{Type: normal.King, Color: normal.White})
	b.SetPiece(sq2(4, 7), normal.Piece{Type: normal.King, Color: normal.Black})
}

func boardsEqual(a, b *normal.BoardState) bool {
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

// Swap two white knights b1<->g1 in standard position — should be legal.
func TestIsLegalSwap_ValidSwap(t *testing.T) {
	b := normal.NewGamePosition()
	s := newStateWithBoard(b)

	swap := chaos.Swap{A: sq2(1, 0), B: sq2(6, 0)}
	ok, err := s.IsLegalSwap(swap)
	if err != nil {
		t.Fatalf("IsLegalSwap error: %v", err)
	}
	if !ok {
		t.Fatalf("expected swap of white knights b1<->g1 to be legal")
	}
}

// One square has no piece — swap should be rejected.
func TestIsLegalSwap_EmptySquare(t *testing.T) {
	b := normal.NewEmptyPosition()
	placeBareKings(b)
	b.SetPiece(sq2(0, 1), normal.Piece{Type: normal.Pawn, Color: normal.White})
	b.SideToMove = normal.White
	s := newStateWithBoard(b)

	swap := chaos.Swap{A: sq2(0, 1), B: sq2(0, 2)}
	ok, err := s.IsLegalSwap(swap)
	if err != nil {
		t.Fatalf("IsLegalSwap error: %v", err)
	}
	if ok {
		t.Fatalf("expected swap with empty square to be rejected")
	}
}

// Both squares are the same — swap should be rejected.
func TestIsLegalSwap_SameSquare(t *testing.T) {
	b := normal.NewGamePosition()
	s := newStateWithBoard(b)

	swap := chaos.Swap{A: sq2(1, 0), B: sq2(1, 0)}
	ok, err := s.IsLegalSwap(swap)
	if err != nil {
		t.Fatalf("IsLegalSwap error: %v", err)
	}
	if ok {
		t.Fatalf("expected same-square swap to be rejected")
	}
}

// White piece swapped with black piece (different colors) — swap should be rejected.
func TestIsLegalSwap_EnemyPiece(t *testing.T) {
	b := normal.NewGamePosition()
	s := newStateWithBoard(b)

	swap := chaos.Swap{A: sq2(0, 0), B: sq2(0, 7)}
	ok, err := s.IsLegalSwap(swap)
	if err != nil {
		t.Fatalf("IsLegalSwap error: %v", err)
	}
	if ok {
		t.Fatalf("expected black/white piece swap to be rejected")
	}
}

// White king e1 swapped with white queen d1 — king anchor rule rejects it.
func TestIsLegalSwap_KingAnchor(t *testing.T) {
	b := normal.NewGamePosition()
	s := newStateWithBoard(b)

	swap := chaos.Swap{A: sq2(4, 0), B: sq2(3, 0)}
	ok, err := s.IsLegalSwap(swap)
	if err != nil {
		t.Fatalf("IsLegalSwap error: %v", err)
	}
	if ok {
		t.Fatalf("expected king-any swap to be rejected")
	}
}

// White pawn at rank 6 swapped with knight at rank 7 — pawn would land on last rank → rejected.
func TestIsLegalSwap_PawnToLastRank_White(t *testing.T) {
	b := normal.NewEmptyPosition()
	placeBareKings(b)
	b.SetPiece(sq2(0, 6), normal.Piece{Type: normal.Pawn, Color: normal.White})
	b.SetPiece(sq2(1, 7), normal.Piece{Type: normal.Knight, Color: normal.White})
	b.SideToMove = normal.White
	s := newStateWithBoard(b)

	swap := chaos.Swap{A: sq2(0, 6), B: sq2(1, 7)}
	ok, err := s.IsLegalSwap(swap)
	if err != nil {
		t.Fatalf("IsLegalSwap error: %v", err)
	}
	if ok {
		t.Fatalf("expected pawn-to-last-rank swap to be rejected")
	}
}

// Black pawn at rank 1 swapped with knight at rank 0 — pawn would land on last rank → rejected.
func TestIsLegalSwap_PawnToLastRank_Black(t *testing.T) {
	b := normal.NewEmptyPosition()
	placeBareKings(b)
	b.SetPiece(sq2(0, 1), normal.Piece{Type: normal.Pawn, Color: normal.Black})
	b.SetPiece(sq2(1, 0), normal.Piece{Type: normal.Knight, Color: normal.Black})
	b.SideToMove = normal.Black
	s := newStateWithBoard(b)

	swap := chaos.Swap{A: sq2(0, 1), B: sq2(1, 0)}
	ok, err := s.IsLegalSwap(swap)
	if err != nil {
		t.Fatalf("IsLegalSwap error: %v", err)
	}
	if ok {
		t.Fatalf("expected pawn-to-last-rank swap to be rejected")
	}
}

// Two pawns on h2 and g2, both completely blocked (h3,g3,f3 occupied) — neither has a legal move → rejected.
func TestIsLegalSwap_MobilityClause_BothPiecesImmobile(t *testing.T) {
	b := normal.NewEmptyPosition()
	placeBareKings(b)

	b.SetPiece(sq2(7, 1), normal.Piece{Type: normal.Pawn, Color: normal.White})
	b.SetPiece(sq2(6, 1), normal.Piece{Type: normal.Pawn, Color: normal.White})
	b.SetPiece(sq2(7, 2), normal.Piece{Type: normal.Bishop, Color: normal.White})
	b.SetPiece(sq2(6, 2), normal.Piece{Type: normal.Bishop, Color: normal.White})
	b.SetPiece(sq2(5, 2), normal.Piece{Type: normal.Bishop, Color: normal.White})
	b.SideToMove = normal.White
	s := newStateWithBoard(b)

	swap := chaos.Swap{A: sq2(7, 1), B: sq2(6, 1)}
	ok, err := s.IsLegalSwap(swap)
	if err != nil {
		t.Fatalf("IsLegalSwap error: %v", err)
	}
	if ok {
		t.Fatalf("expected swap with both pieces immobile to be rejected")
	}
}

// White rook on e2 blocks black rook on e8 from attacking king on e1.
// Swapping rook e2<->knight a2 leaves e2 occupied — king stays safe and board is not mutated.
func TestIsLegalSwap_KingSafety_NoMutationAfterTempSwap(t *testing.T) {
	b := normal.NewEmptyPosition()
	placeBareKings(b)

	b.SetPiece(sq2(4, 1), normal.Piece{Type: normal.Rook, Color: normal.White})
	b.SetPiece(sq2(4, 7), normal.Piece{Type: normal.Rook, Color: normal.Black})
	b.SetPiece(sq2(0, 1), normal.Piece{Type: normal.Knight, Color: normal.White})
	b.SideToMove = normal.White
	s := newStateWithBoard(b)

	before := cloned(b)
	swap := chaos.Swap{A: sq2(4, 1), B: sq2(0, 1)}
	ok, err := s.IsLegalSwap(swap)
	if err != nil {
		t.Fatalf("IsLegalSwap error: %v", err)
	}
	// After temporary swap both squares remain occupied by friendly pieces so the e-file is still blocked.
	if !ok {
		t.Fatalf("expected swap to be safe (both squares occupied blocks the attack)")
	}
	if !boardsEqual(before, b) {
		t.Fatalf("IsLegalSwap mutated board state after temporary swap")
	}
}

// Swap white knights b1<->g1, then undo — pieces exchange and board restores.
func TestTrySwap_Basic(t *testing.T) {
	b := normal.NewGamePosition()
	s := newStateWithBoard(b)

	before := cloned(b)
	swap := chaos.Swap{A: sq2(1, 0), B: sq2(6, 0)}
	undo := s.TrySwap(swap)

	gotA := b.PieceAt(sq2(1, 0))
	gotB := b.PieceAt(sq2(6, 0))
	if gotA.Type != normal.Knight || gotA.Color != normal.White {
		t.Fatalf("expected knight at b1 after swap, got %+v", gotA)
	}
	if gotB.Type != normal.Knight || gotB.Color != normal.White {
		t.Fatalf("expected knight at g1 after swap, got %+v", gotB)
	}

	s.UndoSwap(undo)
	if !boardsEqual(before, b) {
		t.Fatalf("UndoSwap did not restore board state")
	}
}

// Swap white rook a1 with knight b1 — WhiteQueenSide castling right is revoked.
func TestTrySwap_CastlingRightsRevoked_WhiteQueenSide(t *testing.T) {
	b := normal.NewGamePosition()
	s := newStateWithBoard(b)

	b.CastlingRights = normal.WhiteKingSide | normal.WhiteQueenSide | normal.BlackKingSide | normal.BlackQueenSide

	swap := chaos.Swap{A: sq2(0, 0), B: sq2(1, 0)}
	_ = s.TrySwap(swap)

	if (b.CastlingRights & normal.WhiteQueenSide) != 0 {
		t.Fatalf("expected WhiteQueenSide rights revoked after rook leaves a1")
	}
	if (b.CastlingRights & normal.WhiteKingSide) == 0 {
		t.Fatalf("expected WhiteKingSide rights preserved after rook leaves a1")
	}
	if (b.CastlingRights & normal.BlackKingSide) == 0 {
		t.Fatalf("expected BlackKingSide rights preserved after rook leaves a1")
	}
	if (b.CastlingRights & normal.BlackQueenSide) == 0 {
		t.Fatalf("expected BlackQueenSide rights preserved after rook leaves a1")
	}

	s.UndoSwap(&chaos.UndoSwap{
		A:              sq2(0, 0),
		B:              sq2(1, 0),
		SwapPieceA:     normal.Piece{Type: normal.Rook, Color: normal.White},
		SwapPieceB:     normal.Piece{Type: normal.Knight, Color: normal.White},
		PrevCastle:     normal.WhiteKingSide | normal.WhiteQueenSide | normal.BlackKingSide | normal.BlackQueenSide,
		PrevEnPassant:  normal.NoSquare,
		PrevSideToMove: normal.White,
	})
}

// Swap white rook h1 with knight g1 — WhiteKingSide castling right is revoked.
func TestTrySwap_CastlingRightsRevoked_WhiteKingSide(t *testing.T) {
	b := normal.NewEmptyPosition()
	placeBareKings(b)
	b.SetPiece(sq2(7, 0), normal.Piece{Type: normal.Rook, Color: normal.White})
	b.SetPiece(sq2(6, 0), normal.Piece{Type: normal.Knight, Color: normal.White})
	b.CastlingRights = normal.WhiteKingSide
	b.SideToMove = normal.White
	s := newStateWithBoard(b)

	swap := chaos.Swap{A: sq2(7, 0), B: sq2(6, 0)}
	_ = s.TrySwap(swap)

	if (b.CastlingRights & normal.WhiteKingSide) != 0 {
		t.Fatalf("expected WhiteKingSide rights revoked after rook leaves h1")
	}
}

// Swap black rook a8 with knight b8 — BlackQueenSide castling right is revoked.
func TestTrySwap_CastlingRightsRevoked_BlackQueenSide(t *testing.T) {
	b := normal.NewEmptyPosition()
	placeBareKings(b)
	b.SetPiece(sq2(0, 7), normal.Piece{Type: normal.Rook, Color: normal.Black})
	b.SetPiece(sq2(1, 7), normal.Piece{Type: normal.Knight, Color: normal.Black})
	b.CastlingRights = normal.BlackQueenSide
	b.SideToMove = normal.Black
	s := newStateWithBoard(b)

	swap := chaos.Swap{A: sq2(0, 7), B: sq2(1, 7)}
	_ = s.TrySwap(swap)

	if (b.CastlingRights & normal.BlackQueenSide) != 0 {
		t.Fatalf("expected BlackQueenSide rights revoked after rook leaves a8")
	}
}

// Swap black rook h8 with knight g8 — BlackKingSide castling right is revoked.
func TestTrySwap_CastlingRightsRevoked_BlackKingSide(t *testing.T) {
	b := normal.NewEmptyPosition()
	placeBareKings(b)
	b.SetPiece(sq2(7, 7), normal.Piece{Type: normal.Rook, Color: normal.Black})
	b.SetPiece(sq2(6, 7), normal.Piece{Type: normal.Knight, Color: normal.Black})
	b.CastlingRights = normal.BlackKingSide
	b.SideToMove = normal.Black
	s := newStateWithBoard(b)

	swap := chaos.Swap{A: sq2(7, 7), B: sq2(6, 7)}
	_ = s.TrySwap(swap)

	if (b.CastlingRights & normal.BlackKingSide) != 0 {
		t.Fatalf("expected BlackKingSide rights revoked after rook leaves h8")
	}
}

// En passant target is set before swap; TrySwap must clear it.
func TestTrySwap_EnPassantCleared(t *testing.T) {
	b := normal.NewGamePosition()
	s := newStateWithBoard(b)

	b.EnPassantSquare = sq2(3, 5)
	swap := chaos.Swap{A: sq2(1, 0), B: sq2(6, 0)}
	_ = s.TrySwap(swap)

	if b.EnPassantSquare.IsValid() {
		t.Fatalf("expected en passant square cleared after swap")
	}
}

// TrySwap must toggle the side to move after applying a swap.
func TestTrySwap_SideToggled(t *testing.T) {
	b := normal.NewGamePosition()
	s := newStateWithBoard(b)

	if b.SideToMove != normal.White {
		t.Fatalf("expected white to move initially")
	}

	swap := chaos.Swap{A: sq2(1, 0), B: sq2(6, 0)}
	_ = s.TrySwap(swap)

	if b.SideToMove != normal.Black {
		t.Fatalf("expected black to move after swap")
	}
}

// Full round-trip: set up position with castling rights and EP, swap, undo restores everything.
func TestUndoSwap_RestoresAllState(t *testing.T) {
	b := normal.NewEmptyPosition()
	placeBareKings(b)
	b.SetPiece(sq2(0, 0), normal.Piece{Type: normal.Rook, Color: normal.White})
	b.SetPiece(sq2(1, 0), normal.Piece{Type: normal.Knight, Color: normal.White})
	b.SetPiece(sq2(5, 4), normal.Piece{Type: normal.Pawn, Color: normal.Black})
	b.CastlingRights = normal.WhiteKingSide | normal.WhiteQueenSide
	b.EnPassantSquare = sq2(5, 5)
	b.SideToMove = normal.White
	s := newStateWithBoard(b)

	before := cloned(b)
	swap := chaos.Swap{A: sq2(0, 0), B: sq2(1, 0)}
	undo := s.TrySwap(swap)

	s.UndoSwap(undo)
	if !boardsEqual(before, b) {
		t.Fatalf("UndoSwap did not restore all board state")
	}
}

// Standard starting position should have at least one legal swap pair.
func TestPickSwapPair_ReturnsSwapInStandardPosition(t *testing.T) {
	b := normal.NewGamePosition()
	s := newStateWithBoard(b)

	swap, ok := s.PickSwapPair()
	if !ok {
		t.Fatalf("expected at least one legal swap in standard position")
	}
	legal, err := s.IsLegalSwap(swap)
	if err != nil {
		t.Fatalf("IsLegalSwap error on picked swap: %v", err)
	}
	if !legal {
		t.Fatalf("PickSwapPair returned illegal swap %v", swap)
	}
}

// Only kings on the board — PickSwapPair must return false.
func TestPickSwapPair_NoLegalSwaps(t *testing.T) {
	b := normal.NewEmptyPosition()
	placeBareKings(b)
	b.SideToMove = normal.White
	s := newStateWithBoard(b)

	_, ok := s.PickSwapPair()
	if ok {
		t.Fatalf("expected no legal swaps when only kings are on board")
	}
}

// One rook and kings — only one non-king piece, no i<j pair exists.
func TestPickSwapPair_NoSwapWhenOnlyOneNonKingPiece(t *testing.T) {
	b := normal.NewEmptyPosition()
	placeBareKings(b)
	b.SetPiece(sq2(0, 0), normal.Piece{Type: normal.Rook, Color: normal.White})
	b.SideToMove = normal.White
	s := newStateWithBoard(b)

	_, ok := s.PickSwapPair()
	if ok {
		t.Fatalf("expected no swap when only one non-king piece exists")
	}
}

// Two identical seeds should produce the same swap pair.
func TestPickSwapPair_Deterministic(t *testing.T) {
	b := normal.NewGamePosition()
	s1 := newStateWithBoard(b)

	swap1, ok1 := s1.PickSwapPair()
	if !ok1 {
		t.Fatalf("expected swap pair from standard position")
	}

	b2 := normal.NewGamePosition()
	s2 := newStateWithBoard(b2)

	swap2, ok2 := s2.PickSwapPair()
	if !ok2 {
		t.Fatalf("expected swap pair from standard position")
	}

	if swap1.A != swap2.A || swap1.B != swap2.B {
		t.Fatalf("expected deterministic swap pair with same seed, got %v vs %v", swap1, swap2)
	}
}

// TrySwap toggles side; UndoSwap restores original board including side to move.
func TestTrySwapAndUndo_TrySwapThenMakeMove(t *testing.T) {
	b := normal.NewGamePosition()
	s := newStateWithBoard(b)
	before := cloned(b)

	swap := chaos.Swap{A: sq2(1, 0), B: sq2(6, 0)}
	undoSwap := s.TrySwap(swap)

	if b.SideToMove != normal.Black {
		t.Fatalf("expected black to move after swap")
	}

	s.UndoSwap(undoSwap)

	if !boardsEqual(before, b) {
		t.Fatalf("UndoSwap did not restore after TrySwap")
	}
}

func cloned(b *normal.BoardState) *normal.BoardState {
	nb := &normal.BoardState{
		SideToMove:      b.SideToMove,
		CastlingRights:  b.CastlingRights,
		EnPassantSquare: b.EnPassantSquare,
		GameState:       b.GameState,
	}
	nb.ChessBoard = b.ChessBoard
	return nb
}
