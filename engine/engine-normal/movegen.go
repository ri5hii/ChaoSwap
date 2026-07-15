package engine

// This file contains pseudo-legal move generation helpers.
//
// "Pseudo-legal" means the move follows piece movement and occupancy rules, but may still
// be illegal if it leaves the mover's king in check. Callers should filter candidates
// using existing king-safety validation (e.g. trying the move and checking InCheck, or
// using an IsMoveLegal helper if present).
//
// Primary use-case: efficiently detecting whether the side to move has *any* legal move
// (for stalemate/checkmate evaluation) without scanning every possible destination square.

// MoveCandidate represents a single move attempt from->to, with an optional promotion piece.
type MoveCandidate struct {
	From      Square
	To        Square
	Promotion PieceType // None when not a promotion move
}

// AppendPseudoLegalMovesForSide appends pseudo-legal moves for `side` into dst and returns dst.
//
// Invariant: it never returns moves that capture a friendly piece or move off-board.
// It may return moves that are illegal due to king safety (self-check).
func (board *BoardState) AppendPseudoLegalMovesForSide(dst []MoveCandidate, side PieceColor) []MoveCandidate {
	if board == nil {
		return dst
	}

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			piece := board.ChessBoard[rank][file]
			if piece.Type == None || piece.Color != side {
				continue
			}
			from := Square{File: file, Rank: rank}
			dst = board.AppendPseudoLegalMovesForPiece(dst, piece, from)
		}
	}

	return dst
}

// AppendPseudoLegalMovesForPiece appends pseudo-legal moves for the given piece at from into dst.
//
// Note: For pawns reaching the last rank, it appends 4 promotion candidates (Q/R/B/N).
func (board *BoardState) AppendPseudoLegalMovesForPiece(dst []MoveCandidate, piece Piece, from Square) []MoveCandidate {
	if board == nil || piece.Type == None || !from.IsValid() {
		return dst
	}

	switch piece.Type {
	case Pawn:
		return board.appendPawnPseudoMoves(dst, piece, from)
	case Knight:
		return board.appendKnightPseudoMoves(dst, piece, from)
	case Bishop:
		return board.appendSlidingPseudoMoves(dst, piece, from, [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}})
	case Rook:
		return board.appendSlidingPseudoMoves(dst, piece, from, [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}})
	case Queen:
		return board.appendSlidingPseudoMoves(dst, piece, from, [][2]int{
			{1, 0}, {-1, 0}, {0, 1}, {0, -1},
			{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
		})
	case King:
		return board.appendKingPseudoMoves(dst, piece, from)
	default:
		return dst
	}
}

func (board *BoardState) appendPawnPseudoMoves(dst []MoveCandidate, piece Piece, from Square) []MoveCandidate {
	dir := -1
	startRank := 6
	promoRank := 0
	if piece.Color == Black {
		dir = 1
		startRank = 1
		promoRank = 7
	}

	one := Square{File: from.File, Rank: from.Rank + dir}
	if one.IsValid() && board.PieceAt(one).Type == None {
		dst = appendPawnMoveWithPromotion(dst, from, one, promoRank)

		two := Square{File: from.File, Rank: from.Rank + 2*dir}
		if from.Rank == startRank && two.IsValid() && board.PieceAt(two).Type == None {
			dst = append(dst, MoveCandidate{From: from, To: two, Promotion: None})
		}
	}

	for _, df := range []int{-1, 1} {
		capSq := Square{File: from.File + df, Rank: from.Rank + dir}
		if !capSq.IsValid() {
			continue
		}
		t := board.PieceAt(capSq)
		if t.Type != None && t.Color != piece.Color {
			dst = appendPawnMoveWithPromotion(dst, from, capSq, promoRank)
		}
	}

	if board.EnPassantSquare.IsValid() {
		ep := board.EnPassantSquare
		if ep.Rank == from.Rank+dir && (ep.File == from.File-1 || ep.File == from.File+1) {
			dst = append(dst, MoveCandidate{From: from, To: ep, Promotion: None})
		}
	}

	return dst
}

