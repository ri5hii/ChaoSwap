package engine

import "fmt"

type Undo struct {
	From          Square
	To            Square
	Captured      Piece
	PrevCastle    uint8
	PrevEnPassant Square
}

func (board *BoardState) MakeMove(piece Piece, from Square, to Square) *Undo {
	if !from.IsValid() || !to.IsValid() {
		return nil
	}
	if board.PieceAt(from).Type == None {
		return nil
	}
	if piece.Color != board.SideToMove {
		return nil
	}

	target := board.PieceAt(to)
	if target.Type != None && target.Color == piece.Color {
		return nil
	}

	undo := Undo{
		From:          from,
		To:            to,
		Captured:      target,
		PrevCastle:    board.CastlingRights,
		PrevEnPassant: board.EnPassantSquare,
	}

	prevEnPassant := board.EnPassantSquare

	var valid bool
	switch piece.Type {
	case Pawn:
		valid = board.handlePawnMove(piece, from, to)
	case Knight:
		valid = board.handleKnightMove(piece, from, to)
	case Bishop:
		valid = board.handleSlidingMove(piece, from, to)
	case Rook:
		valid = board.handleRookMove(piece, from, to)
	case Queen:
		valid = board.handleSlidingMove(piece, from, to)
	case King:
		valid = board.handleKingMove(piece, from, to)
	default:
		return nil
	}

	if !valid {
		return nil
	}

	if board.EnPassantSquare.Equals(prevEnPassant) {
		board.EnPassantSquare = NoSquare
	}

	kingSquare := board.FindKing(piece.Color)
	if board.IsSquareAttacked(kingSquare, PieceColor(1-piece.Color)) {
		UndoMove(board, undo)
		return nil
	}

	board.SideToMove = 1 - board.SideToMove
	return &undo
}

func (board *BoardState) handlePawnMove(piece Piece, from Square, to Square) bool {
	fileDiff := to.File - from.File
	rankDiff := to.Rank - from.Rank
	absoluteFileDiff := absoluteValue(fileDiff)

	target := board.PieceAt(to)

	// en passant
	if absoluteFileDiff == 1 && to.Equals(board.EnPassantSquare) {
		capRank := to.Rank + 1
		if piece.Color == Black {
			capRank = to.Rank - 1
		}
		board.SetPiece(Square{File: to.File, Rank: capRank}, Piece{Type: None})
		board.SetPiece(to, piece)
		board.SetPiece(from, Piece{Type: None})
		return true
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
			return false
		}

		if to.Rank == 0 {
			board.PawnPromotion(piece, to, Queen)
		}
		return true
	}

	// black pawn
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
		return false
	}

	if to.Rank == 7 {
		board.PawnPromotion(piece, to, Queen)
	}
	return true
}

func (board *BoardState) handleKnightMove(piece Piece, from Square, to Square) bool {
	target := board.PieceAt(to)

	absoluteFileDiff := absoluteValue(to.File - from.File)
	absoluteRankDiff := absoluteValue(to.Rank - from.Rank)

	if (absoluteFileDiff == 2 && absoluteRankDiff == 1) || (absoluteFileDiff == 1 && absoluteRankDiff == 2) {
		if (to.Rank == 0 || to.Rank == 7) && target.Type == Rook {
			board.UpdateRookCastlingRights(from, to, piece)
		}
		board.SetPiece(to, piece)
		board.SetPiece(from, Piece{Type: None})
		return true
	}
	return false
}

func (board *BoardState) handleSlidingMove(piece Piece, from Square, to Square) bool {
	target := board.PieceAt(to)

	if !board.IsPathClear(piece, from, to) {
		return false
	}

	if (to.Rank == 0 || to.Rank == 7) && target.Type == Rook {
		board.UpdateRookCastlingRights(from, to, piece)
	}

	board.SetPiece(to, piece)
	board.SetPiece(from, Piece{Type: None})
	return true
}

