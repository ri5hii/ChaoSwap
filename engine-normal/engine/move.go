package engine

import "fmt"

// Undo captures the minimal state required to restore the board after a successful MakeMove.
//
// It stores:
//   - the moved piece's origin and destination
//   - the piece captured on the destination square (if any)
//   - the previous castling rights and en passant square
//   - the previous side to move
//   - promotion information, if the move promoted a pawn
type Undo struct {
	From           Square
	To             Square
	Captured       Piece
	PrevCastle     uint8
	PrevEnPassant  Square
	PrevSideToMove PieceColor
	PrevPromotion  PieceType
	PromotedTo     PieceType
}

// MakeMove applies a move to the board if it is legal.
//
// It validates:
//   - basic bounds and turn rules
//   - piece-specific movement rules
//   - special rules (castling, en passant, promotion)
//   - king safety (the mover may not leave their king in check)
//
// On success it mutates the board and flips SideToMove, returning an Undo record.
// On failure it returns an error and leaves the board unchanged.
//
// promotionType is only used when a pawn reaches the last rank; pass None otherwise.
func (board *BoardState) MakeMove(piece Piece, from Square, to Square, promotionType PieceType) (*Undo, error) {
	if !from.IsValid() || !to.IsValid() {
		return nil, fmt.Errorf("Invalid move: source or destination square is out of bounds.")
	}

	sourcePiece := board.PieceAt(from)
	if sourcePiece.Type == None {
		return nil, fmt.Errorf("Invalid move: no piece at %s.", SquareNotation(from))
	}

	if piece.Color != board.SideToMove {
		return nil, fmt.Errorf("Invalid move: it's not your turn.")
	}

	target := board.PieceAt(to)
	if target.Type != None && target.Color == piece.Color {
		return nil, fmt.Errorf("Invalid move: destination %s is occupied by your own piece.", SquareNotation(to))
	}

	undo := Undo{
		From:           from,
		To:             to,
		Captured:       target,
		PrevCastle:     board.CastlingRights,
		PrevEnPassant:  board.EnPassantSquare,
		PrevSideToMove: board.SideToMove,
		PrevPromotion:  None,
		PromotedTo:     promotionType,
	}

	prevEnPassant := board.EnPassantSquare

	var valid bool
	var err error

	switch piece.Type {
	case Pawn:
		valid, err = board.handlePawnMove(piece, from, to, &undo)
	case Knight:
		valid, err = board.handleKnightMove(piece, from, to)
	case Bishop:
		valid, err = board.handleSlidingMove(piece, from, to)
	case Rook:
		valid, err = board.handleRookMove(piece, from, to)
	case Queen:
		valid, err = board.handleSlidingMove(piece, from, to)
	case King:
		valid, err = board.handleKingMove(piece, from, to)
	default:
		return nil, fmt.Errorf("Invalid move: unknown piece type.")
	}

	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("Invalid move.")
	}

	if board.EnPassantSquare.Equals(prevEnPassant) {
		board.EnPassantSquare = NoSquare
	}

	kingSquare := board.FindKing(piece.Color)
	inCheck, checkErr := board.IsSquareAttacked(kingSquare, PieceColor(1-piece.Color))
	if checkErr != nil {
		return nil, checkErr
	}
	if inCheck {
		UndoMove(board, undo)
		return nil, fmt.Errorf("Invalid move: cannot leave your king in check.")
	}

	board.SideToMove = 1 - board.SideToMove
	return &undo, nil
}

