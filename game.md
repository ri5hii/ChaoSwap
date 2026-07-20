# ChaoSwap — Game Rules & User Guide

ChaoSwap is a two-player, turn-based chess variant played in the terminal. It layers a **Chaos Swap** mechanic on top of standard chess, giving each player the option to spend their turn performing a random friendly piece swap instead of making a normal move.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Normal Chess Rules](#2-normal-chess-rules)
3. [Chaos Swap Mechanic](#3-chaos-swap-mechanic)
4. [Swap Legality Rules](#4-swap-legality-rules)
5. [State Effects of a Swap](#5-state-effects-of-a-swap)
6. [TUI Controls & User Flow](#6-tui-controls--user-flow)
7. [Move Notation](#7-move-notation)
8. [Game End Conditions](#8-game-end-conditions)
9. [Example Scenarios](#9-example-scenarios)

---

## 1) Overview

Each turn, the active player chooses one of two actions:

| Action       | Description                                                  |
|--------------|--------------------------------------------------------------|
| Normal move  | Make any standard-legal chess move.                          |
| Chaos swap   | Let the engine pick a random legal swap of two friendly pieces. |

Both actions consume the player's turn. The game ends under the same conditions as standard chess (checkmate, stalemate, or draw), except that swap moves count as additional legal actions when determining mate and stalemate.

---

## 2) Normal Chess Rules

ChaoSwap uses standard FIDE chess rules for normal moves, including:

- **Piece movement**: all pieces move according to standard rules (pawns, knights, bishops, rooks, queens, and kings).
- **Captures**: a piece may capture an enemy piece by moving to its square.
- **Pawn rules**:
  - Single push forward.
  - Double push from the starting rank (rank 2 for White, rank 7 for Black) if both squares ahead are empty.
  - Diagonal capture of an adjacent enemy pawn.
  - **En passant**: a pawn that just double-pushed may be captured en passant by an opposing pawn on the immediately following ply only. The right expires after one ply.
  - **Promotion**: a pawn reaching rank 8 (White) or rank 1 (Black) must promote to a queen, rook, bishop, or knight.
- **Castling**: the king may castle kingside (O-O) or queenside (O-O-O) if:
  - Neither the king nor the relevant rook has previously moved.
  - No pieces stand between the king and the rook.
  - The king is not currently in check.
  - The king does not pass through or land on an attacked square.
- **Check**: a player may not make any move that leaves their own king in check.
- **Checkmate**: a player whose king is in check and has no legal move (including no legal swap) loses the game.
- **Stalemate**: a player with no legal move (including no legal swap) and whose king is not in check draws the game.

---

## 3) Chaos Swap Mechanic

### 3.1 What is a Chaos swap?

A Chaos swap exchanges the board positions of **two friendly pieces** in a single action. No capture occurs. The swap is resolved instantly; there is no intermediate board state.

### 3.2 Who chooses the swap?

The **player does not select** which pieces to swap. Instead:

1. The player toggles into **Chaos mode** by pressing `TAB`.
2. The player commits the swap by pressing `S`.
3. The engine selects a legal swap **uniformly at random** from the complete set of legal swaps available to the active side.

This means every legal swap has exactly the same probability of being chosen. There is no steering or filtering by the player beyond the decision to enter Chaos mode.

### 3.3 Turn consumption

- Committing a swap (`S` in Chaos mode) consumes the active player's turn, identical to making a normal move.
- If no legal swap exists when `S` is pressed, **the turn is not consumed** and the board is not changed. A message is displayed instead.
- Toggling Chaos mode on or off (`TAB`) does not consume a turn.

### 3.4 Randomness & reproducibility

The engine's random swap selection is driven by an explicit RNG instance. The seed used for the current session can be logged or displayed for debugging and reproducibility purposes.

---

## 4) Swap Legality Rules

A swap `S(A, B)` — exchanging the pieces on squares A and B — is **legal** if and only if **all** of the following conditions are satisfied:

### 4.1 Both squares are occupied by friendly pieces

- Square A and square B must both be occupied.
- Both pieces must belong to the side to move.
- `A` and `B` must be different squares.

### 4.2 Kings cannot be swapped (King Anchor Rule)

- Neither piece may be a king.
- The king remains fixed during a swap action.

### 4.3 Pawns cannot be swapped to the back or promotion rank (Pawn Zone Restriction)

- If the piece on A is a pawn, it cannot be swapped to rank 1 or rank 8 (its **destination** rank, i.e. rank(B), must not be 1 or 8).
- If the piece on B is a pawn, it cannot be swapped to rank 1 or rank 8 (i.e. rank(A) must not be 1 or 8).

This prevents a pawn from being teleported to a promotion or illegal back-rank square.

### 4.4 Mobility Clause ("Spark Rule")

At least one of the two pieces must have at least one legal **standard** move available **before** the swap is applied.

- "Legal standard move" means a move that does not leave the mover's king in check (pins are respected; pseudo-legal moves do not count).
- Mobility is evaluated in the position **before** the swap.
- If both pieces are completely immobilised (e.g. both pinned with no legal standard moves), the swap is **not legal**.

This prevents a player from stalling indefinitely by swapping two pieces that can never contribute to play.

### 4.5 King safety after the swap ("No Suicide")

After the swap is applied, the active side's king must **not** be in check.

- Only the final post-swap position matters; there is no intermediate check state.
- A swap that would leave the king in check (e.g. by uncovering a discovered check on the friendly king) is illegal.

---

## 5) State Effects of a Swap

When a legal swap is executed, the following state changes occur:

| State field          | Effect                                                                    |
|----------------------|---------------------------------------------------------------------------|
| Board                | The two pieces exchange squares. No captures occur.                       |
| Side to move         | Toggles to the other side.                                                |
| En passant target    | **Cleared unconditionally.** No en passant is possible after a swap.      |
| Halfmove clock       | **Reset to 0.** Swaps are treated as structure-changing moves.            |
| Castling rights      | Permanently revoked if a rook is swapped away from its home corner.       |

### 5.1 Castling rights detail

Castling rights are revoked when a rook is swapped away from its **original corner square**, as follows:

| Corner square | Right revoked              |
|---------------|----------------------------|
| a1            | White queenside castling   |
| h1            | White kingside castling    |
| a8            | Black queenside castling   |
| h8            | Black kingside castling    |

Once revoked, castling rights are **permanently lost** — even if the rook later returns to its original corner via another swap or a normal move.

### 5.2 En passant and swaps

Because every swap unconditionally clears the en passant target square:

- If a pawn has just double-pushed (creating an en passant target) and then the **opponent** performs a swap instead of capturing en passant, the en passant right expires immediately.
- If the side to move performs a swap **instead** of capturing en passant, the en passant right also expires.

---

## 6) TUI Controls & User Flow

### 6.1 Modes

The TUI has two input modes:

| Mode   | Description                                          |
|--------|------------------------------------------------------|
| Normal | Standard chess move input via coordinate notation.  |
| Chaos  | Random swap mode; commit swap with `S`.             |

You can always see the current mode in the status bar at the bottom of the screen.

### 6.2 Keyboard controls

| Key            | Action                                              |
|----------------|-----------------------------------------------------|
| Type characters| Build up the move/command input buffer.             |
| `ENTER`        | Submit the current input buffer as a move or command. |
| `BACKSPACE`    | Delete the last character from the input buffer.    |
| `TAB`          | Toggle between Normal and Chaos mode.               |
| `S`            | (Chaos mode only) Commit a random legal swap.       |
| `Ctrl+C`       | Quit the application immediately.                   |

### 6.3 Making a normal move

1. Ensure you are in **Normal** mode (press `TAB` to switch if needed).
2. Type the move in coordinate notation, e.g. `e2e4`.
3. Press `ENTER` to submit.
4. If the move is illegal, a descriptive error message is shown and you may try again.

**Promotion**: When a pawn reaches the last rank, you must specify a promotion piece:

- Include the piece suffix directly: `e7e8q`, `e7e8r`, `e7e8b`, `e7e8n`.
- Or enter just `e7e8` and press `ENTER` — the TUI will prompt you to choose: `q` (queen), `r` (rook), `b` (bishop), `n` (knight).
- Type `cancel` at the promotion prompt to cancel and re-enter the move.

### 6.4 Performing a Chaos swap

1. Press `TAB` to enter **Chaos** mode.
2. Press `S` to commit the random swap for the current side.
   - If a legal swap exists, the engine picks one at random, applies it, and displays the result (e.g. `Last swap: S(e4,h8)`).
   - If no legal swap exists, the turn is **not** consumed and the message `No legal swap available` is displayed.
3. Press `TAB` again to return to Normal mode without committing a swap.

### 6.5 Commands

All commands are prefixed with `:`. Type the command and press `ENTER`.

| Command             | Description                        |
|---------------------|------------------------------------|
| `:help` or `:h`     | Show the help/command reference.   |
| `:description` or `:d` | Show a short game description.  |
| `:normal`           | Switch to Normal mode.             |
| `:chaos`            | Switch to Chaos mode.              |
| `:quit` or `:q`     | Quit the application.              |

Commands are case-insensitive and ignore extra whitespace.

### 6.6 Summary screen

When the game ends (checkmate, stalemate, or draw), the TUI switches to a **summary screen**:

- The final board position and full move list are displayed.
- No further moves are accepted.
- Use `:quit` to exit.

---

## 7) Move Notation

### 7.1 Normal moves

Coordinate (long algebraic) notation is used for normal moves:

```
<from><to>
```

Examples:

| Input    | Meaning                              |
|----------|--------------------------------------|
| `e2e4`   | Move piece from e2 to e4.            |
| `g1f3`   | Move piece from g1 to f3.            |
| `e1g1`   | Castle kingside (White).             |
| `e1c1`   | Castle queenside (White).            |
| `e7e8q`  | Move pawn e7 to e8, promote to queen.|

### 7.2 Swap notation

Swaps are recorded and displayed using the following format:

```
S(<from>,<to>)
```

Example: `S(e4,h8)` means the pieces on e4 and h8 were exchanged.

This notation is used in the move log and status messages. It is not valid input for Normal mode move entry.

---

## 8) Game End Conditions

| Condition              | Result                       | Notes                                                                 |
|------------------------|------------------------------|-----------------------------------------------------------------------|
| Checkmate              | The other side wins.         | No legal move **or** legal swap exists while in check.               |
| Stalemate              | Draw.                        | No legal move **or** legal swap exists while not in check.           |
| Insufficient material  | Draw.                        | Standard cases: K vs K, K+N vs K, K+B vs K.                         |
| 50-move rule           | Draw.                        | 50 full moves without a pawn move, capture, **or swap**.             |

> **Important**: swap moves are counted as legal actions for the purposes of checkmate and stalemate detection. If a player's only escape from checkmate is a legal swap, they are **not** in checkmate.

> **Note on the 50-move rule**: every swap resets the halfmove clock to 0, so swaps prevent 50-move draw accumulation just as pawn moves and captures do.

---

## 9) Example Scenarios

### 9.1 Basic swap

White has a bishop on c1 and a rook on a1. White enters Chaos mode and presses `S`. The engine randomly selects the swap `S(c1,a1)`. The bishop moves to a1 and the rook moves to c1 — but White's **queenside castling right is now permanently revoked** because the rook left its home corner a1.

### 9.2 Mobility clause blocks a swap

White has a pawn on d4 that is pinned to the king (zero legal standard moves) and a knight on b1 that is also completely blocked (zero legal standard moves). A swap `S(d4,b1)` would be **illegal** because neither piece has any legal standard move. The engine will not generate this pair as a legal swap.

### 9.3 Swap escapes checkmate

Black's king is in check from a White rook on d8, and Black appears to have no normal moves. However, Black has a legal swap `S(f8,b8)` — after the swap, the piece now on f8 blocks the check. Because this legal swap exists, it is **not checkmate**. Black must perform the swap (or any other legal escape).

### 9.4 En passant expires after a swap

White plays `e2e4` (double pawn push), creating an en passant target on e3. Instead of capturing en passant, Black enters Chaos mode and presses `S`, executing a random swap. The en passant target is **immediately cleared** as part of the swap. White's e4 pawn can no longer be captured en passant.

### 9.5 Discovered check via swap

White has a queen on d1 and a knight on d4, blocking the queen's line to the Black king on d8. White executes a swap `S(d4,a4)`, moving the knight off the d-file. This reveals a check on d8 from the queen. The swap is legal and results in a check on Black's king.