func appendPawnMoveWithPromotion(dst []MoveCandidate, from Square, to Square, promoRank int) []MoveCandidate {
	if to.Rank != promoRank {
		return append(dst, MoveCandidate{From: from, To: to, Promotion: None})
	}

	// Generate all promotion choices. King safety is checked later by the caller.
	dst = append(dst,
		MoveCandidate{From: from, To: to, Promotion: Queen},
		MoveCandidate{From: from, To: to, Promotion: Rook},
		MoveCandidate{From: from, To: to, Promotion: Bishop},
		MoveCandidate{From: from, To: to, Promotion: Knight},
	)
	return dst
}

func (board *BoardState) appendKnightPseudoMoves(dst []MoveCandidate, piece Piece, from Square) []MoveCandidate {
	steps := [8][2]int{
		{1, 2}, {2, 1}, {2, -1}, {1, -2},
		{-1, -2}, {-2, -1}, {-2, 1}, {-1, 2},
	}

	for _, s := range steps {
		to := Square{File: from.File + s[0], Rank: from.Rank + s[1]}
		if !to.IsValid() {
			continue
		}
		t := board.PieceAt(to)
		if t.Type != None && t.Color == piece.Color {
			continue
		}
		dst = append(dst, MoveCandidate{From: from, To: to, Promotion: None})
	}

	return dst
}

func (board *BoardState) appendSlidingPseudoMoves(dst []MoveCandidate, piece Piece, from Square, dirs [][2]int) []MoveCandidate {
	for _, d := range dirs {
		df, dr := d[0], d[1]
		for f, r := from.File+df, from.Rank+dr; f >= 0 && f < 8 && r >= 0 && r < 8; f, r = f+df, r+dr {
			to := Square{File: f, Rank: r}
			t := board.PieceAt(to)
			if t.Type == None {
				dst = append(dst, MoveCandidate{From: from, To: to, Promotion: None})
				continue
			}

			// Occupied: capture if enemy, then stop, or stop if friendly.
			if t.Color != piece.Color {
				dst = append(dst, MoveCandidate{From: from, To: to, Promotion: None})
			}
			break
		}
	}
	return dst
}

func (board *BoardState) appendKingPseudoMoves(dst []MoveCandidate, piece Piece, from Square) []MoveCandidate {
	for dr := -1; dr <= 1; dr++ {
		for df := -1; df <= 1; df++ {
			if df == 0 && dr == 0 {
				continue
			}
			to := Square{File: from.File + df, Rank: from.Rank + dr}
			if !to.IsValid() {
				continue
			}
			t := board.PieceAt(to)
			if t.Type != None && t.Color == piece.Color {
				continue
			}
			dst = append(dst, MoveCandidate{From: from, To: to, Promotion: None})
		}
	}

	// Castling candidates are included as pseudo-legal moves; final legality is checked later.
	// White king starts at e1 -> (4,7) in this engine's coordinate system.
	// Black king starts at e8 -> (4,0).
	if piece.Color == White && from.File == 4 && from.Rank == 7 {
		if (board.CastlingRights & WhiteKingSide) != 0 {
			dst = append(dst, MoveCandidate{From: from, To: Square{File: 6, Rank: 7}, Promotion: None})
		}
		if (board.CastlingRights & WhiteQueenSide) != 0 {
			dst = append(dst, MoveCandidate{From: from, To: Square{File: 2, Rank: 7}, Promotion: None})
		}
	}

	if piece.Color == Black && from.File == 4 && from.Rank == 0 {
		if (board.CastlingRights & BlackKingSide) != 0 {
			dst = append(dst, MoveCandidate{From: from, To: Square{File: 6, Rank: 0}, Promotion: None})
		}
		if (board.CastlingRights & BlackQueenSide) != 0 {
			dst = append(dst, MoveCandidate{From: from, To: Square{File: 2, Rank: 0}, Promotion: None})
		}
	}

	return dst
}