// handlePawnMove validates and applies pawn movement, including en passant and promotion.
//
// The undo record is updated for promotion so UndoMove can restore the original pawn.
func (board *BoardState) handlePawnMove(piece Piece, from Square, to Square, undo *Undo) (bool, error) {
	fileDiff := to.File - from.File
	rankDiff := to.Rank - from.Rank
	absoluteFileDiff := absoluteValue(fileDiff)

	target := board.PieceAt(to)

	if absoluteFileDiff == 1 && to.Equals(board.EnPassantSquare) {
		capRank := to.Rank + 1
		if piece.Color == Black {
			capRank = to.Rank - 1
		}
		board.SetPiece(Square{File: to.File, Rank: capRank}, Piece{Type: None})
		board.SetPiece(to, piece)
		board.SetPiece(from, Piece{Type: None})
		return true, nil
	}

	if piece.Color == White {
		if fileDiff == 0 && rankDiff == -1 && target.Type == None {
			board.SetPiece(to, piece)
			board.SetPiece(from, Piece{Type: None})
		} else if fileDiff == 0 && rankDiff == -2 && from.Rank == 6 && board.PieceAt(Square{File: from.File, Rank: 5}).Type == None && target.Type == None {
			board.EnPassantSquare = Square{File: from.File, Rank: 5}
			board.SetPiece(to, piece)
			board.SetPiece(from, Piece{Type: None})
		} else if absoluteFileDiff == 1 && rankDiff == -1 && target.Type != None {
			if to.Rank == 0 && target.Type == Rook {
				board.UpdateRookCastlingRights(from, to, piece)
			}
			board.SetPiece(to, piece)
			board.SetPiece(from, Piece{Type: None})
		} else {
			return false, fmt.Errorf("Invalid move for white pawn.")
		}

		if to.Rank == 0 {
			promo := Queen
			if undo != nil && undo.PromotedTo != None {
				promo = undo.PromotedTo
			}

			if undo != nil {
				undo.PrevPromotion = Pawn
				undo.PromotedTo = promo
			}

			ok, promoErr := board.PawnPromotion(piece, to, promo)
			if promoErr != nil {
				return false, promoErr
			}
			return ok, nil
		}

		return true, nil
	}

	if fileDiff == 0 && rankDiff == 1 && target.Type == None {
		board.SetPiece(to, piece)
		board.SetPiece(from, Piece{Type: None})
	} else if fileDiff == 0 && rankDiff == 2 && from.Rank == 1 && board.PieceAt(Square{File: from.File, Rank: 2}).Type == None && target.Type == None {
		board.EnPassantSquare = Square{File: from.File, Rank: 2}
		board.SetPiece(to, piece)
		board.SetPiece(from, Piece{Type: None})
	} else if absoluteFileDiff == 1 && rankDiff == 1 && target.Type != None {
		if to.Rank == 7 && target.Type == Rook {
			board.UpdateRookCastlingRights(from, to, piece)
		}
		board.SetPiece(to, piece)
		board.SetPiece(from, Piece{Type: None})
	} else {
		return false, fmt.Errorf("Invalid move for black pawn.")
	}

	if to.Rank == 7 {
		promo := Queen
		if undo != nil && undo.PromotedTo != None {
			promo = undo.PromotedTo
		}

		if undo != nil {
			undo.PrevPromotion = Pawn
			undo.PromotedTo = promo
		}

		ok, promoErr := board.PawnPromotion(piece, to, promo)
		if promoErr != nil {
			return false, promoErr
		}
		return ok, nil
	}

	return true, nil
}

// handleKnightMove validates and applies a knight move.
func (board *BoardState) handleKnightMove(piece Piece, from Square, to Square) (bool, error) {
	target := board.PieceAt(to)

	absoluteFileDiff := absoluteValue(to.File - from.File)
	absoluteRankDiff := absoluteValue(to.Rank - from.Rank)

	if (absoluteFileDiff == 2 && absoluteRankDiff == 1) || (absoluteFileDiff == 1 && absoluteRankDiff == 2) {
		if (to.Rank == 0 || to.Rank == 7) && target.Type == Rook {
			board.UpdateRookCastlingRights(from, to, piece)
		}
		board.SetPiece(to, piece)
		board.SetPiece(from, Piece{Type: None})
		return true, nil
	}

	return false, fmt.Errorf("Invalid move for knight.")
}

// PawnPromotion replaces a pawn on the destination square with the chosen promotion piece.
func (board *BoardState) PawnPromotion(pawn Piece, to Square, promotionType PieceType) (bool, error) {
	if pawn.Type != Pawn {
		return false, fmt.Errorf("Invalid promotion: piece is not a pawn.")
	}

	if promotionType != Queen && promotionType != Rook && promotionType != Bishop && promotionType != Knight {
		return false, fmt.Errorf("Invalid promotion: invalid promotion piece type.")
	}

	board.SetPiece(to, Piece{Type: promotionType, Color: pawn.Color})
	return true, nil
}

