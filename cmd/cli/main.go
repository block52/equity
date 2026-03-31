package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/block52/poker-equity/pkg/equity"
)

func main() {
	opts, positional := parseFlags(os.Args[1:])

	if opts.help {
		printUsage()
		os.Exit(0)
	}

	if opts.sims > 100000 {
		opts.sims = 100000
	}
	if opts.sims <= 0 {
		opts.sims = 10000
	}

	var req request
	var err error

	if opts.jsonFile != "" {
		req, err = loadJSON(opts.jsonFile)
	} else {
		req, err = parseHands(positional, opts.board, opts.dead)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if req.Simulations == 0 {
		req.Simulations = opts.sims
	}

	calc := equity.NewCalculator(
		equity.WithSimulations(req.Simulations),
		equity.WithWorkers(4),
	)

	result, err := calc.CalculateEquity(req.Hands, req.Board, req.Dead)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if opts.outputJSON {
		printJSON(result)
	} else {
		printTable(result, req.Hands)
	}
}

type options struct {
	jsonFile   string
	board      string
	dead       string
	sims       int
	outputJSON bool
	help       bool
}

// parseFlags extracts flags from anywhere in the args and returns remaining positional args.
func parseFlags(args []string) (options, []string) {
	opts := options{sims: 10000}
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h":
			opts.help = true
		case arg == "--output-json":
			opts.outputJSON = true
		case (arg == "--json" || arg == "-json") && i+1 < len(args):
			i++
			opts.jsonFile = args[i]
		case strings.HasPrefix(arg, "--json="):
			opts.jsonFile = strings.TrimPrefix(arg, "--json=")
		case (arg == "--board" || arg == "-board") && i+1 < len(args):
			i++
			opts.board = args[i]
		case strings.HasPrefix(arg, "--board="):
			opts.board = strings.TrimPrefix(arg, "--board=")
		case (arg == "--dead" || arg == "-dead") && i+1 < len(args):
			i++
			opts.dead = args[i]
		case strings.HasPrefix(arg, "--dead="):
			opts.dead = strings.TrimPrefix(arg, "--dead=")
		case (arg == "--sims" || arg == "-sims") && i+1 < len(args):
			i++
			if n, err := strconv.Atoi(args[i]); err == nil {
				opts.sims = n
			}
		case strings.HasPrefix(arg, "--sims="):
			if n, err := strconv.Atoi(strings.TrimPrefix(arg, "--sims=")); err == nil {
				opts.sims = n
			}
		default:
			positional = append(positional, arg)
		}
	}

	return opts, positional
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: poker-equity [flags] <hand1>, <hand2> [, <hand3> ...]

Calculate Texas Hold'em poker equity.

Examples:
  poker-equity AS AH, KS KH
  poker-equity AS KS, QH QD --board KH,7C,2D
  poker-equity AS AH, KS KH, QS QH, JC TC --sims 50000
  poker-equity --json data.json
  poker-equity AS AH, KS KH --output-json

Flags:
  --json <file>     path to JSON input file
  --board <cards>   community cards (comma-separated, e.g. KH,7C,2D)
  --dead <cards>    dead/mucked cards (comma-separated)
  --sims <n>        number of simulations (default 10000, max 100000)
  --output-json     output results as JSON
  --help            show this help
`)
}

type request struct {
	Hands       [][]string `json:"hands"`
	Board       []string   `json:"board"`
	Dead        []string   `json:"dead"`
	Simulations int        `json:"simulations"`
}

func loadJSON(path string) (request, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return request{}, fmt.Errorf("reading %s: %w", path, err)
	}
	var req request
	if err := json.Unmarshal(data, &req); err != nil {
		return request{}, fmt.Errorf("parsing JSON: %w", err)
	}
	if len(req.Hands) < 2 {
		return request{}, fmt.Errorf("must provide at least 2 hands")
	}
	return req, nil
}

// parseHands parses positional args like: AS AH, KS KH
func parseHands(args []string, boardStr, deadStr string) (request, error) {
	if len(args) == 0 {
		return request{}, fmt.Errorf("no hands provided. Use --help for usage")
	}

	raw := strings.Join(args, " ")
	handStrs := strings.Split(raw, ",")

	var hands [][]string
	for _, h := range handStrs {
		cards := strings.Fields(strings.TrimSpace(h))
		if len(cards) == 0 {
			continue
		}
		for i := range cards {
			cards[i] = strings.ToUpper(cards[i])
		}
		if len(cards) != 2 {
			return request{}, fmt.Errorf("each hand must have exactly 2 cards, got %d: %v", len(cards), cards)
		}
		hands = append(hands, cards)
	}

	if len(hands) < 2 {
		return request{}, fmt.Errorf("must provide at least 2 hands separated by commas")
	}

	var board []string
	if boardStr != "" {
		for _, c := range strings.Split(boardStr, ",") {
			c = strings.TrimSpace(strings.ToUpper(c))
			if c != "" {
				board = append(board, c)
			}
		}
	}

	var dead []string
	if deadStr != "" {
		for _, c := range strings.Split(deadStr, ",") {
			c = strings.TrimSpace(strings.ToUpper(c))
			if c != "" {
				dead = append(dead, c)
			}
		}
	}

	return request{Hands: hands, Board: board, Dead: dead}, nil
}

func printTable(result *equity.CalculationResult, hands [][]string) {
	fmt.Printf("\n  Stage: %s | Simulations: %d | Duration: %.2fms\n\n",
		result.Stage, result.Simulations, float64(result.Duration.Microseconds())/1000.0)
	fmt.Printf("  %-10s  %-10s  %8s  %8s  %8s\n", "Hand", "Cards", "Equity", "Win", "Tie")
	fmt.Printf("  %-10s  %-10s  %8s  %8s  %8s\n", "----------", "----------", "--------", "--------", "--------")

	for i, r := range result.Results {
		handStr := strings.Join(hands[i], " ")
		fmt.Printf("  Player %-3d  %-10s  %7.2f%%  %7.2f%%  %7.2f%%\n",
			i+1,
			handStr,
			r.Total*100,
			r.Equity*100,
			r.TieEquity*100,
		)
	}
	fmt.Println()
}

func printJSON(result *equity.CalculationResult) {
	out := struct {
		Results     []equity.EquityResult `json:"results"`
		Simulations int                   `json:"simulations"`
		Stage       string                `json:"stage"`
		DurationMs  float64               `json:"duration_ms"`
		HandsPerSec float64               `json:"hands_per_sec"`
		BoardCards  []string              `json:"board_cards"`
	}{
		Results:     result.Results,
		Simulations: result.Simulations,
		Stage:       result.Stage.String(),
		DurationMs:  float64(result.Duration.Microseconds()) / 1000.0,
		HandsPerSec: result.HandsPerSec,
		BoardCards:  result.BoardCards,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}