func (board *BoardState) handleRookMove(piece Piece, from Square, to Square) bool {
	if !board.IsPathClear(piece, from, to) {
		return false
	}

	board.UpdateRookCastlingRights(from, to, piece)

	board.SetPiece(to, piece)
	board.SetPiece(from, Piece{Type: None})
	return true
}

func (board *BoardState) handleKingMove(piece Piece, from Square, to Square) bool {
	absoluteFileDiff := absoluteValue(to.File - from.File)
	absoluteRankDiff := absoluteValue(to.Rank - from.Rank)

	target := board.PieceAt(to)

	// normal move
	if absoluteFileDiff <= 1 && absoluteRankDiff <= 1 {
		if (to.Rank == 0 || to.Rank == 7) && target.Type == Rook {
			board.UpdateRookCastlingRights(from, to, piece)
		}
		board.clearCastlingRights(piece.Color)
		board.SetPiece(to, piece)
		board.SetPiece(from, Piece{Type: None})
		return true
	}

	// castling
	if absoluteRankDiff == 0 && absoluteFileDiff == 2 {
		return board.tryCastle(piece, from, to)
	}

	return false
}

func (board *BoardState) tryCastle(king Piece, from Square, to Square) bool {
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
		return false
	}

	if board.CastlingRights&rights == 0 {
		return false
	}

	if board.IsPathAttacked(from, to, opponent) {
		return false
	}

	if board.PieceAt(rookFrom).Type != Rook {
		return false
	}

	if !board.IsPathClear(Piece{Type: Rook}, rookFrom, from) {
		return false
	}

	board.SetPiece(to, king)
	board.SetPiece(from, Piece{Type: None})
	board.SetPiece(rookTo, Piece{Type: Rook, Color: king.Color})
	board.SetPiece(rookFrom, Piece{Type: None})

	board.CastlingRights &^= rights
	return true
}

func (board *BoardState) clearCastlingRights(color PieceColor) {
	if color == White {
		board.CastlingRights &^= WhiteKingSide | WhiteQueenSide
	} else {
		board.CastlingRights &^= BlackKingSide | BlackQueenSide
	}
}

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