// handleSlidingMove validates and applies bishop/queen sliding moves.
func (board *BoardState) handleSlidingMove(piece Piece, from Square, to Square) (bool, error) {
	target := board.PieceAt(to)

	clear, err := board.IsPathClear(piece, from, to)
	if err != nil {
		return false, err
	}
	if !clear {
		return false, fmt.Errorf("Invalid move: path is not clear for sliding piece.")
	}

	if (to.Rank == 0 || to.Rank == 7) && target.Type == Rook {
		board.UpdateRookCastlingRights(from, to, piece)
	}

	board.SetPiece(to, piece)
	board.SetPiece(from, Piece{Type: None})
	return true, nil
}

// handleRookMove validates and applies a rook move and updates castling rights.
func (board *BoardState) handleRookMove(piece Piece, from Square, to Square) (bool, error) {
	clear, err := board.IsPathClear(piece, from, to)
	if err != nil {
		return false, err
	}
	if !clear {
		return false, fmt.Errorf("Invalid move: path is not clear for rook.")
	}

	board.UpdateRookCastlingRights(from, to, piece)

	board.SetPiece(to, piece)
	board.SetPiece(from, Piece{Type: None})
	return true, nil
}

// handleKingMove validates and applies king moves, including castling.
func (board *BoardState) handleKingMove(piece Piece, from Square, to Square) (bool, error) {
	absoluteFileDiff := absoluteValue(to.File - from.File)
	absoluteRankDiff := absoluteValue(to.Rank - from.Rank)

	target := board.PieceAt(to)

	if absoluteFileDiff <= 1 && absoluteRankDiff <= 1 {
		if (to.Rank == 0 || to.Rank == 7) && target.Type == Rook {
			board.UpdateRookCastlingRights(from, to, piece)
		}
		board.clearCastlingRights(piece.Color)
		board.SetPiece(to, piece)
		board.SetPiece(from, Piece{Type: None})
		return true, nil
	}

	if absoluteRankDiff == 0 && absoluteFileDiff == 2 {
		return board.tryCastle(piece, from, to)
	}

	return false, fmt.Errorf("Invalid move for king.")
}

// tryCastle validates and applies castling for the given king move.
func (board *BoardState) tryCastle(king Piece, from Square, to Square) (bool, error) {
	opponent := PieceColor(1 - king.Color)

	var rookFrom, rookTo Square
	var rights uint8

	if to.File == 6 {
		rookFrom = Square{File: 7, Rank: from.Rank}
		rookTo = Square{File: 5, Rank: from.Rank}
		if king.Color == White {
			rights = WhiteKingSide
		} else {
			rights = BlackKingSide
		}
	} else if to.File == 2 {
		rookFrom = Square{File: 0, Rank: from.Rank}
		rookTo = Square{File: 3, Rank: from.Rank}
		if king.Color == White {
			rights = WhiteQueenSide
		} else {
			rights = BlackQueenSide
		}
	} else {
		return false, fmt.Errorf("Invalid move: king can only castle to c1/c8 or g1/g8 based on piece color.")
	}

	if board.CastlingRights&rights == 0 {
		return false, fmt.Errorf("Invalid move: castling rights for this side are not available.")
	}

	attacked, err := board.IsPathAttacked(from, to, opponent)
	if err != nil {
		return false, err
	}
	if attacked {
		return false, fmt.Errorf("Invalid move: cannot castle through or into check.")
	}

	if board.PieceAt(rookFrom).Type != Rook {
		return false, fmt.Errorf("Invalid move: no rook in the correct position for castling.")
	}

	clear, clearErr := board.IsPathClear(Piece{Type: Rook}, rookFrom, from)
	if clearErr != nil {
		return false, clearErr
	}
	if !clear {
		return false, fmt.Errorf("Invalid move: path is not clear for castling.")
	}

	board.SetPiece(to, king)
	board.SetPiece(from, Piece{Type: None})
	board.SetPiece(rookTo, Piece{Type: Rook, Color: king.Color})
	board.SetPiece(rookFrom, Piece{Type: None})

	board.CastlingRights &^= rights
	return true, nil
}

// clearCastlingRights removes all castling rights for the given color.
func (board *BoardState) clearCastlingRights(color PieceColor) {
	if color == White {
		board.CastlingRights &^= WhiteKingSide | WhiteQueenSide
	} else {
		board.CastlingRights &^= BlackKingSide | BlackQueenSide
	}
}

