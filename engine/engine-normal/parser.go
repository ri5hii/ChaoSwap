package engine

import (
	"fmt"
	"strings"
)

// ParseInput normalizes user input for commands and moves.
//
// It trims surrounding whitespace, removes internal spaces and hyphens,
// and lowercases the result.
func ParseInput(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, " ", "")
	input = strings.ReplaceAll(input, "-", "")
	input = strings.ToLower(input)
	if input == "" {
		return ""
	}
	return input
}

// IsMoveNotationValid validates coordinate move notation and extracts its parts.
//
// Supported forms:
//   - "e2e4" (4 characters)
//   - "e7e8q" (5 characters, promotion suffix: q/r/b/n)
//
// It returns:
//   - ok: whether the notation is syntactically valid and squares are in-bounds
//   - from: source square
//   - to: destination square
//   - promotionType: the requested promotion piece type, or None if absent
//   - err: a descriptive error for invalid input
func IsMoveNotationValid(notation string) (ok bool, from Square, to Square, promotionType PieceType, err error) {
	if len(notation) != 4 && len(notation) != 5 {
		return false, Square{}, Square{}, None, fmt.Errorf(
			"invalid move notation: %q. Use 4 chars (e.g. e2e4) or 5 with promotion suffix (e.g. e7e8q).",
			notation,
		)
	}

	fromSquare, fromOk, err1 := ParseNotationToSquare(notation[:2])
	toSquare, toOk, err2 := ParseNotationToSquare(notation[2:4])

	if err1 != nil {
		return false, fromSquare, toSquare, None, err1
	}
	if err2 != nil {
		return false, fromSquare, toSquare, None, err2
	}

	promo := None
	if len(notation) == 5 {
		switch notation[4] {
		case 'q':
			promo = Queen
		case 'r':
			promo = Rook
		case 'b':
			promo = Bishop
		case 'n':
			promo = Knight
		default:
			return false, fromSquare, toSquare, None, fmt.Errorf(
				"invalid promotion suffix: %q. Use one of: q, r, b, n.",
				string(notation[4]),
			)
		}
	}

	validSquares := fromOk && toOk && fromSquare.IsValid() && toSquare.IsValid()
	return validSquares, fromSquare, toSquare, promo, nil
}