func UndoMove(board *BoardState, undo Undo) {
	board.SetPiece(undo.From, board.PieceAt(undo.To))
	board.SetPiece(undo.To, undo.Captured)
	board.CastlingRights = undo.PrevCastle
	board.EnPassantSquare = undo.PrevEnPassant

	piece := board.PieceAt(undo.From)

	// undo enpassant capture
	if piece.Type == Pawn && undo.From.File != undo.To.File && undo.Captured.Type == None {
		captureRank := undo.To.Rank
		if piece.Color == White {
			captureRank += 1
		} else {
			captureRank -= 1
		}
		board.SetPiece(Square{File: undo.To.File, Rank: captureRank}, Piece{Type: Pawn, Color: PieceColor(1 - piece.Color)})
	}

	//undo rook piece movement in case of castling
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

func (board *BoardState) PawnPromotion(pawn Piece, to Square, promotionType PieceType) bool {
	if pawn.Type != Pawn {
		fmt.Println("Only a pawn can be promoted!")
		return false
	}

	if promotionType != Queen && promotionType != Rook && promotionType != Bishop && promotionType != Knight {
		fmt.Println("Invalid promotion type.")
		return false
	}

	board.SetPiece(to, Piece{Type: promotionType, Color: pawn.Color})
	return true
}

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

func (board *BoardState) IsSquareAttacked(target Square, ThreatColor PieceColor) bool {
	// Rook and Queen attacks along ranks and files
	straightDirections := [4][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
	for i := 0; i < len(straightDirections); i++ {
		direction := straightDirections[i]
		fileStep, rankStep := direction[0], direction[1]
		for file, rank := target.File+fileStep, target.Rank+rankStep; file >= 0 && file <= 7 && rank >= 0 && rank <= 7; file, rank = file+fileStep, rank+rankStep {
			piece := board.ChessBoard[rank][file]
			if piece.Type == None {
				continue
			}
			if piece.Color == ThreatColor && (piece.Type == Rook || piece.Type == Queen) {
				return true
			}
			break
		}
	}
	// Bishop and Queen attacks along diagonals
	diagonalDirections := [4][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	for i := 0; i < len(diagonalDirections); i++ {
		direction := diagonalDirections[i]
		fileStep, rankStep := direction[0], direction[1]
		for file, rank := target.File+fileStep, target.Rank+rankStep; file >= 0 && file <= 7 && rank >= 0 && rank <= 7; file, rank = file+fileStep, rank+rankStep {
			piece := board.ChessBoard[rank][file]
			if piece.Type == None {
				continue
			}
			if piece.Color == ThreatColor && (piece.Type == Bishop || piece.Type == Queen) {
				return true
			}
			break
		}
	}
	// Knight attacks
	knightOffsets := [8][2]int{{1, 2}, {2, 1}, {-1, 2}, {-2, 1}, {1, -2}, {2, -1}, {-1, -2}, {-2, -1}}
	for i := 0; i < len(knightOffsets); i++ {
		offset := knightOffsets[i]
		file := target.File + offset[0]
		rank := target.Rank + offset[1]
		if file >= 0 && file <= 7 && rank >= 0 && rank <= 7 {
			piece := board.ChessBoard[rank][file]
			if piece.Color == ThreatColor && piece.Type == Knight {
				return true
			}
		}
	}
	// Pawn attacks — white pawns attack upward (decreasing rank), black downward
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
			piece := board.ChessBoard[rank][file]
			if piece.Color == ThreatColor && piece.Type == Pawn {
				return true
			}
		}
	}
	// King attacks
	kingOffsets := [8][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	for i := 0; i < len(kingOffsets); i++ {
		offset := kingOffsets[i]
		file := target.File + offset[0]
		rank := target.Rank + offset[1]
		if file >= 0 && file <= 7 && rank >= 0 && rank <= 7 {
			piece := board.ChessBoard[rank][file]
			if piece.Color == ThreatColor && piece.Type == King {
				return true
			}
		}
	}
	return false
}

func (board *BoardState) IsPathClear(piece Piece, from Square, to Square) bool {
	fileDiff := to.File - from.File
	rankDiff := to.Rank - from.Rank
	absFileDiff := absoluteValue(fileDiff)
	absRankDiff := absoluteValue(rankDiff)

	switch piece.Type {
	case Bishop:
		if absFileDiff != absRankDiff {
			return false
		}
	case Rook:
		if absFileDiff != 0 && absRankDiff != 0 {
			return false
		}
	case Queen:
		if absFileDiff != absRankDiff && absFileDiff != 0 && absRankDiff != 0 {
			return false
		}
	default:
		return false
	}

	fileStep := sign(fileDiff)
	rankStep := sign(rankDiff)

	for file, rank := from.File+fileStep, from.Rank+rankStep; file != to.File || rank != to.Rank; file, rank = file+fileStep, rank+rankStep {
		if board.PieceAt(Square{File: file, Rank: rank}).Type != None {
			return false
		}
	}

	return true
}

func (board *BoardState) IsPathAttacked(from Square, to Square, ThreatColor PieceColor) bool {
	if board.IsSquareAttacked(from, ThreatColor) || board.IsSquareAttacked(to, ThreatColor) {
		return true
	}
	fileStep := sign(to.File - from.File)
	rankStep := sign(to.Rank - from.Rank)
	for file, rank := from.File+fileStep, from.Rank+rankStep; file != to.File || rank != to.Rank; file, rank = file+fileStep, rank+rankStep {
		if board.IsSquareAttacked(Square{File: file, Rank: rank}, ThreatColor) {
			return true
		}
	}
	return false
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