// UpdateRookCastlingRights updates castling rights for rook moves and rook captures.
func (board *BoardState) UpdateRookCastlingRights(from Square, to Square, piece Piece) {
	if piece.Type == Rook {
		switch {
		case from.Equals(Square{0, 7}):
			board.CastlingRights &^= WhiteQueenSide
		case from.Equals(Square{7, 7}):
			board.CastlingRights &^= WhiteKingSide
		case from.Equals(Square{0, 0}):
			board.CastlingRights &^= BlackQueenSide
		case from.Equals(Square{7, 0}):
			board.CastlingRights &^= BlackKingSide
		}
	}

	captured := board.PieceAt(to)
	if captured.Type == Rook {
		switch {
		case to.Equals(Square{0, 7}):
			board.CastlingRights &^= WhiteQueenSide
		case to.Equals(Square{7, 7}):
			board.CastlingRights &^= WhiteKingSide
		case to.Equals(Square{0, 0}):
			board.CastlingRights &^= BlackQueenSide
		case to.Equals(Square{7, 0}):
			board.CastlingRights &^= BlackKingSide
		}
	}
}

// UndoMove restores the board state to what it was before a successful MakeMove.
//
// It restores:
//   - pieces on the moved squares
//   - captured piece (including en passant reconstruction)
//   - castling rights and en passant square
//   - side to move
//   - pawn identity after promotion
func UndoMove(board *BoardState, undo Undo) {
	board.SetPiece(undo.From, board.PieceAt(undo.To))
	board.SetPiece(undo.To, undo.Captured)
	board.CastlingRights = undo.PrevCastle
	board.EnPassantSquare = undo.PrevEnPassant
	board.SideToMove = undo.PrevSideToMove

	if undo.PrevPromotion == Pawn {
		promoted := board.PieceAt(undo.From)
		board.SetPiece(undo.From, Piece{Type: Pawn, Color: promoted.Color})
	}

	piece := board.PieceAt(undo.From)

	if piece.Type == Pawn && undo.From.File != undo.To.File && undo.Captured.Type == None {
		captureRank := undo.To.Rank
		if piece.Color == White {
			captureRank += 1
		} else {
			captureRank -= 1
		}
		board.SetPiece(Square{File: undo.To.File, Rank: captureRank}, Piece{Type: Pawn, Color: PieceColor(1 - piece.Color)})
	}

	if piece.Type == King && absoluteValue(undo.To.File-undo.From.File) == 2 {
		if undo.To.File == 6 {
			rookFrom := Square{File: 7, Rank: undo.From.Rank}
			rookTo := Square{File: 5, Rank: undo.From.Rank}
			board.SetPiece(rookFrom, board.PieceAt(rookTo))
			board.SetPiece(rookTo, Piece{Type: None})
		} else {
			rookFrom := Square{File: 0, Rank: undo.From.Rank}
			rookTo := Square{File: 3, Rank: undo.From.Rank}
			board.SetPiece(rookFrom, board.PieceAt(rookTo))
			board.SetPiece(rookTo, Piece{Type: None})
		}
	}
}

// FindKing returns the square of the king for the given color, or NoSquare if not found.
func (board *BoardState) FindKing(color PieceColor) Square {
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			p := board.ChessBoard[rank][file]
			if p.Type == King && p.Color == color {
				return Square{File: file, Rank: rank}
			}
		}
	}
	return NoSquare
}

