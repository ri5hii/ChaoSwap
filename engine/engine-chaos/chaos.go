package engine

import (
	"math/rand/v2"
	"time"

	normal "github.com/ri5hii/ChaoSwap/engine/engine-normal"
)

// Package engine implements the Chaos Chess (Swap Variant) engine.
//
// Chaos Chess is a variant where the side to move may optionally perform
// a random legal swap of two friendly pieces instead of a standard chess move.
// Swaps toggle the side to move, clear en passant, and may revoke castling rights
// when a rook leaves its starting corner.
//
// Swap contains the details of a piece swap between two squares.
// It is used to record the swap action for move history and potential undo functionality.
type Swap struct {
	A normal.Square
	B normal.Square
}

// UndoSwap captures the state required to restore the board after a swap.
//
// It stores:
//   - the two squares involved in the swap
//   - the pieces that were on those squares before the swap
//   - the previous castling rights and en passant square
//   - the previous side to move
type UndoSwap struct {
	A                      normal.Square
	B                      normal.Square
	SwapPieceA, SwapPieceB normal.Piece
	PrevCastle             uint8
	PrevEnPassant          normal.Square
	PrevSideToMove         normal.PieceColor
}

// State represents the current game state for Chaos Swap,
// including the underlying board state and any additional information needed to manage swaps and randomness.
type State struct {
	Base *normal.BoardState
	RNG  *rand.Rand
}

// NewState initializes a new Chaos Swap game state with the standard chess starting position and a seeded RNG.
func NewState() *State {
	return &State{
		Base: normal.NewGamePosition(),
		RNG:  rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano()>>32))),
	}
}

// IsLegalSwap reports whether a swap is legal in the current position.
//
// It checks, in order:
//   - occupancy and ownership (both squares must hold friendly pieces)
//   - king anchor (neither piece may be a king)
//   - pawn zone restriction (pawns cannot swap onto rank 1 or 8)
//   - mobility clause (at least one piece must have a legal standard move)
//   - king safety (the side to move must not be in check after the swap)
func (s *State) IsLegalSwap(swap Swap) (bool, error) {
	pieceA := s.Base.PieceAt(swap.A)
	pieceB := s.Base.PieceAt(swap.B)

	// occupancy and ownership checks
	if pieceA.Type == normal.None || pieceB.Type == normal.None {
		return false, nil
	}
	if swap.A.Equals(swap.B) {
		return false, nil
	}
	sideToMove := s.Base.SideToMove
	if pieceA.Color != sideToMove || pieceB.Color != sideToMove {
		return false, nil
	}

	// king anchor rule
	if pieceA.Type == normal.King || pieceB.Type == normal.King {
		return false, nil
	}

	// pawn restriction rule
	if pieceA.Type == normal.Pawn {
		rank := swap.B.Rank
		if rank == 0 || rank == 7 {
			return false, nil
		}
	}
	if pieceB.Type == normal.Pawn {
		rank := swap.A.Rank
		if rank == 0 || rank == 7 {
			return false, nil
		}
	}

	// mobility clause
	mobile := func(piece normal.Piece, square normal.Square) bool {
		candidates := s.Base.AppendPseudoLegalMovesForPiece(nil, piece, square)
		for _, move := range candidates {
			ok, err := s.Base.IsMoveLegal(piece, move.From, move.To, move.Promotion)
			if err == nil && ok {
				return true
			}
		}
		return false
	}
	if !mobile(pieceA, swap.A) && !mobile(pieceB, swap.B) {
		return false, nil
	}
	s.Base.SetPiece(swap.A, pieceB)
	s.Base.SetPiece(swap.B, pieceA)

	inCheck, err := s.Base.InCheck(sideToMove)

	s.Base.SetPiece(swap.A, pieceA)
	s.Base.SetPiece(swap.B, pieceB)

	if err != nil {
		return false, err
	}
	if inCheck {
		return false, nil
	}
	return true, nil
}

// TrySwap applies a swap to the board state without checking legality.
//
// It:
//   - records pre-swap state into an UndoSwap record
//   - revokes castling rights if a rook leaves a starting corner
//   - exchanges the pieces on square A and B
//   - clears the en passant target square
//   - toggles the side to move
//
// Callers should validate legality via IsLegalSwap before calling TrySwap.
func (s *State) TrySwap(swap Swap) *UndoSwap {
	pieceA := s.Base.PieceAt(swap.A)
	pieceB := s.Base.PieceAt(swap.B)
	undoSwap := &UndoSwap{
		A:              swap.A,
		B:              swap.B,
		SwapPieceA:     pieceA,
		SwapPieceB:     pieceB,
		PrevCastle:     s.Base.CastlingRights,
		PrevEnPassant:  s.Base.EnPassantSquare,
		PrevSideToMove: s.Base.SideToMove,
	}

	if pieceA.Type == normal.Rook {
		switch swap.A {
		case normal.Square{File: 0, Rank: 0}:
			s.Base.CastlingRights &^= normal.WhiteQueenSide
		case normal.Square{File: 7, Rank: 0}:
			s.Base.CastlingRights &^= normal.WhiteKingSide
		case normal.Square{File: 0, Rank: 7}:
			s.Base.CastlingRights &^= normal.BlackQueenSide
		case normal.Square{File: 7, Rank: 7}:
			s.Base.CastlingRights &^= normal.BlackKingSide
		}
	}

	if pieceB.Type == normal.Rook {
		switch swap.B {
		case normal.Square{File: 0, Rank: 0}:
			s.Base.CastlingRights &^= normal.WhiteQueenSide
		case normal.Square{File: 7, Rank: 0}:
			s.Base.CastlingRights &^= normal.WhiteKingSide
		case normal.Square{File: 0, Rank: 7}:
			s.Base.CastlingRights &^= normal.BlackQueenSide
		case normal.Square{File: 7, Rank: 7}:
			s.Base.CastlingRights &^= normal.BlackKingSide
		}
	}

	s.Base.SetPiece(swap.A, pieceB)
	s.Base.SetPiece(swap.B, pieceA)

	s.Base.EnPassantSquare = normal.NoSquare

	s.Base.SideToMove = 1 - s.Base.SideToMove

	return undoSwap
}

// UndoSwap restores the board state from an UndoSwap record created by TrySwap.
func (s *State) UndoSwap(undo *UndoSwap) {
	s.Base.SetPiece(undo.A, undo.SwapPieceA)
	s.Base.SetPiece(undo.B, undo.SwapPieceB)

	s.Base.CastlingRights = undo.PrevCastle

	s.Base.EnPassantSquare = undo.PrevEnPassant

	s.Base.SideToMove = undo.PrevSideToMove
}

func (s *State) PickSwapPair() {
}
