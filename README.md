# Poker Equity Calculator

A high-performance Texas Hold'em poker equity calculator exposed as a REST API microservice. Uses Monte Carlo simulation with a fast bitwise hand evaluator to calculate win probabilities for 2-9 players at any stage of a hand.

## Features

- **Monte Carlo simulation** with configurable iteration count (up to 100k)
- **Fast hand evaluator** using bit manipulation and lookup tables (~100x faster than combinatorial)
- **All stages supported**: preflop, flop, turn, and river
- **2-9 player** equity calculations
- **Parallel workers** for high-throughput simulation
- **Dead card support** for mucked/known cards
- **Deterministic river evaluation** (no simulation needed)
- **CORS enabled** for browser-based clients
- **Docker ready** for Digital Ocean App Platform deployment

## Card Notation

Cards use a 2-character mnemonic format: `[Rank][Suit]`

**Ranks**: `A`, `2`-`9`, `T`, `J`, `Q`, `K`
**Suits**: `C` (clubs), `D` (diamonds), `H` (hearts), `S` (spades)

Examples: `AS` (Ace of Spades), `TH` (Ten of Hearts), `2C` (Two of Clubs)

## Quick Start

### CLI

```bash
# Preflop: AA vs KK
go run ./cmd/cli AS AH, KS KH

# With board cards
go run ./cmd/cli AS KS, QH QD --board KH,7C,2D

# 4-way with custom simulations
go run ./cmd/cli AS AH, KS KH, QS QH, JC TC --sims 50000

# From a JSON file
go run ./cmd/cli --json example.json

# JSON output
go run ./cmd/cli AS AH, KS KH --output-json
```

Output:

```
  Stage: Preflop | Simulations: 10000 | Duration: 1.67ms

  Hand        Cards         Equity       Win       Tie
  ----------  ----------  --------  --------  --------
  Player 1    AS AH         82.15%    81.14%     1.01%
  Player 2    KS KH         17.85%    16.84%     1.01%
```

**CLI Flags:**

| Flag            | Description                                      |
|-----------------|--------------------------------------------------|
| `--board`       | Community cards, comma-separated (e.g. `KH,7C,2D`) |
| `--dead`        | Dead/mucked cards, comma-separated               |
| `--sims <n>`    | Number of simulations (default 10000, max 100000) |
| `--json <file>` | Read input from a JSON file                      |
| `--output-json` | Output results as JSON                           |
| `--help`        | Show usage                                       |

**JSON file format** (`example.json`):

```json
{
  "hands": [["AS", "AH"], ["KS", "KH"]],
  "board": ["2D", "7C", "JH"],
  "simulations": 20000
}
```

### HTTP Server

```bash
go run ./cmd/server
```

The server starts on port `8080` by default. Set the `PORT` environment variable to change it.

### Docker

```bash
docker build -t poker-equity .
docker run -p 8080:8080 poker-equity
```

## API Reference

### `GET /health`

Health check endpoint.

```bash
curl http://localhost:8080/health
```

```json
{
  "status": "ok",
  "service": "poker-equity"
}
```

### `POST /api/v1/equity`

Calculate equity for multiple hands.

**Request body:**

| Field         | Type       | Required | Description                          |
|---------------|------------|----------|--------------------------------------|
| `hands`       | `string[][]` | Yes   | Array of hands, each with 2 cards    |
| `board`       | `string[]`   | No    | Community cards (0, 3, 4, or 5)      |
| `dead`        | `string[]`   | No    | Dead/mucked cards                    |
| `simulations` | `int`        | No    | Number of simulations (default 10000, max 100000) |

**Examples:**

Preflop - AA vs KK:
```bash
curl -X POST http://localhost:8080/api/v1/equity \
  -H "Content-Type: application/json" \
  -d '{
    "hands": [["AS", "AH"], ["KS", "KH"]],
    "simulations": 50000
  }'
```

Flop - AK vs QQ on K-7-2:
```bash
curl -X POST http://localhost:8080/api/v1/equity \
  -H "Content-Type: application/json" \
  -d '{
    "hands": [["AS", "KS"], ["QH", "QD"]],
    "board": ["KH", "7C", "2D"]
  }'
```

River - deterministic evaluation:
```bash
curl -X POST http://localhost:8080/api/v1/equity \
  -H "Content-Type: application/json" \
  -d '{
    "hands": [["AS", "KS"], ["QH", "QD"]],
    "board": ["2S", "7S", "QC", "3H", "9S"]
  }'
```