// IsSquareAttacked reports whether target is attacked by any piece of ThreatColor.
func (board *BoardState) IsSquareAttacked(target Square, ThreatColor PieceColor) (bool, error) {
	straightDirections := [4][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
	for i := 0; i < len(straightDirections); i++ {
		direction := straightDirections[i]
		fileStep, rankStep := direction[0], direction[1]
		for file, rank := target.File+fileStep, target.Rank+rankStep; file >= 0 && file <= 7 && rank >= 0 && rank <= 7; file, rank = file+fileStep, rank+rankStep {
			p := board.ChessBoard[rank][file]
			if p.Type == None {
				continue
			}
			if p.Color == ThreatColor && (p.Type == Rook || p.Type == Queen) {
				return true, nil
			}
			break
		}
	}

	diagonalDirections := [4][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	for i := 0; i < len(diagonalDirections); i++ {
		direction := diagonalDirections[i]
		fileStep, rankStep := direction[0], direction[1]
		for file, rank := target.File+fileStep, target.Rank+rankStep; file >= 0 && file <= 7 && rank >= 0 && rank <= 7; file, rank = file+fileStep, rank+rankStep {
			p := board.ChessBoard[rank][file]
			if p.Type == None {
				continue
			}
			if p.Color == ThreatColor && (p.Type == Bishop || p.Type == Queen) {
				return true, nil
			}
			break
		}
	}

	knightOffsets := [8][2]int{{1, 2}, {2, 1}, {-1, 2}, {-2, 1}, {1, -2}, {2, -1}, {-1, -2}, {-2, -1}}
	for i := 0; i < len(knightOffsets); i++ {
		offset := knightOffsets[i]
		file := target.File + offset[0]
		rank := target.Rank + offset[1]
		if file >= 0 && file <= 7 && rank >= 0 && rank <= 7 {
			p := board.ChessBoard[rank][file]
			if p.Color == ThreatColor && p.Type == Knight {
				return true, nil
			}
		}
	}

	pawnRankOffset := -1
	if ThreatColor == Black {
		pawnRankOffset = 1
	}
	pawnFileOffsets := [2]int{-1, 1}
	for i := 0; i < len(pawnFileOffsets); i++ {
		fileOffset := pawnFileOffsets[i]
		file := target.File + fileOffset
		rank := target.Rank + pawnRankOffset
		if file >= 0 && file <= 7 && rank >= 0 && rank <= 7 {
			p := board.ChessBoard[rank][file]
			if p.Color == ThreatColor && p.Type == Pawn {
				return true, nil
			}
		}
	}

	kingOffsets := [8][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	for i := 0; i < len(kingOffsets); i++ {
		offset := kingOffsets[i]
		file := target.File + offset[0]
		rank := target.Rank + offset[1]
		if file >= 0 && file <= 7 && rank >= 0 && rank <= 7 {
			p := board.ChessBoard[rank][file]
			if p.Color == ThreatColor && p.Type == King {
				return true, nil
			}
		}
	}

	return false, nil
}

// IsPathClear reports whether all intermediate squares between from and to are empty
// for the given sliding piece type.
func (board *BoardState) IsPathClear(piece Piece, from Square, to Square) (bool, error) {
	fileDiff := to.File - from.File
	rankDiff := to.Rank - from.Rank
	absFileDiff := absoluteValue(fileDiff)
	absRankDiff := absoluteValue(rankDiff)

	switch piece.Type {
	case Bishop:
		if absFileDiff != absRankDiff {
			return false, fmt.Errorf("Invalid move: bishop must move diagonally.")
		}
	case Rook:
		if absFileDiff != 0 && absRankDiff != 0 {
			return false, fmt.Errorf("Invalid move: rook must move along a rank or file.")
		}
	case Queen:
		if absFileDiff != absRankDiff && absFileDiff != 0 && absRankDiff != 0 {
			return false, fmt.Errorf("Invalid move: queen must move along a rank, file, or diagonal.")
		}
	default:
		return false, fmt.Errorf("Invalid piece type for path checking.")
	}

	fileStep := sign(fileDiff)
	rankStep := sign(rankDiff)

	for file, rank := from.File+fileStep, from.Rank+rankStep; file != to.File || rank != to.Rank; file, rank = file+fileStep, rank+rankStep {
		if board.PieceAt(Square{File: file, Rank: rank}).Type != None {
			return false, nil
		}
	}

	return true, nil
}

// IsPathAttacked reports whether any square in the segment from->to (inclusive) is attacked
// by a piece of ThreatColor.
func (board *BoardState) IsPathAttacked(from Square, to Square, ThreatColor PieceColor) (bool, error) {
	a, err := board.IsSquareAttacked(from, ThreatColor)
	if err != nil {
		return false, err
	}
	if a {
		return true, nil
	}

	b, err := board.IsSquareAttacked(to, ThreatColor)
	if err != nil {
		return false, err
	}
	if b {
		return true, nil
	}

	fileStep := sign(to.File - from.File)
	rankStep := sign(to.Rank - from.Rank)
	for file, rank := from.File+fileStep, from.Rank+rankStep; file != to.File || rank != to.Rank; file, rank = file+fileStep, rank+rankStep {
		attacked, err := board.IsSquareAttacked(Square{File: file, Rank: rank}, ThreatColor)
		if err != nil {
			return false, err
		}
		if attacked {
			return true, nil
		}
	}

	return false, nil
}

func sign(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

func absoluteValue(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
