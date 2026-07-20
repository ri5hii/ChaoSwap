# ChaoSwap

A terminal chess variant where players can spend their turn on a **random friendly piece swap** instead of a normal move.

> See [`game.md`](./game.md) for the full game rules, mechanics, TUI controls, move notation, and example scenarios.

---

## Build & Run

```bash
go build ./tui/cmd/chaoSwap/
./chaoSwap start
```

Subcommands: `start`, `description`, `help`.

---

## Requirements

| Dependency | Version |
|------------|---------|
| Go         | 1.25+   |
| Bubble Tea | v2      |
| Lipgloss   | v2      |

---

## Code Layout

```
├── engine/
│   ├── engine-normal/       Standard chess rules, board state, move validation, FEN
│   └── engine-chaos/        Swap pair generation, swap legality (spark rule, king safety, pawn zone)
└── tui/
    ├── app/                 Bubble Tea model, update loop, board/move-log rendering
    └── cmd/chaoSwap/        CLI entrypoint, subcommand routing
```

- **engine-normal**: Board representation, move generation, check/checkmate/stalemate detection, en passant, castling, promotion.
- **engine-chaos**: `PickSwapPair` (uniform random selection), `TrySwap` / `UndoSwap`, legality checks (pawn zone restriction, mobility clause, no-suicide rule, castling-right tracking).
- **tui/app**: `Model`, `Update`, `View`; board and move-log renderers; input handling; mode switching (Normal / Chaos).
- **tui/cmd/chaoSwap**:  `main()` dispatches `start` to the TUI, `help`/`description` to informational output.