Multi-way pot (4 players):
```bash
curl -X POST http://localhost:8080/api/v1/equity \
  -H "Content-Type: application/json" \
  -d '{
    "hands": [["AS", "AH"], ["KS", "KH"], ["QS", "QH"], ["JC", "TC"]],
    "simulations": 20000
  }'
```

**Response:**

```json
{
  "results": [
    {
      "hand_index": 0,
      "wins": 8216,
      "ties": 34,
      "losses": 1750,
      "equity": 0.8216,
      "tie_equity": 0.0017,
      "total": 0.8233
    },
    {
      "hand_index": 1,
      "wins": 1750,
      "ties": 34,
      "losses": 8216,
      "equity": 0.175,
      "tie_equity": 0.0017,
      "total": 0.1767
    }
  ],
  "simulations": 10000,
  "stage": "Preflop",
  "duration_ms": 3.456,
  "hands_per_sec": 5787037,
  "board_cards": null
}
```

### `POST /api/v1/evaluate`

Evaluate a single poker hand (5-7 cards).

**Request body:**

| Field    | Type       | Required | Description          |
|----------|------------|----------|----------------------|
| `cards`  | `string[]` | Yes      | 5-7 cards to evaluate |

**Example:**

```bash
curl -X POST http://localhost:8080/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{"cards": ["AS", "KS", "QS", "JS", "TS"]}'
```

**Response:**

```json
{
  "rank": "Straight Flush",
  "category": 8,
  "score": 134479872
}
```

Hand categories (0-8):

| Category | Name              |
|----------|-------------------|
| 0        | High Card         |
| 1        | One Pair          |
| 2        | Two Pair          |
| 3        | Three of a Kind   |
| 4        | Straight          |
| 5        | Flush             |
| 6        | Full House        |
| 7        | Four of a Kind    |
| 8        | Straight Flush    |

## Deploy to Digital Ocean

### App Platform (recommended)

1. Fork this repository
2. In the Digital Ocean console, create a new App
3. Connect your GitHub repo
4. Digital Ocean will auto-detect the `Dockerfile`
5. Deploy

Or use the included app spec:

```bash
doctl apps create --spec .do/app.yaml
```

### Droplet

```bash
# On your droplet
git clone https://github.com/block52/poker-equity.git
cd poker-equity
docker build -t poker-equity .
docker run -d -p 8080:8080 --restart unless-stopped poker-equity
```

## Development

### Run tests

```bash
go test ./...
```

### Run tests with verbose output

```bash
go test ./pkg/equity/ -v
```

### Run benchmarks

```bash
go test ./pkg/equity/ -bench=. -benchmem
```

### Run speed diagnostics

```bash
go test ./pkg/equity/ -v -run TestSpeedDiagnostics
```

## Project Structure

```
.
├── cmd/
│   ├── cli/
│   │   └── main.go          # CLI tool
│   └── server/
│       └── main.go          # HTTP server and API handlers
├── pkg/
│   ├── types/
│   │   └── deck.go          # Card, Suit, Deck types
│   └── equity/
│       ├── equity.go         # Monte Carlo equity calculator
│       ├── evaluator.go      # Standard hand evaluator
│       ├── evaluator_fast.go # Optimized bitwise evaluator
│       └── equity_test.go    # Comprehensive test suite
├── .do/
│   └── app.yaml             # Digital Ocean App Platform spec
├── example.json              # Example JSON input for CLI
├── Dockerfile
├── go.mod
└── README.md
```

## How It Works

1. **Card Representation**: Cards use the Block52 schema - each card has a suit (1-4), rank (1-13), computed value (0-51), and a 2-character mnemonic string.

2. **Hand Evaluation**: The fast evaluator uses 13-bit rank bitmasks per suit, precomputed straight/flush lookup tables, and rank frequency counting to classify hands in constant time.

3. **Monte Carlo Simulation**: For non-river stages, the calculator runs N simulations in parallel across multiple workers. Each simulation shuffles the remaining deck, completes the board, evaluates all hands, and records wins/ties.

4. **River Evaluation**: When all 5 community cards are known, hands are evaluated deterministically with no simulation needed.

## Performance

Typical performance on modern hardware:

| Operation                | Speed            |
|--------------------------|------------------|
| Fast 7-card evaluation   | ~30M evals/sec   |
| 2-way preflop (10k sims) | ~3ms             |
| 9-way preflop (10k sims) | ~15ms            |
| 2-way flop (10k sims)    | ~2ms             |
| River evaluation          | <0.01ms          |

## License

MIT License - see [LICENSE](LICENSE) for details.
